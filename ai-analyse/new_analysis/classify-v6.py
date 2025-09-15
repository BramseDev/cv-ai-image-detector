#!/usr/bin/env python3

"""
DeepFake Detector - Inference Script

Ein einfaches Script zur Klassifizierung von Bildern mit dem trainierten Modell.

Usage:
# Verwendet erstes .pt Model im selben Verzeichnis, analysiert ein oder mehrere Bilder:
python classify.py image1.jpg image2.jpg

# Aktiviere TTA Vorverarbeitung
python classify.py --tta image1.jpg

# Gibt einen zu verwendenden Threshold an (ohne den Parameter wird gespeicherter Model-Threshold genutzt)
python classify.py --threshold 0.5 image1.jpg

# Gib einen Ordner mit zu klassifizierenden Bildern an
python classify.py folder/to/images/

# Gibt ein explizites Model an das verwendet werden soll
python classify.py --model path/to/model.pt image1.jpg

# Nutzt alle .pt Models im angegebenen Models Ordner und bewertet mit Ensemble
python classify.py --models path/to/models/ image1.jpg

# Nutzt einen Test-Folder mit fake/real Unterordnern um Genauigkeit der Model-Vorhersagen zu berechnen
python classify.py --test path/to/test_folder

# Live-Modus: Models werden einmal geladen, dann interaktive Eingabe
python classify.py --models path/to/models/ --live

Autor: AI Detector
Version: 1.7.1
"""

import os
import sys
import json
import argparse
import shlex
from pathlib import Path
from glob import glob
from typing import List, Dict, Optional, Tuple
import warnings

warnings.filterwarnings('ignore')

import numpy as np
import torch
import torch.nn as nn
import timm
import albumentations as A
import cv2
from PIL import Image
from torch.utils.data import Dataset, DataLoader
from sklearn.metrics import accuracy_score, precision_recall_fscore_support, confusion_matrix

# ========================================================================================
# CONSTANTS & SETUP
# ========================================================================================
DEFAULT_IMG_SIZE = 448
DEFAULT_TTA_AUGS = 8
DEFAULT_BATCH_SIZE = 32
DEFAULT_WORKERS = 0
SUPPORTED_EXTENSIONS = ['.jpg', '.jpeg', '.png', '.webp', '.bmp']
MODEL_EXTENSIONS = ['.pt', '.pth']
IMAGENET_MEAN = np.array([0.485, 0.456, 0.406], dtype=np.float32)
IMAGENET_STD = np.array([0.229, 0.224, 0.225], dtype=np.float32)

device = torch.device("cuda" if torch.cuda.is_available() else "cpu")

# Optimize CUDNN for consistent input sizes
if torch.backends.cudnn.is_available():
    torch.backends.cudnn.benchmark = True

# ========================================================================================
# IMAGE PROCESSING & FORENSICS
# ========================================================================================

def imread_rgb(path: Path) -> np.ndarray:
    """Robust image reading function."""
    try:
        data = np.fromfile(str(path), dtype=np.uint8)
        img = cv2.imdecode(data, cv2.IMREAD_COLOR) 
        if img is not None:
            return cv2.cvtColor(img, cv2.COLOR_BGR2RGB)
    except:
        pass

    with Image.open(path) as im:
        im = im.convert("RGB")
        return np.array(im)

def fft_log_mag(gray_f32: np.ndarray) -> np.ndarray:
    """FFT log magnitude computation."""
    f = np.fft.fft2(gray_f32)
    fshift = np.fft.fftshift(f)
    magnitude = np.abs(fshift)
    logmag = np.log1p(magnitude)
    return logmag.astype(np.float32)

def srm_filter_bank():
    """SRM filter bank for forensic analysis."""
    k = []
    k.append(np.array([[-1, 2, -1], [2, -4, 2], [-1, 2, -1]], dtype=np.float32) / 4.0)
    k.append(np.array([[-1, 1, 0], [-1, 1, 0], [-1, 1, 0]], dtype=np.float32))
    k.append(np.array([[-1, -1, -1], [1, 1, 1], [0, 0, 0]], dtype=np.float32))
    k.append(np.array([[0, -1, 1], [-1, 1, 0], [1, 0, 0]], dtype=np.float32))
    k.append(np.array([[1, -1, 0], [0, 1, -1], [0, 0, 0]], dtype=np.float32))
    return k

_SRM_KS = srm_filter_bank()

def apply_srm_bank(gray_f32: np.ndarray) -> np.ndarray:
    """Apply SRM filter bank."""
    out = []
    g = (gray_f32 * 255.0).astype(np.float32)
    for kern in _SRM_KS:
        resp = cv2.filter2D(g, ddepth=cv2.CV_32F, kernel=kern)
        out.append(resp.astype(np.float32))
    return np.stack(out, axis=0)  # 5xHxW

def zscore_per_sample(x: np.ndarray) -> np.ndarray:
    """Per-sample Z-score normalization."""
    mu = x.mean(axis=(1, 2), keepdims=True)
    sd = x.std(axis=(1, 2), keepdims=True) + 1e-6
    return (x - mu) / sd

def build_forensic_channels(rgb_uint8: np.ndarray) -> np.ndarray:
    """Build forensic channels (FFT + SRM)."""
    gray = cv2.cvtColor(rgb_uint8, cv2.COLOR_RGB2GRAY).astype(np.float32) / 255.0
    fftc = fft_log_mag(gray)  # HxW
    srmc = apply_srm_bank(gray)  # 5xHxW
    forensic = np.concatenate([fftc[None, ...], srmc], axis=0)  # 6xHxW
    forensic = zscore_per_sample(forensic)
    return forensic

