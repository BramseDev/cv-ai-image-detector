#!/usr/bin/env python3
"""
Farbbalance-Analysator für Bilder

Untersucht:
- RGB-Kanalverteilung
- Histogramme
- Farbton/Sättigung/Helligkeit (HSV)
- Farbdynamik und Kontrast
- Heuristischer KI-Farbscore

Usage: python3 analyze_colorbalance.py <image_path>
"""

import cv2
import numpy as np
import json
import sys
import os
from scipy import stats


class ColorBalanceConfig:
    def __init__(self,
                 histogram_std_threshold=60.0,      # Statt 15.0
                 channel_diff_threshold=20.0,       # Statt 30.0
                 dominant_color_threshold=0.3,      # Statt 0.4
                 saturation_threshold=0.6,          # Statt 0.3
                 brightness_threshold=0.15,         # Statt 0.2
                 contrast_threshold=40.0,           # Statt 50.0
                 color_cast_threshold=15.0,         # Statt 20.0
                 imbalance_weight_histogram=0.35,   # Statt 0.3
                 imbalance_weight_channel=0.3,      # Statt 0.25
                 imbalance_weight_dominant=0.2,     # Unverändert
                 imbalance_weight_saturation=0.1,   # Statt 0.15
                 imbalance_weight_brightness=0.05): # Statt 0.1
        self.histogram_std_threshold = histogram_std_threshold
        self.channel_diff_threshold = channel_diff_threshold
        self.dominant_color_threshold = dominant_color_threshold
        self.saturation_threshold = saturation_threshold
        self.brightness_threshold = brightness_threshold
        self.contrast_threshold = contrast_threshold
        self.color_cast_threshold = color_cast_threshold
        self.imbalance_weight_histogram = imbalance_weight_histogram
        self.imbalance_weight_channel = imbalance_weight_channel
        self.imbalance_weight_dominant = imbalance_weight_dominant
        self.imbalance_weight_saturation = imbalance_weight_saturation
        self.imbalance_weight_brightness = imbalance_weight_brightness


