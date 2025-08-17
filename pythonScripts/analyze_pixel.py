import cv2
import numpy as np
from scipy import stats
from scipy.fft import fft2, fftshift
import json
import sys

def analyze_noise_patterns(img):
    """Analysiert Rauschcharakteristiken - KI-Bilder haben oft unnatürliche Rauschmuster"""
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    # Lokale Standardabweichung berechnen
    kernel = np.ones((5,5), np.float32) / 25
    mean_img = cv2.filter2D(gray.astype(np.float32), -1, kernel)
    noise = gray.astype(np.float32) - mean_img

    noise_std = np.std(noise)
    noise_skewness = stats.skew(noise.flatten())

    # KI-Bilder haben oft zu gleichmäßiges Rauschen
    uniformity_score = np.std([np.std(noise[i:i+50, j:j+50])
                              for i in range(0, gray.shape[0]-50, 50)
                              for j in range(0, gray.shape[1]-50, 50)])

    return {
        'noise_std': float(noise_std),
        'noise_skewness': float(noise_skewness),
        'uniformity_score': float(uniformity_score),
        'ai_noise_indicator': uniformity_score < 10.0
    }

def analyze_frequency_domain(img):
    """Frequenzbereichsanalyse - unnatürliche Frequenzmuster erkennen"""
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    fft = fft2(gray)
    fft_shifted = fftshift(fft)
    magnitude_spectrum = np.log(np.abs(fft_shifted) + 1)

    center = (magnitude_spectrum.shape[0] // 2, magnitude_spectrum.shape[1] // 2)
    y, x = np.ogrid[:magnitude_spectrum.shape[0], :magnitude_spectrum.shape[1]]
    radius = np.sqrt((x - center[1])**2 + (y - center[0])**2)

    low_freq = magnitude_spectrum[radius < 20].mean()
    high_freq = magnitude_spectrum[radius >= 100].mean()
    freq_ratio = high_freq / (low_freq + 1e-8)

    return {
        'low_freq_energy': float(low_freq),
        'high_freq_energy': float(high_freq),
        'freq_ratio': float(freq_ratio),
        'ai_freq_indicator': freq_ratio < 0.7
    }

def check_color_consistency(img):
    """Prüft Farbkonsistenz - KORRIGIERT für echte Fotos"""
    hsv = cv2.cvtColor(img, cv2.COLOR_BGR2HSV)

    saturation_mean = np.mean(hsv[:,:,1])
    saturation_std = np.std(hsv[:,:,1])
    hue_std = np.std(hsv[:,:,0])

    # KORRIGIERTE Schwellenwerte für echte Fotos:
    # Echte Fotos haben oft moderate Sättigung (50-200) und natürliche Varianz (>30)

    # AI-Indikatoren (weniger aggressiv):
    oversaturation_indicator = saturation_mean > 200  # ↑ von 150 - weniger aggressiv
    undersaturation_indicator = saturation_mean < 30  # Neue Erkennung für zu blasse AI-Bilder
    uniformity_indicator = saturation_std < 15        # ↓ von 20 - weniger aggressiv
    unnatural_hue_indicator = hue_std < 10           # Neue Erkennung für unnatürliche Farbverteilung

    # Kombinierte AI-Wahrscheinlichkeit (konservativer)
    ai_indicators = [oversaturation_indicator, undersaturation_indicator,
                    uniformity_indicator, unnatural_hue_indicator]
    ai_color_score = sum(ai_indicators) / len(ai_indicators)

    # BONUS: Reduziere Score für natürliche Farbverteilungen
    if 50 <= saturation_mean <= 180 and saturation_std >= 25:
        ai_color_score *= 0.6  # -40% für natürliche Farbverteilung

    return {
        'saturation_mean': float(saturation_mean),
        'saturation_std': float(saturation_std),
        'hue_std': float(hue_std),
        'oversaturation_indicator': oversaturation_indicator,
        'undersaturation_indicator': undersaturation_indicator,
        'color_uniformity_indicator': uniformity_indicator,
        'unnatural_hue_indicator': unnatural_hue_indicator,
        'ai_color_score': float(ai_color_score),
        'color_naturalness_score': float(1.0 - ai_color_score)  # Inverse für Authentizität
    }

def calculate_ai_pixel_score(results):
    """Berechnet Gesamt-KI-Wahrscheinlichkeit aus Pixel-Analysen"""
    indicators = []

    if 'noise_analysis' in results:
        indicators.append(results['noise_analysis'].get('ai_noise_indicator', False))
    if 'frequency_analysis' in results:
        indicators.append(results['frequency_analysis'].get('ai_freq_indicator', False))
    if 'color_analysis' in results:
        color = results['color_analysis']
        indicators.append(color.get('oversaturation_indicator', False))
        indicators.append(color.get('color_uniformity_indicator', False))

    ai_score = sum(indicators) / len(indicators) if indicators else 0.0

    return {
        'ai_probability_score': ai_score,
        'positive_indicators': sum(indicators),
        'total_indicators': len(indicators)
    }

def analyze_pixel_patterns(img_path):
    """Hauptfunktion für komplette Pixel-Level-Analyse"""
    try:
        img = cv2.imread(img_path)
        if img is None:
            return {'error': f'Konnte Bild nicht laden: {img_path}'}

        results = {
            'noise_analysis': analyze_noise_patterns(img),
            'frequency_analysis': analyze_frequency_domain(img),
            'color_analysis': check_color_consistency(img)
        }

        results['overall_assessment'] = calculate_ai_pixel_score(results)
        return results

    except Exception as e:
        return {'error': str(e)}

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

def main():
    if len(sys.argv) != 2:
        print("Usage: python3 analyze_pixel.py <image_path>", file=sys.stderr)
        sys.exit(1)

    result = analyze_pixel_patterns(sys.argv[1])
    # JSON-serializable machen
    result = make_json_serializable(result)
    print(json.dumps(result, indent=2))

if __name__ == "__main__":
    main()