# ========================================================================================
# DATA PROCESSING & DATASET
# ========================================================================================

def make_val_aug(size: int):
    """Validation/inference augmentation."""
    return A.Compose([
        A.LongestMaxSize(max_size=size),
        A.PadIfNeeded(min_height=size, min_width=size, border_mode=cv2.BORDER_CONSTANT),
    ])

def make_tta_transforms(size: int, num_augments: int = DEFAULT_TTA_AUGS):
    """Test-time augmentation transforms."""
    transforms = []

    # Original (centered)
    transforms.append(A.Compose([
        A.LongestMaxSize(max_size=size),
        A.PadIfNeeded(min_height=size, min_width=size, border_mode=cv2.BORDER_CONSTANT)
    ]))

    # Horizontal flip
    transforms.append(A.Compose([
        A.LongestMaxSize(max_size=size),
        A.PadIfNeeded(min_height=size, min_width=size, border_mode=cv2.BORDER_CONSTANT),
        A.HorizontalFlip(p=1.0)
    ]))

    # Simple random crops from slightly larger image
    for _ in range(min(3, num_augments - len(transforms))):
        transforms.append(A.Compose([
            A.LongestMaxSize(max_size=int(size * 1.1)),
            A.PadIfNeeded(min_height=int(size * 1.1), min_width=int(size * 1.1),
                         border_mode=cv2.BORDER_CONSTANT),
            A.RandomCrop(height=size, width=size, p=1.0)
        ]))

    # Light rotations
    for angle in [5, -5]:
        if len(transforms) < num_augments:
            transforms.append(A.Compose([
                A.LongestMaxSize(max_size=size),
                A.PadIfNeeded(min_height=size, min_width=size, border_mode=cv2.BORDER_CONSTANT),
                A.Rotate(limit=[angle, angle], p=1.0, border_mode=cv2.BORDER_CONSTANT)
            ]))

    return transforms[:num_augments]

class ImageDataset(Dataset):
    """Dataset for image classification with optional TTA."""

    def __init__(self, image_paths: List[str], img_size: int = DEFAULT_IMG_SIZE,
                 tta_transforms: Optional[List] = None):
        self.image_paths = image_paths
        self.img_size = img_size
        self.tta_transforms = tta_transforms
        self.base_transform = make_val_aug(img_size)

    def __len__(self):
        if self.tta_transforms:
            return len(self.image_paths) * len(self.tta_transforms)
        return len(self.image_paths)

    def __getitem__(self, idx):
        if self.tta_transforms:
            path_idx = idx // len(self.tta_transforms)
            transform_idx = idx % len(self.tta_transforms)
            transform = self.tta_transforms[transform_idx]
            image_path = self.image_paths[path_idx]
        else:
            image_path = self.image_paths[idx]
            transform = self.base_transform

        img = imread_rgb(Path(image_path))
        img = transform(image=img)["image"]

        forensic = build_forensic_channels(img)  # 6xHxW
        rgb = img.astype(np.float32) / 255.0
        rgb = (rgb - IMAGENET_MEAN) / IMAGENET_STD
        rgb = np.transpose(rgb, (2, 0, 1)).astype(np.float32)  # 3xHxW

        x = np.concatenate([rgb, forensic], axis=0).astype(np.float32)  # 9xHxW

        return torch.from_numpy(x), image_path

# ========================================================================================
# MODEL ARCHITECTURE
# ========================================================================================

class StemToRGB(nn.Module):
    """Stem adapter to convert 9 channels to 3."""

    def __init__(self, in_ch: int):
        super().__init__()
        self.proj = nn.Conv2d(in_ch, 3, kernel_size=1, bias=False)
        self.bn = nn.BatchNorm2d(3)
        self.act = nn.SiLU(inplace=True)

    def forward(self, x):
        return self.act(self.bn(self.proj(x)))

def build_model(checkpoint_path: str, arch_name: str = None):
    """Build model with automatic architecture detection."""

    def _direct(arch):
        return timm.create_model(arch, pretrained=False, in_chans=9,
                               num_classes=1, global_pool='avg')

    def _stem(arch):
        class ModelWithStem(nn.Module):
            def __init__(self):
                super().__init__()
                self.stem = StemToRGB(9)
                self.backbone = timm.create_model(arch, pretrained=False, in_chans=3,
                                                num_classes=1, global_pool='avg')

            def forward(self, x):
                x = self.stem(x)
                return self.backbone(x)

        return ModelWithStem()

    state = torch.load(checkpoint_path, map_location="cpu")
    sd = state.get("model", state)
    keys = list(sd.keys())

    if arch_name is None:
        model_name = Path(checkpoint_path).stem.lower()
        if "efficientnetv2_xl" in model_name:
            arch_name = "tf_efficientnetv2_xl"
        elif "efficientnetv2_l" in model_name:
            arch_name = "tf_efficientnetv2_l"
        elif "efficientnetv2_m" in model_name:
            arch_name = "tf_efficientnetv2_m"
        elif "efficientnetv2_s" in model_name:
            arch_name = "tf_efficientnetv2_s"
        elif "efficientnetv2_b3" in model_name:
            arch_name = "tf_efficientnetv2_b3"
        elif "efficientnetv2_b2" in model_name:
            arch_name = "tf_efficientnetv2_b2"
        elif "efficientnetv2_b1" in model_name:
            arch_name = "tf_efficientnetv2_b1"
        elif "efficientnetv2_b0" in model_name:
            arch_name = "tf_efficientnetv2_b0"
        else:
            arch_name = "tf_efficientnetv2_m"  # default

    if any(k.startswith("stem.") for k in keys):
        model = _stem(arch_name)
    else:
        model = _direct(arch_name)

    model.load_state_dict(sd, strict=False)
    return model

