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
    ai_indicators = []

    # 1. Double compression - KOMPLETT ÜBERARBEITET
    double_comp = result.get('double_compression_score', 0.0)
    if double_comp > 0.025:  # War 0.008 - jetzt realistisch
        ai_indicators.append(True)
    else:
        ai_indicators.append(False)

    # 2. Quantization - REALISTISCH für echte Fotos
    quant_tables = result.get('quant_tables', {})
    comp0_mean = 0
    comp0_std = 0

    if quant_tables:
        comp0_mean = quant_tables.get('comp0_mean', 0)
        comp0_std = quant_tables.get('comp0_std', 0)

        # Echte Fotos haben oft NIEDRIGE Quantization (hohe Qualität)
        # AI hat oft HOHE Quantization (mittlere Qualität)
        if comp0_mean > 60:  # War 50 - jetzt für extreme AI-Quantization
            ai_indicators.append(True)
        else:
            ai_indicators.append(False)

        # High variation in quantization = AI processing
        if comp0_std > 50:  # War 40 - jetzt strenger
            ai_indicators.append(True)
        else:
            ai_indicators.append(False)

    # 3. DCT Histogram - KORRIGIERT für echte Fotos
    histogram = result.get('dct_histogram', [])
    if histogram and len(histogram) > 50:
        hist_array = np.array(histogram)
        max_val = np.max(hist_array)
        max_idx = np.argmax(hist_array)

        # KORRIGIERT: Echte Fotos können massive Spikes haben!
        # Nur sehr spezifische AI-Muster sind verdächtig
        if max_val > 8000000 and 48 <= max_idx <= 52:  # War 500000 - jetzt extrem
            ai_indicators.append(True)
        else:
            ai_indicators.append(False)

        # Additional: Check for unnatural concentration - KORRIGIERT
        total_values = np.sum(hist_array)
        if max_val / total_values > 0.9:  # War 0.8 - jetzt strenger für AI
            ai_indicators.append(True)
        else:
            ai_indicators.append(False)

    # 4. NEUER Test: Quantization-Histogram Kombination
    if comp0_mean > 60 and len(histogram) > 0 and np.max(histogram) > 1000000:  # AI-typische Kombination
        ai_indicators.append(True)
    else:
        ai_indicators.append(False)

    # 5. DATASET Pattern matching - KORRIGIERT
    # Basierend auf deinen Daten: AI-Bilder typischerweise 0.008-0.02
    if 0.008 <= double_comp <= 0.02 and comp0_mean > 50:
        ai_indicators.append(True)  # Diese Kombination ist AI in deinen Daten
    else:
        ai_indicators.append(False)

    # FINALE BERECHNUNG - ERSETZT calculate_final_score
    ai_probability = sum(ai_indicators) / len(ai_indicators) if ai_indicators else 0.0

    # BOOST: Wenn multiple Indikatoren zustimmen, erhöhe Confidence
    positive_count = sum(ai_indicators)
    if positive_count >= 4:
        ai_probability = min(ai_probability * 1.5, 1.0)  # +50% boost
    elif positive_count >= 3:
        ai_probability = min(ai_probability * 1.3, 1.0)  # +30% boost
    elif positive_count >= 2:
        ai_probability = min(ai_probability * 1.1, 1.0)  # +10% boost

    # PENALTY: Wenn es echte Foto-Charakteristika hat, reduziere Score
    if comp0_mean < 20 and max_val > 3000000:  # Hohe Qualität + natürlicher Spike
        ai_probability *= 0.7  # -30% penalty für echte Foto-Merkmale

    return {
        'ai_probability': float(ai_probability),
        'indicators': {
            'high_double_compression': bool(double_comp > 0.025),
            'suspicious_quantization': bool(comp0_mean > 60),
            'high_variation_quantization': bool(comp0_std > 50),
            'unnatural_histogram': bool(len(histogram) > 50 and np.max(histogram) > 8000000),
            'histogram_concentration': bool(len(histogram) > 50 and np.max(histogram) / np.sum(histogram) > 0.9),
            'dataset_pattern_match': bool(0.008 <= double_comp <= 0.02 and comp0_mean > 50),
            'ai_quantization_combo': bool(comp0_mean > 60 and len(histogram) > 0 and np.max(histogram) > 1000000)
        },
        'confidence': float(min(len(ai_indicators) / 6.0, 1.0)),  # 6 Indikatoren total
        'details': {
            'double_compression_value': float(double_comp),
            'quantization_mean': float(comp0_mean),
            'quantization_std': float(comp0_std),
            'histogram_max_value': float(np.max(histogram)) if histogram else 0,
            'histogram_max_index': int(np.argmax(histogram)) if histogram else -1,
            'positive_indicators': int(sum(ai_indicators)),
            'total_indicators': len(ai_indicators)
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