import cv2
import numpy as np
import json
import sys

def analyze_object_coherence(image_path):
    """Physik-basierte Kohärenzanalyse ohne AI"""

    try:
        img = cv2.imread(image_path)
        if img is None:
            return {"error": "Could not load image"}

        # Einfache Computer Vision Methoden
        perspective_score = check_perspective_consistency(img)
        lighting_coherence = check_lighting_physics(img)
        edge_consistency = analyze_edge_patterns(img)

        # Kombinierter Score
        coherence_score = (perspective_score + lighting_coherence + edge_consistency) / 3

        analysis = {
            "object_analysis": {
                "perspective_consistency": perspective_score,
                "lighting_coherence": lighting_coherence,
                "edge_consistency": edge_consistency,
                "ai_coherence_score": coherence_score,
                "anomalies": []
            }
        }

        return analysis

    except Exception as e:
        return {"error": str(e)}

def check_perspective_consistency(img):
    """Prüft Fluchtpunkt-Konsistenz"""
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    edges = cv2.Canny(gray, 50, 150)
    lines = cv2.HoughLinesP(edges, 1, np.pi/180, threshold=100, minLineLength=50, maxLineGap=10)

    if lines is None:
        return 0.1  # Wenig Linien = verdächtig

    # Vereinfachte Perspektiv-Analyse
    line_count = len(lines)
    if line_count < 10:
        return 0.9
    elif line_count > 200:
        return 0.7  # Zu chaotisch
    else:
        return 0.3  # Normal

def check_lighting_physics(img):
    """Prüft Beleuchtungs-Physik"""
    lab = cv2.cvtColor(img, cv2.COLOR_BGR2LAB)
    l_channel = lab[:,:,0]

    # Gradient-Analyse
    grad_x = cv2.Sobel(l_channel, cv2.CV_64F, 1, 0, ksize=3)
    grad_y = cv2.Sobel(l_channel, cv2.CV_64F, 0, 1, ksize=3)

    gradient_magnitude = np.sqrt(grad_x**2 + grad_y**2)
    gradient_std = np.std(gradient_magnitude)

    # AI-generierte Bilder haben oft zu gleichmäßige Beleuchtung
    if gradient_std < 20:
        return 0.6  # Zu gleichmäßig = verdächtig
    elif gradient_std > 80:
        return 0.2  # Sehr ungleichmäßig = natürlich
    else:
        return 0.3  # Normal

def analyze_edge_patterns(img):
    """Analysiert Edge-Muster ohne AI"""
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)

    # Verschiedene Edge-Detektoren
    canny = cv2.Canny(gray, 50, 150)
    laplacian = cv2.Laplacian(gray, cv2.CV_64F)

    # Edge-Dichte
    edge_density = np.mean(canny > 0)
    laplacian_var = np.var(laplacian)

    # AI-Bilder haben oft zu saubere oder zu chaotische Edges
    if edge_density < 0.05 or edge_density > 0.3:
        return 0.5  # Unnatürliche Edge-Dichte

    if laplacian_var < 100:
        return 0.4  # Zu glatt

    return 0.1  # Normal

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python analyze_coherence.py <image_path>")
        sys.exit(1)

    result = analyze_object_coherence(sys.argv[1])
    print(json.dumps(result, indent=2))