# ========================================================================================
# THRESHOLD LOADING
# ========================================================================================

def load_threshold_from_json(json_path: Path) -> Optional[float]:
    try:
        with open(json_path, "r", encoding="utf-8") as f:
            data = json.load(f)

        if isinstance(data, dict):
            if "best_threshold" in data and isinstance(data["best_threshold"], (int, float)):
                return float(data["best_threshold"])
            if "threshold" in data and isinstance(data["threshold"], (int, float)):
                return float(data["threshold"])

        if isinstance(data, (int, float)):
            return float(data)

    except Exception:
        pass

    return None

def guess_threshold_candidates_for_model(model_path: Path) -> List[Path]:
    """Generate candidate filenames for saved thresholds."""
    stem = model_path.stem  # e.g., efficientnetv2_m_best
    base = stem.replace("_best", "")
    folder = model_path.parent

    candidates = [
        folder / f"{stem}_threshold.json",
        folder / f"{base}_threshold.json",
        folder / "threshold.json",
    ]

    # Filter duplicates while preserving order
    seen = set()
    uniq: List[Path] = []
    for c in candidates:
        if c not in seen:
            seen.add(c)
            uniq.append(c)

    return uniq

def auto_load_model_threshold(model_path: str) -> Optional[float]:
    """Try to load a threshold matching the model."""
    model_p = Path(model_path)

    for cand in guess_threshold_candidates_for_model(model_p):
        if cand.exists():
            thr = load_threshold_from_json(cand)
            if thr is not None:
                return thr

    return None

# ========================================================================================
# COMMON PROCESSING FUNCTIONS
# ========================================================================================

def create_dataloader(image_paths: List[str], img_size: int, use_tta: bool,
                     tta_augments: int, batch_size: int, num_workers: int) -> Tuple[DataLoader, Optional[List]]:
    """Create DataLoader with optional TTA transforms."""
    tta_transforms = None
    if use_tta:
        tta_transforms = make_tta_transforms(img_size, tta_augments)
        dataset = ImageDataset(image_paths, img_size, tta_transforms=tta_transforms)
    else:
        dataset = ImageDataset(image_paths, img_size, tta_transforms=None)

    dataloader = DataLoader(dataset, batch_size=batch_size, shuffle=False, 
                          num_workers=num_workers, pin_memory=True)
    return dataloader, tta_transforms

def print_threshold_info(threshold: Optional[float], threshold_mode: str, show_progress: bool):
    """Print threshold information."""
    if not show_progress:
        return
        
    if threshold_mode == "per-model" or threshold is None:
        print("INFO: USED_THRESHOLD: per-model")
    else:
        print(f"INFO: USED_THRESHOLD: {threshold:.3f}")

def calculate_ensemble_vote(model_probs: Dict[str, float], model_thresholds: Dict[str, float], 
                          override_threshold: Optional[float], confidence_threshold: float) -> Tuple[float, float]:
    """Calculate ensemble vote with proper weighting."""
    weighted_sum = 0.0
    weight_sum = 0.0
    unweighted_votes = []

    for model_name, prob in model_probs.items():
        thr = float(override_threshold) if isinstance(override_threshold, (int, float)) else model_thresholds.get(model_name, 0.25)
        decision = 1.0 if prob >= thr else 0.0
        conf = max(0.0, min(1.0, 2.0 * abs(prob - thr)))

        if conf >= confidence_threshold:
            weighted_sum += decision * conf
            weight_sum += conf

        unweighted_votes.append(decision)

    if weight_sum > 0:
        vote = weighted_sum / weight_sum
    else:
        vote = float(np.mean(unweighted_votes))

    vote = max(0.0, min(1.0, vote))  # Clip to [0,1]
    confidence = max(0.0, min(1.0, 2.0 * abs(vote - 0.5)))
    
    return vote, confidence

def print_image_result(image_path: str, pred_label: str, prob: float, conf: float, 
                      true_label: Optional[str] = None):
    """Print result for a single image."""
    img_name = Path(image_path).name
    
    if true_label is not None:
        if true_label != pred_label:
            print(f"MISCLASSIFIED: img: {img_name} true: {true_label} pred: {pred_label} prob: {prob:.3f} conf: {conf:.3f}")
        else:
            print(f"img: {img_name} true: {true_label} pred: {pred_label} prob: {prob:.3f} conf: {conf:.3f}")
    else:
        print(f"img: {img_name} pred: {pred_label} prob: {prob:.3f} conf: {conf:.3f}")

# ========================================================================================
# PREDICTOR CLASSES
# ========================================================================================

