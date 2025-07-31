import cv2
import numpy as np
import json
import sys
import os
from scipy.fft import fft2
from skimage import feature

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

def detect_ai_specific_patterns(img):
    """Erkennt AI-spezifische Muster wie typische GAN/Diffusion-Artefakte"""
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    # 1. Spektrale Anomalien - AI-Modelle hinterlassen charakteristische Frequenz-Signaturen
    fft = np.abs(fft2(gray))
    fft_log = np.log(fft + 1)

    # Suche nach unnatürlichen Frequenzspitzen
    center = np.array(fft.shape) // 2
    y, x = np.ogrid[:fft.shape[0], :fft.shape[1]]
    mask = (x - center[1])**2 + (y - center[0])**2 <= (min(fft.shape)//4)**2

    high_freq_content = np.mean(fft_log[~mask])
    low_freq_content = np.mean(fft_log[mask])
    freq_balance = high_freq_content / (low_freq_content + 1e-8)

    # 2. Texture Coherence Analysis - AI hat oft zu kohärente Texturen
    patch_size = 32
    coherence_scores = []

    for i in range(0, gray.shape[0] - patch_size, patch_size):
        for j in range(0, gray.shape[1] - patch_size, patch_size):
            patch = gray[i:i+patch_size, j:j+patch_size]
            # Local Binary Pattern für Texturanalyse
            lbp = feature.local_binary_pattern(patch, 8, 1, method='uniform')
            coherence = np.std(lbp)
            coherence_scores.append(coherence)

    texture_consistency = np.std(coherence_scores)

    # 3. Luminance Gradient Analysis - AI hat oft unnatürliche Helligkeitsverteilungen
    grad_x = cv2.Sobel(gray, cv2.CV_64F, 1, 0, ksize=3)
    grad_y = cv2.Sobel(gray, cv2.CV_64F, 0, 1, ksize=3)
    gradient_magnitude = np.sqrt(grad_x**2 + grad_y**2)
    gradient_uniformity = np.std(gradient_magnitude)

    return {
        'frequency_balance': float(freq_balance),
        'texture_consistency': float(texture_consistency),
        'gradient_uniformity': float(gradient_uniformity),
        'ai_pattern_indicator': bool(freq_balance < 0.5 or texture_consistency > 20 or gradient_uniformity < 10),
        'spectral_anomaly_score': float(freq_balance),
        'texture_anomaly_score': float(texture_consistency),
        'gradient_anomaly_score': float(gradient_uniformity)
    }

def detect_compression_timeline(img):
    """Analysiert Kompressionsgeschichte - mehrfache Kompression deutet auf Bearbeitung hin"""
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    # Simuliere verschiedene Kompressionsgrade
    compression_artifacts = []

    for quality in [95, 85, 75, 65, 55]:
        # Simuliere JPEG-Kompression
        encode_param = [int(cv2.IMWRITE_JPEG_QUALITY), quality]
        _, encimg = cv2.imencode('.jpg', gray, encode_param)
        decoded = cv2.imdecode(encimg, cv2.IMREAD_GRAYSCALE)

        # Berechne Mean Squared Error
        mse = np.mean((gray.astype(float) - decoded.astype(float))**2)
        compression_artifacts.append(mse)

    # Analysiere Kompressionsresistenz
    compression_gradient = np.diff(compression_artifacts)
    resistance_score = np.mean(compression_gradient)

    # Berechne Kompressionslinearität
    compression_linearity = np.corrcoef(range(len(compression_artifacts)), compression_artifacts)[0,1]

    return {
        'compression_resistance': float(resistance_score),
        'compression_linearity': float(compression_linearity),
        'likely_recompressed': bool(resistance_score < -10),
        'compression_timeline_score': float(resistance_score),
        'compression_artifacts_progression': [float(x) for x in compression_artifacts]
    }

def detect_synthetic_noise(img):
    """Erkennt synthetisches Rauschen typisch für AI-generierte Bilder"""
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    # 1. Rausch-Autokorrelation
    # Echtes Rauschen hat spezifische Autokorrelationseigenschaften
    noise = gray - cv2.GaussianBlur(gray, (5,5), 1.0)

    # Berechne Autokorrelation
    autocorr = cv2.matchTemplate(noise, noise[10:30, 10:30], cv2.TM_CCOEFF_NORMED)
    autocorr_peak = np.max(autocorr)

    # 2. Spektrale Eigenschaften des Rauschens
    noise_fft = np.abs(fft2(noise))
    noise_spectrum_flatness = np.std(noise_fft) / (np.mean(noise_fft) + 1e-8)

    # 3. Lokale Rauschvariation
    patch_size = 16
    noise_variations = []

    for i in range(0, gray.shape[0] - patch_size, patch_size):
        for j in range(0, gray.shape[1] - patch_size, patch_size):
            patch_noise = noise[i:i+patch_size, j:j+patch_size]
            noise_std = np.std(patch_noise)
            noise_variations.append(noise_std)

    noise_uniformity = np.std(noise_variations)

    return {
        'autocorr_peak': float(autocorr_peak),
        'spectrum_flatness': float(noise_spectrum_flatness),
        'noise_uniformity': float(noise_uniformity),
        'synthetic_noise_indicator': bool(autocorr_peak > 0.8 or noise_uniformity < 1.0),
        'noise_analysis_score': float((autocorr_peak + (1.0 - noise_uniformity/5.0)) / 2.0)
    }

def analyze_advanced_artifacts(img_path):
    """Hauptfunktion für erweiterte Artefakt-Analyse"""
    try:
        img = cv2.imread(img_path, cv2.IMREAD_UNCHANGED)
        if img is None:
            return {'error': f'Konnte Bild nicht laden: {img_path}'}

        # Führe alle erweiterten Analysen durch
        results = {
            'ai_patterns': detect_ai_specific_patterns(img),
            'compression_timeline': detect_compression_timeline(img),
            'synthetic_noise': detect_synthetic_noise(img)
        }

        # Berechne erweiterten AI-Score
        results['advanced_assessment'] = calculate_advanced_ai_score(results)

        # JSON-serializable machen
        results = make_json_serializable(results)

        return results

    except Exception as e:
        return {'error': str(e)}

def calculate_advanced_ai_score(results):
    """Berechnet AI-Score basierend auf erweiterten Analysen"""
    indicators = []

    # AI-Pattern Indikatoren
    if 'ai_patterns' in results:
        indicators.append(results['ai_patterns'].get('ai_pattern_indicator', False))

    # Compression Timeline Indikatoren
    if 'compression_timeline' in results:
        indicators.append(results['compression_timeline'].get('likely_recompressed', False))

    # Synthetic Noise Indikatoren
    if 'synthetic_noise' in results:
        indicators.append(results['synthetic_noise'].get('synthetic_noise_indicator', False))

    # Berechne Score
    ai_score = sum(indicators) / len(indicators) if indicators else 0.0
    confidence = len(indicators) / 3.0  # Max 3 erweiterte Indikatoren

    return {
        'advanced_ai_probability': float(ai_score),
        'advanced_confidence': float(confidence),
        'advanced_indicators': {
            'unnatural_patterns': bool(results.get('ai_patterns', {}).get('ai_pattern_indicator', False)),
            'compression_history': bool(results.get('compression_timeline', {}).get('likely_recompressed', False)),
            'synthetic_noise': bool(results.get('synthetic_noise', {}).get('synthetic_noise_indicator', False))
        }
    }

def main():
    if len(sys.argv) != 2:
        print("Usage: python3 advanced-artifacts.py <image_path>", file=sys.stderr)
        sys.exit(1)

    img_path = sys.argv[1]
    result = analyze_advanced_artifacts(img_path)
    print(json.dumps(result, indent=2))

if __name__ == "__main__":
    main()