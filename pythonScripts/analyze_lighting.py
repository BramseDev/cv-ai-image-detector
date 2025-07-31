import cv2
import numpy as np
import json
import sys

def analyze_lighting_physics(image_path):
    """Analysiert Lichtverhältnisse und physikalische Konsistenz"""

    img = cv2.imread(image_path)
    hsv = cv2.cvtColor(img, cv2.COLOR_BGR2HSV)

    analysis = {
        "lighting_analysis": {
            "light_source_consistency": 0.0,
            "shadow_direction_consistency": 0.0,
            "exposure_uniformity": 0.0,
            "ai_lighting_score": 0.0,
            "anomalies": []
        }
    }

    # 1. Helligkeit-Verteilung analysieren
    brightness = hsv[:,:,2]

    # Erkenne sehr helle Bereiche (Lichtquellen)
    bright_mask = brightness > 200
    bright_regions = cv2.connectedComponents(bright_mask.astype(np.uint8))[0]

    # 2. Schatten-Richtung analysieren
    shadow_consistency = analyze_shadow_directions(img)
    analysis["lighting_analysis"]["shadow_direction_consistency"] = shadow_consistency

    # 3. Exposure-Anomalien
    exposure_score = check_exposure_anomalies(brightness)
    analysis["lighting_analysis"]["exposure_uniformity"] = exposure_score

    # 4. Lichtquelle-Konsistenz
    light_consistency = check_light_source_consistency(bright_regions, brightness)
    analysis["lighting_analysis"]["light_source_consistency"] = light_consistency

    # Gesamtscore
    ai_indicators = 0
    total_checks = 3

    if shadow_consistency < 0.8:  # Inkonsistente Schatten
        ai_indicators += 1
        analysis["lighting_analysis"]["anomalies"].append("Inconsistent shadow directions")

    if exposure_score < 0.7:  # Unnatürliche Belichtung
        ai_indicators += 1
        analysis["lighting_analysis"]["anomalies"].append("Unnatural exposure patterns")

    if light_consistency < 0.7:  # Inkonsistente Lichtquellen
        ai_indicators += 1
        analysis["lighting_analysis"]["anomalies"].append("Inconsistent light sources")

    analysis["lighting_analysis"]["ai_lighting_score"] = ai_indicators / total_checks

    return analysis

def analyze_shadow_directions(img):
    """Analysiert Schatten-Richtungen für Konsistenz"""
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    # Gradient-basierte Kantenerkennung
    grad_x = cv2.Sobel(gray, cv2.CV_64F, 1, 0, ksize=3)
    grad_y = cv2.Sobel(gray, cv2.CV_64F, 0, 1, ksize=3)

    # Berechne dominante Gradient-Richtung
    angles = np.arctan2(grad_y, grad_x)

    # Histogramm der Winkel
    hist, _ = np.histogram(angles, bins=36, range=(-np.pi, np.pi))

    # Konsistenz = wie konzentriert sind die Schatten-Richtungen?
    max_bin = np.max(hist)
    total = np.sum(hist)

    consistency = max_bin / total if total > 0 else 0
    return min(consistency * 2, 1.0)  # Normalisiere

def check_exposure_anomalies(brightness):
    """Prüft auf unnatürliche Belichtungsverteilung"""

    # Berechne Helligkeits-Histogramm
    hist = cv2.calcHist([brightness], [0], None, [256], [0, 256])

    # Normalisiere
    hist = hist.flatten() / hist.sum()

    # Natürliche Bilder haben meist eine Gauss-ähnliche Verteilung
    # AI-Bilder oft zu gleichmäßig oder zu extreme Verteilungen

    # Berechne "Natürlichkeit" der Verteilung
    mean_brightness = np.mean(brightness)
    std_brightness = np.std(brightness)

    # Sehr niedrige STD = zu gleichmäßig (AI-typisch)
    # Sehr hohe STD = zu kontrastreich (AI-typisch)

    if std_brightness < 20:  # Zu gleichmäßig
        return 0.2
    elif std_brightness > 80:  # Zu kontrastreich
        return 0.3
    else:
        return 0.8  # Normal

def check_light_source_consistency(bright_regions, brightness):
    """Prüft Konsistenz der Lichtquellen"""

    if bright_regions < 2:
        return 0.7  # Eine Lichtquelle ist normal

    # Mehrere Lichtquellen sollten ähnliche Intensität haben
    # (vereinfachte Annahme)

    bright_pixels = brightness[brightness > 200]
    if len(bright_pixels) == 0:
        return 0.5

    brightness_std = np.std(bright_pixels)

    # Niedrige STD = konsistente Lichtquellen
    # Hohe STD = inkonsistente Lichtquellen (AI-verdächtig)

    consistency = max(0, 1.0 - (brightness_std / 50))
    return consistency

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python lighting_analysis.py <image_path>")
        sys.exit(1)

    result = analyze_lighting_physics(sys.argv[1])
    print(json.dumps(result, indent=2))