class SinglePredictor:
    """Single model predictor."""

    def __init__(self, model_path: str, img_size: int = DEFAULT_IMG_SIZE,
                 tta_augments: int = DEFAULT_TTA_AUGS):
        self.model_path = model_path
        self.img_size = img_size
        self.tta_augments = tta_augments

        self.model = build_model(model_path)
        self.model.to(device)
        self.model.eval()

        self.default_threshold = auto_load_model_threshold(model_path)

    def predict_realtime(self, image_paths: List[str], use_tta: bool = True,
                        threshold: Optional[float] = None, show_progress: bool = True,
                        batch_size: int = DEFAULT_BATCH_SIZE, num_workers: int = DEFAULT_WORKERS,
                        y_true: Optional[List[int]] = None) -> Tuple[List[float], Dict]:
        """Predict with real-time output for each image."""
        
        # Determine threshold
        thr = float(threshold) if threshold is not None else (
            self.default_threshold if self.default_threshold is not None else 0.5
        )
        
        # Print threshold info
        print_threshold_info(thr, "global", show_progress)

        # Create dataloader
        try:
            dataloader, tta_transforms = create_dataloader(
                image_paths, self.img_size, use_tta, self.tta_augments, 
                batch_size, num_workers
            )
        except Exception as e:
            if "out of memory" in str(e).lower() and batch_size > 1:
                print(f"WARNING: OOM with batch_size={batch_size}, trying batch_size=1")
                dataloader, tta_transforms = create_dataloader(
                    image_paths, self.img_size, use_tta, self.tta_augments, 
                    1, num_workers
                )
            else:
                raise

        probs_all = []
        processed_images = 0
        fake_count = 0
        real_count = 0
        y_pred = []
        conf_list = []

        with torch.no_grad():
            for batch_x, batch_paths in dataloader:
                batch_x = batch_x.to(device, non_blocking=True)
                logits = self.model(batch_x).view(-1, 1)
                probs = torch.sigmoid(logits).cpu().numpy().flatten()
                probs_all.extend(probs)

                # Process batch results immediately
                if use_tta:
                    n_augments_actual = len(tta_transforms)
                    batch_images = len(probs) // n_augments_actual

                    for i in range(batch_images):
                        if processed_images < len(image_paths):
                            # Get TTA predictions for this image
                            start_idx = i * n_augments_actual
                            end_idx = start_idx + n_augments_actual
                            image_probs = probs[start_idx:end_idx]
                            final_prob = np.mean(image_probs)

                            # Classify and print
                            pred_label = "FAKE" if final_prob >= thr else "REAL"
                            conf = max(0.0, min(1.0, 2.0 * abs(final_prob - thr)))

                            if show_progress:
                                true_label = None
                                if y_true is not None:
                                    true_label = "FAKE" if y_true[processed_images] == 1 else "REAL"
                                print_image_result(image_paths[processed_images], pred_label, final_prob, conf, true_label)

                            if pred_label == "FAKE":
                                fake_count += 1
                            else:
                                real_count += 1
                                
                            y_pred.append(1 if pred_label == "FAKE" else 0)
                            conf_list.append(conf)
                            processed_images += 1
                else:
                    # No TTA - direct mapping
                    for i, prob in enumerate(probs):
                        if processed_images < len(image_paths):
                            pred_label = "FAKE" if prob >= thr else "REAL"
                            conf = max(0.0, min(1.0, 2.0 * abs(prob - thr)))

                            if show_progress:
                                true_label = None
                                if y_true is not None:
                                    true_label = "FAKE" if y_true[processed_images] == 1 else "REAL"
                                print_image_result(image_paths[processed_images], pred_label, prob, conf, true_label)

                            if pred_label == "FAKE":
                                fake_count += 1
                            else:
                                real_count += 1
                                
                            y_pred.append(1 if pred_label == "FAKE" else 0)
                            conf_list.append(conf)
                            processed_images += 1

        # Final aggregation for TTA
        if use_tta:
            n_samples = len(image_paths)
            n_augments_actual = len(tta_transforms)
            probs_all = np.array(probs_all).reshape(n_samples, n_augments_actual)
            probs_all = probs_all.mean(axis=1).tolist()

        # Show summary
        if show_progress and len(image_paths) > 1:
            print(f"\nSUMMARY: {real_count} REAL, {fake_count} FAKE")

        results = {
            'y_pred': y_pred,
            'conf_list': conf_list,
            'fake_count': fake_count,
            'real_count': real_count
        }

        return probs_all, results

    def predict(self, image_paths: List[str], use_tta: bool = True) -> List[float]:
        """Legacy predict method for compatibility."""
        probs, _ = self.predict_realtime(image_paths, use_tta, show_progress=False)
        return probs

