import sys
import os
import json
import subprocess
import numpy as np

# Basispfad zum compression-Ordner
this_dir = os.path.dirname(__file__)
base_dir = os.path.abspath(os.path.join(this_dir, '../pkg/analyzer/compression'))

# Unterordner definieren
jpeg_dir = os.path.join(base_dir, 'jpeg')
png_dir = os.path.join(base_dir, 'png')
script_dir = this_dir  # Für Skripte im selben Ordner

def run_module(dir_path, script_name, img_path):
    script_path = os.path.join(dir_path, script_name)
    if not os.path.isfile(script_path):
        print(f"[Fehler] Modul-Skript nicht gefunden: {script_path}", file=sys.stderr)
        return {}

    proc = subprocess.run(
        [sys.executable, script_path, img_path],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True
    )

    if proc.returncode != 0:
        print(f"[Fehler] Fehler bei {script_name} auf {img_path}:\n{proc.stderr}", file=sys.stderr)
        return {}

    try:
        return json.loads(proc.stdout)
    except json.JSONDecodeError as e:
        print(f"[Fehler] Ungültiges JSON von {script_name}: {e}", file=sys.stderr)
        return {}

def calculate_jpeg_ai_score(result):
    """Berechnet AI-Wahrscheinlichkeit basierend auf JPEG-Kompressionsmerkmalen"""
    ai_indicators = []

    # 1. Double Compression Score
    double_comp = result.get('double_compression_score', 0.0)
    if double_comp > 0.15:  # Starke Doppelkompression = verdächtig
        ai_indicators.append(True)
    else:
        ai_indicators.append(False)

    # 2. Quantization Tables Analysis
    quant_tables = result.get('quant_tables', {})
    comp0_mean = 0
    comp0_std = 0

    if quant_tables:
        # KI-Bilder haben oft unnatürliche Quantisierungsparameter
        comp0_mean = quant_tables.get('comp0_mean', 0)
        comp0_std = quant_tables.get('comp0_std', 0)

        # Zu niedrige Werte = zu hohe Qualität = verdächtig
        if comp0_mean < 15:
            ai_indicators.append(True)
        else:
            ai_indicators.append(False)

        # Zu gleichmäßige Quantisierung = verdächtig
        if comp0_std < 5:
            ai_indicators.append(True)
        else:
            ai_indicators.append(False)

    # 3. DCT Histogram Analysis
    histogram = result.get('dct_histogram', [])
    hist_skewness = 0

    if histogram and len(histogram) > 10:
        hist_array = np.array(histogram)

        # Unnatürliche Histogramm-Form
        hist_skewness = np.abs(np.mean(hist_array) - np.median(hist_array))
        if hist_skewness > 100:  # Zu asymmetrisch
            ai_indicators.append(True)
        else:
            ai_indicators.append(False)

    # Gesamt-Score berechnen
    ai_probability = sum(ai_indicators) / len(ai_indicators) if ai_indicators else 0.0

    return {
        'ai_probability': float(ai_probability),
        'indicators': {
            'high_double_compression': bool(double_comp > 0.15 if double_comp else False),
            'suspicious_quantization': bool(comp0_mean < 15 if quant_tables else False),
            'uniform_quantization': bool(comp0_std < 5 if quant_tables else False),
            'unnatural_histogram': bool(hist_skewness > 100 if histogram else False)
        },
        'confidence': float(len(ai_indicators) / 4.0),  # Max 4 Indikatoren
        'details': {
            'double_compression_value': float(double_comp),
            'quantization_mean': float(comp0_mean),
            'quantization_std': float(comp0_std),
            'histogram_skewness': float(hist_skewness)
        }
    }

