import cv2
import numpy as np
import json
import sys
import os  # HINZUGEFÜGT
from scipy import ndimage
from skimage import feature, filters

def make_json_serializable(obj):
    """Konvertiert numpy-Typen zu Python-Typen für JSON"""
    if isinstance(obj, dict):
        return {key: make_json_serializable(value) for key, value in obj.items()}
    elif isinstance(obj, list):
        return [make_json_serializable(item) for item in obj]
    elif isinstance(obj, np.bool_):
        return bool(obj)
    elif isinstance(obj, np.integer):
        return int(obj)
    elif isinstance(obj, np.floating):
        return float(obj)
    else:
        return obj

def detect_compression_artifacts(img):
    """Erkennt JPEG-Kompressionsartefakte"""
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    # Blockiness Detection (8x8 JPEG-Blöcke)
    blockiness_score = 0.0
    block_edges = 0

    # Vertikale Blockgrenzen
    for i in range(8, gray.shape[0], 8):
        if i < gray.shape[0] - 1:
            diff = np.abs(gray[i, :].astype(float) - gray[i-1, :].astype(float))
            block_edges += np.sum(diff > 10)

    # Horizontale Blockgrenzen
    for j in range(8, gray.shape[1], 8):
        if j < gray.shape[1] - 1:
            diff = np.abs(gray[:, j].astype(float) - gray[:, j-1].astype(float))
            block_edges += np.sum(diff > 10)

    total_possible_edges = (gray.shape[0] // 8) * gray.shape[1] + (gray.shape[1] // 8) * gray.shape[0]
    blockiness_score = block_edges / max(total_possible_edges, 1)

    return {
        'blockiness_score': float(blockiness_score),
        'ai_blockiness_indicator': bool(blockiness_score > 0.3)  # Explizit zu bool konvertieren
    }

def detect_ringing_artifacts(img):
    """Erkennt Ringing-Artefakte um Kanten"""
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    # Kantenerkennung
    edges = cv2.Canny(gray, 50, 150)

    # Sobel-Operator für Kantenrichtung
    sobel_x = cv2.Sobel(gray, cv2.CV_64F, 1, 0, ksize=3)
    sobel_y = cv2.Sobel(gray, cv2.CV_64F, 0, 1, ksize=3)
    sobel_mag = np.sqrt(sobel_x**2 + sobel_y**2)

    # Ringing-Score basierend auf Oszillationen nahe Kanten
    kernel = np.ones((3,3), np.uint8)
    edge_dilated = cv2.dilate(edges, kernel, iterations=2)

    # Bereich um Kanten analysieren
    ringing_regions = edge_dilated > 0
    if np.sum(ringing_regions) > 0:
        ringing_intensity = np.std(sobel_mag[ringing_regions])
    else:
        ringing_intensity = 0.0

    return {
        'ringing_intensity': float(ringing_intensity),
        'ai_ringing_indicator': bool(ringing_intensity > 15.0)
    }

def detect_color_bleeding(img):
    """Erkennt Farbblutungen typisch für starke Kompression"""
    # Konvertiere zu verschiedenen Farbräumen
    hsv = cv2.cvtColor(img, cv2.COLOR_BGR2HSV)
    yuv = cv2.cvtColor(img, cv2.COLOR_BGR2YUV)

    # Analysiere Chrominanz-Kanäle
    u_channel = yuv[:,:,1]
    v_channel = yuv[:,:,2]

    # Suche nach abrupten Farbübergängen
    u_grad = np.gradient(u_channel.astype(float))
    v_grad = np.gradient(v_channel.astype(float))

    color_gradient_magnitude = np.sqrt(u_grad[0]**2 + u_grad[1]**2 + v_grad[0]**2 + v_grad[1]**2)
    bleeding_score = np.mean(color_gradient_magnitude)

    return {
        'color_bleeding_score': float(bleeding_score),
        'ai_bleeding_indicator': bool(bleeding_score < 2.0)
    }

def detect_upscaling_artifacts(img):
    """Erkennt Upscaling-Artefakte von AI-Bildern"""
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    # Analysiere Hochfrequenz-Details
    # Laplacian für Schärfe-Detection
    laplacian = cv2.Laplacian(gray, cv2.CV_64F)
    sharpness_score = np.var(laplacian)

    # Analysiere Interpolations-Muster
    # Downscale und wieder upscale
    small = cv2.resize(gray, (gray.shape[1]//2, gray.shape[0]//2), interpolation=cv2.INTER_LINEAR)
    upscaled = cv2.resize(small, (gray.shape[1], gray.shape[0]), interpolation=cv2.INTER_LINEAR)

    # Differenz zwischen Original und Re-Interpolation
    interpolation_diff = np.mean(np.abs(gray.astype(float) - upscaled.astype(float)))

    return {
        'sharpness_score': float(sharpness_score),
        'interpolation_diff': float(interpolation_diff),
        'ai_upscaling_indicator': bool(sharpness_score > 1000 and interpolation_diff < 5)
    }

def detect_noise_reduction_artifacts(img):
    """Erkennt übermäßige Rauschunterdrückung"""
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    # Lokale Standardabweichung als Rauschmaß
    kernel = np.ones((5,5), np.float32) / 25
    mean_img = cv2.filter2D(gray.astype(np.float32), -1, kernel)
    variance_img = cv2.filter2D((gray.astype(np.float32) - mean_img)**2, -1, kernel)

    # Bereiche mit sehr niedrigem Rauschen finden
    low_noise_areas = variance_img < 1.0
    smooth_percentage = np.sum(low_noise_areas) / low_noise_areas.size

    # Textur-Analyse mit Local Binary Patterns
    lbp = feature.local_binary_pattern(gray, 8, 1, method='uniform')
    texture_uniformity = np.std(lbp)

    return {
        'smooth_percentage': float(smooth_percentage),
        'texture_uniformity': float(texture_uniformity),
        'ai_denoise_indicator': bool(smooth_percentage > 0.7 and texture_uniformity < 10)
    }

def detect_png_specific_artifacts(img, img_path):
    """PNG-spezifische Artefakt-Erkennung - KORRIGIERT für echte Fotos"""
    _, ext = os.path.splitext(img_path.lower())

    if ext != '.png':
        return {'png_analysis': 'not_applicable'}

    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    # 1. Farbkomplexität-Analyse (KORRIGIERT)
    unique_colors = len(np.unique(img.reshape(-1, img.shape[2]), axis=0))

    # 2. Dithering-Erkennung (KORRIGIERT - weniger aggressiv)
    # Berechne lokale Varianz in kleinen Patches
    patch_size = 8
    dithering_scores = []

    for i in range(0, gray.shape[0] - patch_size, patch_size * 2):  # Weniger Patches
        for j in range(0, gray.shape[1] - patch_size, patch_size * 2):
            patch = gray[i:i+patch_size, j:j+patch_size]
            local_var = np.var(patch)
            dithering_scores.append(local_var)

    dithering_score = np.mean(dithering_scores) if dithering_scores else 0

    # 3. Kantenschärfe-Analyse (NEU)
    edges = cv2.Canny(gray, 50, 150)
    edge_density = np.sum(edges > 0) / (edges.shape[0] * edges.shape[1])

    # KORRIGIERTE Bewertung für echte Fotos:
    # - Echte Fotos: hohe Farbkomplexität (>10000), moderate Kantendichte (0.05-0.2)
    # - AI-Bilder: oft niedrige Komplexität (<1000) oder künstlich hohe (>100000)

    ai_png_indicator = False

    if unique_colors < 1000:  # Zu wenige Farben = AI
        ai_png_indicator = True
    elif unique_colors > 100000:  # Künstlich viele Farben = AI
        ai_png_indicator = True
    elif dithering_score > 50:  # Starkes Dithering = AI-Artefakt
        ai_png_indicator = True
    elif edge_density < 0.01:  # Zu wenige Kanten = überglättet = AI
        ai_png_indicator = True
    elif edge_density > 0.3:  # Zu viele Kanten = überschärft = AI
        ai_png_indicator = True

    return {
        'png_analysis': 'analyzed',
        'unique_colors': int(unique_colors),
        'dithering_score': float(dithering_score),
        'edge_density': float(edge_density),
        'ai_png_indicator': bool(ai_png_indicator),
        'color_complexity_category': (
            'very_low' if unique_colors < 1000 else
            'low' if unique_colors < 5000 else
            'normal' if unique_colors < 50000 else
            'high' if unique_colors < 100000 else
            'artificial'
        )
    }

def calculate_artifact_ai_score(results):
    """Berechnet Gesamt-AI-Wahrscheinlichkeit basierend auf Artefakt-Analysen"""
    indicators = []

    # Sammle alle AI-Indikatoren
    if 'compression_artifacts' in results:
        indicators.append(results['compression_artifacts'].get('ai_blockiness_indicator', False))

    if 'ringing_artifacts' in results:
        indicators.append(results['ringing_artifacts'].get('ai_ringing_indicator', False))

    if 'color_bleeding' in results:
        indicators.append(results['color_bleeding'].get('ai_bleeding_indicator', False))

    if 'upscaling_artifacts' in results:
        indicators.append(results['upscaling_artifacts'].get('ai_upscaling_indicator', False))

    if 'noise_reduction' in results:
        indicators.append(results['noise_reduction'].get('ai_denoise_indicator', False))

    # PNG-spezifische Indikatoren
    if 'png_artifacts' in results and results['png_artifacts'].get('png_analysis') != 'not_applicable':
        indicators.append(results['png_artifacts'].get('ai_png_indicator', False))

    # Berechne Gesamt-Score
    ai_score = sum(indicators) / len(indicators) if indicators else 0.0
    confidence = len(indicators) / 6.0  # Max 6 Indikatoren (inklusive PNG)

    return {
        'ai_probability_score': float(ai_score),
        'confidence_level': float(confidence),
        'total_indicators': len(indicators),
        'positive_indicators': sum(indicators),
        'artifact_summary': {
            'has_compression_artifacts': bool(results.get('compression_artifacts', {}).get('ai_blockiness_indicator', False)),
            'has_ringing': bool(results.get('ringing_artifacts', {}).get('ai_ringing_indicator', False)),
            'has_color_bleeding': bool(results.get('color_bleeding', {}).get('ai_bleeding_indicator', False)),
            'has_upscaling': bool(results.get('upscaling_artifacts', {}).get('ai_upscaling_indicator', False)),
            'over_denoised': bool(results.get('noise_reduction', {}).get('ai_denoise_indicator', False)),
            'png_specific_issues': bool(results.get('png_artifacts', {}).get('ai_png_indicator', False))
        }
    }

def detect_artifacts(img_path):
    """Hauptfunktion für Artefakt-Erkennung"""
    try:
        img = cv2.imread(img_path, cv2.IMREAD_UNCHANGED)
        if img is None:
            return {'error': f'Konnte Bild nicht laden: {img_path}'}

        # Basis-Analysen (für beide Formate)
        results = {
            'compression_artifacts': detect_compression_artifacts(img),
            'ringing_artifacts': detect_ringing_artifacts(img),
            'color_bleeding': detect_color_bleeding(img),
            'upscaling_artifacts': detect_upscaling_artifacts(img),
            'noise_reduction': detect_noise_reduction_artifacts(img)
        }

        # Format-spezifische Analysen
        _, ext = os.path.splitext(img_path.lower())
        if ext == '.png':
            results['png_artifacts'] = detect_png_specific_artifacts(img, img_path)

        # Gesamt-AI-Score berechnen
        results['overall_assessment'] = calculate_artifact_ai_score(results)

        # JSON-serializable machen
        results = make_json_serializable(results)

        return results

    except Exception as e:
        return {'error': str(e)}

def main():
    if len(sys.argv) != 2:
        print("Usage: python3 detect-artifacts.py <image_path>", file=sys.stderr)
        sys.exit(1)

    img_path = sys.argv[1]
    result = detect_artifacts(img_path)
    print(json.dumps(result, indent=2))

if __name__ == "__main__":
    main()