class EnsemblePredictor:
    """Ensemble predictor for multiple models with per-model thresholds."""

    def __init__(self, model_dir: str, confidence_threshold: float = 0.1,
                 img_size: int = DEFAULT_IMG_SIZE, tta_augments: int = DEFAULT_TTA_AUGS):
        self.model_dir = Path(model_dir)
        self.confidence_threshold = confidence_threshold
        self.img_size = img_size
        self.tta_augments = tta_augments

        self.models: Dict[str, nn.Module] = {}
        self.model_paths: Dict[str, Path] = {}
        self.model_thresholds: Dict[str, float] = {}

        model_files: List[Path] = []
        for ext in MODEL_EXTENSIONS:
            model_files.extend(self.model_dir.glob(f"*{ext}"))

        if not model_files:
            raise RuntimeError(f"No model files found in {model_dir}")

        for model_path in sorted(model_files):
            model_name = model_path.stem
            try:
                model = build_model(str(model_path))
                model.to(device)
                model.eval()

                self.models[model_name] = model
                self.model_paths[model_name] = model_path

                thr = auto_load_model_threshold(str(model_path))
                # Default to 0.25 if no JSON/threshold found
                self.model_thresholds[model_name] = float(thr) if isinstance(thr, (int, float)) else 0.25

            except Exception as e:
                print(f"Warning: Could not load {model_path}: {e}")

        if not self.models:
            raise RuntimeError("No models could be loaded!")

    def predict_realtime(self, image_paths: List[str], use_tta: bool = True,
                        override_threshold: Optional[float] = None, show_progress: bool = True,
                        batch_size: int = DEFAULT_BATCH_SIZE, num_workers: int = DEFAULT_WORKERS,
                        y_true: Optional[List[int]] = None) -> Tuple[List[float], Dict]:
        """Predict with real-time output for each image."""
        
        # Print threshold info
        threshold_mode = "global" if override_threshold is not None else "per-model"
        print_threshold_info(override_threshold, threshold_mode, show_progress)

        # Create dataloader
        try:
            dataloader, tta_transforms = create_dataloader(
                image_paths, self.img_size, use_tta, self.tta_augments, 
                batch_size, num_workers
            )
        except Exception as e:
            if "out of memory" in str(e).lower() and batch_size > 1:
                print(f"WARNING: OOM with batch_size={batch_size}, trying batch_size=1")
                dataloader, tta_transforms = create_dataloader(
                    image_paths, self.img_size, use_tta, self.tta_augments, 
                    1, num_workers
                )
            else:
                raise

        all_predictions: Dict[str, List[float]] = {}
        
        # Initialize predictions for all models
        for model_name in self.models.keys():
            all_predictions[model_name] = []

        processed_images = 0
        fake_count = 0
        real_count = 0
        y_pred = []
        conf_list = []

        with torch.no_grad():
            for batch_x, batch_paths in dataloader:
                batch_x = batch_x.to(device, non_blocking=True)

                # Process batch through all models simultaneously
                batch_model_probs = {}
                for model_name, model in self.models.items():
                    logits = model(batch_x).view(-1, 1)
                    probs = torch.sigmoid(logits).cpu().numpy().flatten()
                    all_predictions[model_name].extend(probs)
                    batch_model_probs[model_name] = probs

                # Process batch results immediately
                if use_tta:
                    n_augments_actual = len(tta_transforms)
                    batch_images = len(batch_model_probs[list(self.models.keys())[0]]) // n_augments_actual

                    for i in range(batch_images):
                        if processed_images < len(image_paths):
                            # Aggregate TTA predictions for this image across all models
                            model_probs = {}
                            for model_name in self.models.keys():
                                start_idx = i * n_augments_actual
                                end_idx = start_idx + n_augments_actual
                                image_probs = batch_model_probs[model_name][start_idx:end_idx]
                                model_probs[model_name] = np.mean(image_probs)

                            # Calculate ensemble vote using consistent weighting
                            vote, confidence = calculate_ensemble_vote(
                                model_probs, self.model_thresholds, override_threshold, self.confidence_threshold
                            )

                            pred_label = "FAKE" if vote >= 0.5 else "REAL"

                            if show_progress:
                                true_label = None
                                if y_true is not None:
                                    true_label = "FAKE" if y_true[processed_images] == 1 else "REAL"
                                print_image_result(image_paths[processed_images], pred_label, vote, confidence, true_label)

                            if pred_label == "FAKE":
                                fake_count += 1
                            else:
                                real_count += 1
                                
                            y_pred.append(1 if pred_label == "FAKE" else 0)
                            conf_list.append(confidence)
                            processed_images += 1
                else:
                    # No TTA - direct processing
                    batch_size_actual = len(batch_model_probs[list(self.models.keys())[0]])

                    for i in range(batch_size_actual):
                        if processed_images < len(image_paths):
                            model_probs = {}
                            for model_name in self.models.keys():
                                model_probs[model_name] = batch_model_probs[model_name][i]

                            # Calculate ensemble vote using consistent weighting
                            vote, confidence = calculate_ensemble_vote(
                                model_probs, self.model_thresholds, override_threshold, self.confidence_threshold
                            )

                            pred_label = "FAKE" if vote >= 0.5 else "REAL"

                            if show_progress:
                                true_label = None
                                if y_true is not None:
                                    true_label = "FAKE" if y_true[processed_images] == 1 else "REAL"
                                print_image_result(image_paths[processed_images], pred_label, vote, confidence, true_label)

                            if pred_label == "FAKE":
                                fake_count += 1
                            else:
                                real_count += 1
                                
                            y_pred.append(1 if pred_label == "FAKE" else 0)
                            conf_list.append(confidence)
                            processed_images += 1

        # TTA Aggregation for all models (for final return values)
        if use_tta:
            n_samples = len(image_paths)
            n_augments_actual = len(tta_transforms)
            for model_name in all_predictions.keys():
                probs_array = np.array(all_predictions[model_name]).reshape(n_samples, n_augments_actual)
                all_predictions[model_name] = probs_array.mean(axis=1).tolist()

        # Final ensemble calculation for return (consistent with real-time)
        ensemble_votes: List[float] = []
        for i in range(len(image_paths)):
            model_probs = {}
            for model_name in self.models.keys():
                model_probs[model_name] = all_predictions[model_name][i]
            
            vote, _ = calculate_ensemble_vote(
                model_probs, self.model_thresholds, override_threshold, self.confidence_threshold
            )
            ensemble_votes.append(vote)

        # Show summary
        if show_progress and len(image_paths) > 1:
            print(f"\nSUMMARY: {real_count} REAL, {fake_count} FAKE")

        results = {
            'y_pred': y_pred,
            'conf_list': conf_list,
            'fake_count': fake_count,
            'real_count': real_count
        }

        return ensemble_votes, results

    def predict(self, image_paths: List[str], use_tta: bool = True,
            override_threshold: Optional[float] = None) -> List[float]:
        """Legacy predict method for compatibility."""
        probs, _ = self.predict_realtime(image_paths, use_tta, override_threshold, show_progress=False)
        return probs

    def default_threshold(self) -> float:
        """Keep for compatibility; average of available per-model thresholds."""
        thrs = [t for t in self.model_thresholds.values() if isinstance(t, (int, float))]
        return float(np.mean(thrs)) if thrs else 0.5