def calculate_png_ai_score(result):
    """Berechnet AI-Wahrscheinlichkeit für PNG-Dateien - KORRIGIERT für echte Fotos"""
    ai_indicators = []

    # PNG-spezifische Analysen (abhängig von png_analyzer.py Output)
    suspicious_compression = False
    unusual_color_count = False

    # 1. Kompressionstyp-Analyse (KORRIGIERT)
    if 'compression_type' in result:
        comp_type = result.get('compression_type', '')
        # KORRIGIERT: "optimal" ist NORMAL für echte Fotos, nicht verdächtig
        if 'uncompressed' in comp_type.lower() or 'raw' in comp_type.lower():
            ai_indicators.append(True)  # Unkomprimiert = verdächtig für AI
            suspicious_compression = True
        else:
            ai_indicators.append(False)  # Normale Kompression = authentisch

    # 2. Farbanzahl-Analyse (KORRIGIERT für echte Fotos)
    if 'pixel_entropy' in result:
        entropy = result.get('pixel_entropy', 0)

        # KORRIGIERT: Echte Fotos haben hohe Entropie (>6), AI oft niedrige (<4)
        if entropy < 4.0:  # Zu wenig Komplexität = AI
            ai_indicators.append(True)
            unusual_color_count = True
        elif entropy > 7.5:  # Sehr hohe Komplexität = echtes Foto
            ai_indicators.append(False)
        else:
            ai_indicators.append(False)  # Normale Komplexität

    # 3. Compression Ratio Analyse (KORRIGIERT)
    if 'compression_ratio' in result:
        ratio = result.get('compression_ratio', 1.0)

        # KORRIGIERT: Sehr niedrige Ratio (0.0-0.1) = perfekte Kompression = AUTHENTISCH
        if ratio <= 0.1:
            ai_indicators.append(False)  # Perfekte Kompression = echtes Foto
        elif ratio < 0.3:
            ai_indicators.append(False)  # Gute Kompression = wahrscheinlich echt
        elif ratio > 0.8:
            ai_indicators.append(True)   # Schlechte Kompression = verdächtig
        else:
            ai_indicators.append(False)  # Normale Kompression

    # 4. IDAT-Anteil-Analyse (NEU)
    if 'idat_bytes' in result and 'filesize' in result:
        idat_bytes = result.get('idat_bytes', 0)
        filesize = result.get('filesize', 1)
        idat_ratio = idat_bytes / filesize if filesize > 0 else 0

        # Sehr hoher IDAT-Anteil (>0.98) = minimale Metadaten = oft AI
        if idat_ratio > 0.98:
            ai_indicators.append(True)
        else:
            ai_indicators.append(False)

    # Wenn keine spezifischen Indikatoren vorhanden, als authentisch bewerten
    if not ai_indicators:
        ai_indicators.append(False)  # Default: nicht verdächtig

    ai_probability = sum(ai_indicators) / len(ai_indicators) if ai_indicators else 0.0

    # BONUS: Reduziere Score für echte Foto-Charakteristika
    if 'pixel_entropy' in result and result.get('pixel_entropy', 0) > 7.0:
        ai_probability *= 0.7  # -30% für hohe Entropie (echtes Foto)

    if 'compression_ratio' in result and result.get('compression_ratio', 1.0) <= 0.05:
        ai_probability *= 0.5  # -50% für perfekte Kompression (echtes Foto)

    return {
        'ai_probability': float(ai_probability),
        'indicators': {
            'suspicious_compression': suspicious_compression,
            'unusual_color_count': unusual_color_count,
            'poor_compression_efficiency': bool('compression_ratio' in result and result.get('compression_ratio', 0) > 0.8),
            'minimal_metadata': bool('idat_bytes' in result and 'filesize' in result and
                                   (result.get('idat_bytes', 0) / result.get('filesize', 1)) > 0.98)
        },
        'confidence': float(len(ai_indicators) / 4.0),  # Max 4 Indikatoren
        'details': {
            'compression_type': result.get('compression_type', 'unknown'),
            'pixel_entropy': result.get('pixel_entropy', 0),
            'compression_ratio': result.get('compression_ratio', 0),
            'idat_ratio': result.get('idat_bytes', 0) / result.get('filesize', 1) if result.get('filesize', 0) > 0 else 0
        }
    }

def analyze_image(img_path):
    _, ext = os.path.splitext(img_path.lower())
    result = {}

    if ext in ('.jpg', '.jpeg'):
        result['quant_tables'] = run_module(jpeg_dir, 'quant_tables.py', img_path).get('quant_tables', {})
        result['dct_histogram'] = run_module(jpeg_dir, 'dct_histogramm.py', img_path).get('histogram', [])
        result['double_compression_score'] = run_module(jpeg_dir, 'double_compression.py', img_path).get('double_compression_score', 0.0)

        # AI-Score für JPEG hinzufügen
        result['compression_ai_analysis'] = calculate_jpeg_ai_score(result)

    elif ext == '.png':
        result = run_module(png_dir, 'png_analyzer.py', img_path)

        # AI-Score für PNG hinzufügen
        result['compression_ai_analysis'] = calculate_png_ai_score(result)

    else:
        result['error'] = f"Format '{ext}' nicht unterstützt"

    return result

def main():
    if len(sys.argv) < 2:
        print("Usage: python3 scripts/analyze_compression.py <image1> [<image2> ...]", file=sys.stderr)
        sys.exit(1)

    all_results = {}
    for img in sys.argv[1:]:
        key = os.path.basename(img)
        all_results[key] = analyze_image(img)

    print(json.dumps(all_results, indent=2))

if __name__ == "__main__":
    main()