def analyze_color_balance(image_path, config=None):
    """Führt eine vollständige Farbbalance-Analyse durch."""
    if config is None:
        config = ColorBalanceConfig()

    # Bild laden
    img = cv2.imread(image_path)
    if img is None:
        return {"error": f"Kann Bild nicht laden: {image_path}"}

    # Bild in verschiedene Farbräume konvertieren
    rgb_img = cv2.cvtColor(img, cv2.COLOR_BGR2RGB)
    hsv_img = cv2.cvtColor(img, cv2.COLOR_BGR2HSV)
    lab_img = cv2.cvtColor(img, cv2.COLOR_BGR2LAB)

    results = {}

    # 1. Grundlegende RGB-Statistik
    r, g, b = cv2.split(rgb_img)
    channels = {"red": r, "green": g, "blue": b}

    basic_stats = {}
    for name, channel in channels.items():
        basic_stats[name] = {
            "mean": float(np.mean(channel)),
            "std": float(np.std(channel)),
            "min": int(np.min(channel)),
            "max": int(np.max(channel)),
            "median": float(np.median(channel)),
            "dynamic_range": int(np.max(channel) - np.min(channel))
        }

    # 2. Farbkanal-Verhältnisse
    ratios = {
        "r_to_g": float(np.mean(r) / np.mean(g)) if np.mean(g) > 0 else 0,
        "r_to_b": float(np.mean(r) / np.mean(b)) if np.mean(b) > 0 else 0,
        "g_to_b": float(np.mean(g) / np.mean(b)) if np.mean(b) > 0 else 0
    }

    # 3. HSV-Analyse
    h, s, v = cv2.split(hsv_img)
    hsv_stats = {
        "hue": {
            "mean": float(np.mean(h)),
            "std": float(np.std(h)),
            "mode": float(stats.mode(h.flatten(), keepdims=True)[0][0])
        },
        "saturation": {
            "mean": float(np.mean(s)),
            "std": float(np.std(s)),
            "distribution": [
                float(np.percentile(s, 25)),  # Unteres Quartil
                float(np.percentile(s, 50)),  # Median
                float(np.percentile(s, 75))   # Oberes Quartil
            ]
        },
        "value": {
            "mean": float(np.mean(v)),
            "std": float(np.std(v)),
            "dynamic_range": int(np.max(v) - np.min(v))
        }
    }

    # 4. Kontrastanalyse (LAB)
    l, a, b_channel = cv2.split(lab_img)
    contrast = {
        "luminance_std": float(np.std(l)),
        "a_channel_std": float(np.std(a)),
        "b_channel_std": float(np.std(b_channel))
    }

    # 5. Histogramm-Charakteristik
    histogram_features = {}
    for name, channel in channels.items():
        hist = cv2.calcHist([channel], [0], None, [256], [0, 256])
        hist = hist.flatten() / hist.sum()  # Normalisieren

        # Entropie
        non_zero = hist[hist > 0]
        entropy = -np.sum(non_zero * np.log2(non_zero))

        # Peaks im Histogramm
        smoothed = np.convolve(hist, np.ones(5)/5, mode='same')
        peaks = np.sum((smoothed[1:-1] > smoothed[:-2]) & (smoothed[1:-1] > smoothed[2:]))

        histogram_features[name] = {
            "entropy": float(entropy),
            "peaks": int(peaks),
            "skewness": float(stats.skew(channel.flatten()))
        }

    # 6. Farbungleichgewicht-Analyse mit Config
    channel_stds = [basic_stats[c]["std"] for c in ["red", "green", "blue"]]
    std_ratio = max(channel_stds) / (min(channel_stds) + 0.001)

    imbalance_indicators = []

    # Verschärfte Imbalance-Erkennung
    if std_ratio > 1.5:  # Statt 2.0
        imbalance_indicators.append("uneven_channel_variation")

    # Histogram Imbalance (sensibler)
    if basic_stats["red"]["std"] < config.histogram_std_threshold:
        imbalance_indicators.append("low_red_variation")
    if basic_stats["green"]["std"] < config.histogram_std_threshold:
        imbalance_indicators.append("low_green_variation")
    if basic_stats["blue"]["std"] < config.histogram_std_threshold:
        imbalance_indicators.append("low_blue_variation")

    # Channel Ratio Imbalance (sensibler)
    if abs(ratios["r_to_g"] - 1.0) > 0.15:  # Statt 0.5
        imbalance_indicators.append("unnatural_rg_ratio")
    if abs(ratios["r_to_b"] - 1.0) > 0.15:  # Statt 0.5
        imbalance_indicators.append("unnatural_rb_ratio")
    if abs(ratios["g_to_b"] - 1.0) > 0.15:  # Statt 0.5
        imbalance_indicators.append("unnatural_gb_ratio")

    # Saturation Imbalance
    if hsv_stats["saturation"]["std"] < config.saturation_threshold * 100:
        imbalance_indicators.append("uniform_saturation")

    if contrast["luminance_std"] < config.contrast_threshold:
        imbalance_indicators.append("low_contrast")

    # Imbalance Score berechnen
    imbalance_score = float(len(imbalance_indicators) / 8.0)  # Normalisiert auf max 8 Indikatoren

    # 7. KI-Verdacht auf Farbbasis (verschärft)
    ai_color_score = 0.0

    if basic_stats["red"]["mean"] < 50:  # Statt 40
        ai_color_score += 0.2

    if histogram_features["red"]["skewness"] > 2.0:  # Statt 2.5
        ai_color_score += 0.2

    if ratios["r_to_g"] < 0.7 and ratios["r_to_b"] < 0.7:  # Statt 0.6
        ai_color_score += 0.2

    if hsv_stats["saturation"]["std"] > 55 or hsv_stats["saturation"]["std"] < 25:  # Verschärft
        ai_color_score += 0.2

    if imbalance_score == 0 and ai_color_score > 0:
        ai_color_score += 0.2  # Zu perfekte Balance kann KI-Hinweis sein

    # Verwende Imbalance Score direkt
    final_ai_score = max(imbalance_score, ai_color_score)

    results = {
        "basic_stats": basic_stats,
        "channel_ratios": ratios,
        "hsv_analysis": hsv_stats,
        "contrast": contrast,
        "histogram_features": histogram_features,
        "imbalance_indicators": imbalance_indicators,
        "imbalance_score": imbalance_score,
        "ai_color_score": round(min(final_ai_score, 1.0), 2)
    }

    return results


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"error": "Kein Bildpfad angegeben"}))
        sys.exit(1)

    image_path = sys.argv[1]
    if not os.path.isfile(image_path):
        print(json.dumps({"error": f"Datei nicht gefunden: {image_path}"}))
        sys.exit(1)

    try:
        results = analyze_color_balance(image_path)
        print(json.dumps(results))
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)