# ========================================================================================
# UTILITY & PRINTING
# ========================================================================================

def find_images(path: str) -> List[str]:
    """Find all images in a directory or return single file."""
    path_obj = Path(path)

    if path_obj.is_file():
        if path_obj.suffix.lower() in SUPPORTED_EXTENSIONS:
            return [str(path_obj)]
        else:
            return []

    image_paths: List[str] = []
    for ext in SUPPORTED_EXTENSIONS:
        image_paths.extend(glob(str(path_obj / f"**/*{ext}"), recursive=True))
        image_paths.extend(glob(str(path_obj / f"**/*{ext.upper()}"), recursive=True))

    return sorted(list(set(image_paths)))

def find_first_model_in_script_dir() -> Optional[str]:
    """Find first .pt/.pth model in the same directory as this script."""
    script_dir = Path(__file__).parent

    for ext in MODEL_EXTENSIONS:
        model_files = list(script_dir.glob(f"*{ext}"))
        if model_files:
            return str(model_files[0])

    return None

def parse_image_inputs(user_input: str) -> List[str]:
    """Parse user input to extract multiple image paths with robust cross-platform handling."""
    image_paths = []
    
    # Try different parsing strategies
    parsing_strategies = [
        # Strategy 1: Try shlex.split (works well on Unix-like systems)
        lambda s: shlex.split(s),
        # Strategy 2: Simple split on spaces (fallback)
        lambda s: s.split(),
        # Strategy 3: Split on spaces but keep quoted strings together
        lambda s: [part.strip('"\'') for part in s.split() if part.strip()]
    ]
    
    parsed_paths = None
    for strategy in parsing_strategies:
        try:
            parsed_paths = strategy(user_input)
            break
        except (ValueError, Exception):
            continue
    
    if parsed_paths is None:
        # Ultimate fallback: just split on whitespace
        parsed_paths = user_input.split()
    
    for path in parsed_paths:
        # Clean up the path
        path = path.strip().strip('"\'')
        
        # Convert to absolute path to handle relative paths correctly
        try:
            abs_path = os.path.abspath(path)
            if os.path.exists(abs_path):
                found_images = find_images(abs_path)
                if found_images:
                    image_paths.extend(found_images)
                else:
                    print(f"WARNING: No images found in '{path}'")
            else:
                # Try original path if absolute path doesn't work
                if os.path.exists(path):
                    found_images = find_images(path)
                    if found_images:
                        image_paths.extend(found_images)
                    else:
                        print(f"WARNING: No images found in '{path}'")
                else:
                    print(f"WARNING: Path '{path}' does not exist")
        except Exception as e:
            print(f"WARNING: Error processing path '{path}': {e}")
    
    return image_paths

def run_test_evaluation(test_dir: str, predictor, threshold: Optional[float] = 0.5,
                       use_tta: bool = True, threshold_mode: str = "global",
                       batch_size: int = DEFAULT_BATCH_SIZE, num_workers: int = DEFAULT_WORKERS):
    """Run test evaluation with ground truth - results shown in real-time, stats at end."""
    test_path = Path(test_dir)
    real_dir = test_path / "real"
    fake_dir = test_path / "fake"

    if not real_dir.exists() or not fake_dir.exists():
        print(f"ERROR: Test directory must contain 'real' and 'fake' subdirectories")
        return

    real_images = find_images(str(real_dir))
    fake_images = find_images(str(fake_dir))

    if not real_images and not fake_images:
        print("ERROR: No images found in test directories")
        return

    print(f"INFO: Found {len(real_images)} real and {len(fake_images)} fake test images")
    print("INFO: Running predictions...")

    all_images = real_images + fake_images
    y_true = [0] * len(real_images) + [1] * len(fake_images)  # 0=real, 1=fake

    # Use the unified predict_realtime method with y_true for proper output
    if isinstance(predictor, EnsemblePredictor):
        override_thr = threshold if (threshold_mode == "global" and isinstance(threshold, (int, float))) else None
        y_pred_score, results = predictor.predict_realtime(
            all_images, use_tta=use_tta, override_threshold=override_thr, 
            show_progress=True, batch_size=batch_size, num_workers=num_workers, y_true=y_true
        )
    else:
        y_pred_score, results = predictor.predict_realtime(
            all_images, use_tta=use_tta, threshold=threshold,
            show_progress=True, batch_size=batch_size, num_workers=num_workers, y_true=y_true
        )

    y_pred = results['y_pred']
    conf_list = results['conf_list']

    # Calculate and display final statistics
    accuracy = accuracy_score(y_true, y_pred)
    precision, recall, f1, _ = precision_recall_fscore_support(y_true, y_pred, average='binary')
    cm = confusion_matrix(y_true, y_pred)

    print(f"\n{'='*80}")
    print("FINAL TEST STATISTICS:")
    print(f"{'='*80}")
    print(f"ACCURACY: {accuracy:.4f}")
    print(f"PRECISION: {precision:.4f}")
    print(f"RECALL: {recall:.4f}")
    print(f"F1_SCORE: {f1:.4f}")
    print("CONFUSION_MATRIX (rows=true, cols=pred):")
    print(f" True Real (TN): {cm[0,0]}")
    print(f" False Fake (FP): {cm[0,1]}")
    print(f" False Real (FN): {cm[1,0]}")
    print(f" True Fake (TP): {cm[1,1]}")

# ========================================================================================
# LIVE MODE
# ========================================================================================

def run_live_mode(predictor, threshold: Optional[float], threshold_mode: str, 
                 results_are_votes: bool, use_tta: bool, batch_size: int, num_workers: int):
    """Interactive live mode for continuous processing."""
    print("\n" + "="*80)
    print("LIVE MODE AKTIV")
    print("="*80)
    print("Models sind geladen und bereit!")
    print("Befehle:")
    print("  - Bildpfad(e) oder Ordnerpfad eingeben für Klassifizierung")
    print("    Beispiel: image1.jpg image2.jpg oder 'path with spaces/image.jpg'")
    print("  - 'test path/to/testfolder' für Test-Evaluation")
    print("  - 'tta' zum Ein-/Ausschalten von TTA (aktuell: {})".format("EIN" if use_tta else "AUS"))
    print("  - 'batch <size>' zum Ändern der Batch-Size (aktuell: {})".format(batch_size))
    print("  - 'workers <count>' zum Ändern der Worker-Anzahl (aktuell: {})".format(num_workers))
    print("  - 'info' für Modellinformationen")
    print("  - 'exit' zum Beenden")
    print("-" * 80)

    current_tta = use_tta
    current_batch_size = batch_size
    current_workers = num_workers

    while True:
        try:
            user_input = input("\n> ").strip()

            if not user_input:
                continue

            if user_input.lower() in ['exit', 'quit', 'q']:
                print("Live-Modus beendet.")
                break

            elif user_input.lower() == 'tta':
                current_tta = not current_tta
                print(f"TTA {'aktiviert' if current_tta else 'deaktiviert'}")
                continue

            elif user_input.lower().startswith('batch '):
                try:
                    new_batch_size = int(user_input[6:].strip())
                    if new_batch_size > 0:
                        current_batch_size = new_batch_size
                        print(f"Batch-Size geändert auf: {current_batch_size}")
                    else:
                        print("ERROR: Batch-Size muss größer als 0 sein")
                except ValueError:
                    print("ERROR: Ungültige Batch-Size. Verwenden Sie: batch <number>")
                continue

            elif user_input.lower().startswith('workers '):
                try:
                    new_workers = int(user_input[8:].strip())
                    if new_workers >= 0:
                        current_workers = new_workers
                        print(f"Workers geändert auf: {current_workers}")
                    else:
                        print("ERROR: Workers muss 0 oder größer sein")
                except ValueError:
                    print("ERROR: Ungültige Workers-Anzahl. Verwenden Sie: workers <number>")
                continue

            elif user_input.lower() == 'info':
                if isinstance(predictor, EnsemblePredictor):
                    print(f"Ensemble mit {len(predictor.models)} Models:")
                    for name in predictor.models.keys():
                        thr = predictor.model_thresholds.get(name, 0.25)
                        print(f"  - {name} (threshold: {thr:.3f})")
                else:
                    print(f"Single Model: {predictor.model_path}")
                    print(f"Threshold: {predictor.default_threshold}")
                
                print(f"Einstellungen: TTA={'EIN' if current_tta else 'AUS'}, Batch-Size={current_batch_size}, Workers={current_workers}")
                continue

            elif user_input.lower().startswith('test '):
                # Extract test directory path
                test_path = user_input[5:].strip().strip('"\'')
                if not os.path.exists(test_path):
                    print(f"ERROR: Test-Ordner '{test_path}' existiert nicht")
                    continue

                print(f"INFO: Starte Test-Evaluation für '{test_path}'...")
                run_test_evaluation(test_path, predictor, threshold, use_tta=current_tta, 
                                  threshold_mode=threshold_mode, batch_size=current_batch_size, 
                                  num_workers=current_workers)
                continue

            # Parse multiple image paths
            image_paths = parse_image_inputs(user_input)
            if not image_paths:
                print("WARNING: Keine gültigen Bilder gefunden")
                continue

            print(f"INFO: Verarbeite {len(image_paths)} Bilder...")

            # Run prediction with real-time output
            if isinstance(predictor, EnsemblePredictor):
                predictor.predict_realtime(
                    image_paths,
                    use_tta=current_tta,
                    override_threshold=(threshold if threshold_mode == "global" else None),
                    show_progress=True,
                    batch_size=current_batch_size,
                    num_workers=current_workers
                )
            else:
                predictor.predict_realtime(
                    image_paths,
                    use_tta=current_tta,
                    threshold=threshold,
                    show_progress=True,
                    batch_size=current_batch_size,
                    num_workers=current_workers
                )

        except KeyboardInterrupt:
            print("\nLive-Modus beendet (Ctrl+C)")
            break
        except Exception as e:
            print(f"ERROR: {e}")
            continue

# ========================================================================================
# MAIN
# ========================================================================================

def main():
    parser = argparse.ArgumentParser(
        description="DeepFake Detector - Classify images as real or fake",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )

    parser.add_argument("inputs", nargs="*",
                       help="Image files or directories to classify")
    parser.add_argument("--threshold", type=float, default=None,
                       help="Classification threshold; if omitted, uses saved model threshold or default")
    parser.add_argument("--model", help="Single model file to use")
    parser.add_argument("--models", help="Directory with multiple models for ensemble")
    parser.add_argument("--test", help="Test folder with 'real' and 'fake' subdirectories")
    parser.add_argument("--tta", action="store_true", help="Enable test-time augmentation (TTA)")
    parser.add_argument("--tta-augments", type=int, default=DEFAULT_TTA_AUGS,
                       help=f"Number of TTA augmentations (default: {DEFAULT_TTA_AUGS})")
    parser.add_argument("--img-size", type=int, default=DEFAULT_IMG_SIZE,
                       help=f"Inference image size (default: {DEFAULT_IMG_SIZE})")
    parser.add_argument("--batch-size", type=int, default=DEFAULT_BATCH_SIZE,
                       help=f"Batch size for inference (default: {DEFAULT_BATCH_SIZE})")
    parser.add_argument("--workers", type=int, default=DEFAULT_WORKERS,
                       help=f"Number of data loading workers (default: {DEFAULT_WORKERS})")
    parser.add_argument("--live", action="store_true",
                       help="Enter live mode - models stay loaded for continuous processing")

    args = parser.parse_args()

    img_size = int(args.img_size)
    tta_augments = int(args.tta_augments)
    batch_size = int(args.batch_size)
    num_workers = int(args.workers)
    use_tta = args.tta

    image_paths: List[str] = []

    # Für Live-Modus keine Bilder beim Start erforderlich
    if not args.live:
        if args.test:
            pass
        elif args.inputs:
            for inp in args.inputs:
                if os.path.exists(inp):
                    found_images = find_images(inp)
                    if found_images:
                        image_paths.extend(found_images)
                    else:
                        print(f"WARNING: No images found in {inp}")
                else:
                    print(f"WARNING: {inp} does not exist")
        else:
            print("ERROR: Please specify input images, directories, or use --test")
            return

    predictor = None
    used_threshold: Optional[float] = 0.5  # default fallback
    threshold_mode: str = "global"
    results_are_votes: bool = False  # whether predictions represent vote shares

    if args.models:
        if not os.path.exists(args.models):
            print(f"ERROR: Models directory {args.models} does not exist")
            return

        try:
            print("INFO: Loading models...")
            predictor = EnsemblePredictor(model_dir=args.models,
                                        confidence_threshold=0.1,
                                        img_size=img_size,
                                        tta_augments=tta_augments)
            print(f"INFO: Using ensemble with {len(predictor.models)} models from {args.models}")

            if args.threshold is not None:
                used_threshold = float(args.threshold)
                threshold_mode = "global"
            else:
                used_threshold = None
                threshold_mode = "per-model"

            results_are_votes = True

        except Exception as e:
            print(f"ERROR: Could not load ensemble: {e}")
            return

    elif args.model:
        if not os.path.exists(args.model):
            print(f"ERROR: Model path {args.model} does not exist")
            return

        try:
            print("INFO: Loading model...")
            predictor = SinglePredictor(model_path=args.model,
                                      img_size=img_size,
                                      tta_augments=tta_augments)
            print(f"INFO: Using single model {args.model}")

            if args.threshold is not None:
                used_threshold = float(args.threshold)
            else:
                used_threshold = predictor.default_threshold if predictor.default_threshold is not None else 0.5

            threshold_mode = "global"
            results_are_votes = False

        except Exception as e:
            print(f"ERROR: Could not load model: {e}")
            return

    else:
        auto_model = find_first_model_in_script_dir()
        if auto_model:
            try:
                print("INFO: Loading auto-detected model...")
                predictor = SinglePredictor(model_path=auto_model,
                                          img_size=img_size,
                                          tta_augments=tta_augments)
                print(f"INFO: Using auto-detected model {auto_model}")

                if args.threshold is not None:
                    used_threshold = float(args.threshold)
                else:
                    used_threshold = predictor.default_threshold if predictor.default_threshold is not None else 0.5

                threshold_mode = "global"
                results_are_votes = False

            except Exception as e:
                print(f"ERROR: Could not load auto-detected model: {e}")
                return
        else:
            print("ERROR: No model specified and no .pt files found in script directory")
            return

    # Live-Modus starten
    if args.live:
        run_live_mode(predictor, used_threshold, threshold_mode, results_are_votes, 
                     use_tta, batch_size, num_workers)
        return

    # Normaler Modus
    if args.test:
        run_test_evaluation(args.test, predictor, used_threshold, use_tta=use_tta, 
                          threshold_mode=threshold_mode, batch_size=batch_size, num_workers=num_workers)
    else:
        if not image_paths:
            print("INFO: No images found to process")
            return

        print(f"INFO: Processing {len(image_paths)} images...")

        try:
            if isinstance(predictor, EnsemblePredictor):
                predictor.predict_realtime(image_paths, use_tta=use_tta,
                                         override_threshold=(used_threshold if threshold_mode == "global" else None),
                                         show_progress=True, batch_size=batch_size, num_workers=num_workers)
            else:
                predictor.predict_realtime(image_paths, use_tta=use_tta,
                                         threshold=used_threshold,
                                         show_progress=True, batch_size=batch_size, num_workers=num_workers)

        except Exception as e:
            print(f"ERROR: Prediction failed: {e}")
            return

if __name__ == "__main__":
    main()