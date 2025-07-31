#!/usr/bin/env python3
"""
PNG-Analyzer: Extrahiert Kompressions-Metriken für PNG-Dateien:
- filesize vs. roher Bilddaten-Größe
- Shannon-Entropie der Pixel
- IDAT-Chunk-Größe

Usage: python3 png_analyzer.py <bild.png>
"""
import sys
import os
import json
import struct
import numpy as np
from PIL import Image

def png_stats(path):
    # 1) Dateigröße und Raw-Daten-Größe
    filesize = os.path.getsize(path)
    arr = np.array(Image.open(path))
    raw_size = arr.nbytes
    compression_ratio = filesize / raw_size if raw_size else None

    # 2) Shannon-Entropie der Pixel
    flat = arr.ravel()
    freq = np.bincount(flat, minlength=256) / flat.size
    entropy = float(-(freq[freq > 0] * np.log2(freq[freq > 0])).sum())

    # 3) Größe aller IDAT-Chunks
    idat_total = 0
    with open(path, 'rb') as f:
        data = f.read()
    # PNG-Signatur (8 Bytes) überspringen
    idx = 8
    while idx < len(data):
        length = struct.unpack('>I', data[idx:idx+4])[0]
        ctype = data[idx+4:idx+8]
        if ctype == b'IDAT':
            idat_total += length
        idx += 12 + length

    return {
        "filesize": filesize,
        "raw_size": raw_size,
        "compression_ratio": compression_ratio,
        "pixel_entropy": entropy,
        "idat_bytes": idat_total
    }

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage: python3 png_analyzer.py <image.png>")
        sys.exit(1)
    stats = png_stats(sys.argv[1])
    print(json.dumps(stats, indent=2))




# ---

# ### 1. `compression_ratio`
# > **Was es ist:** Verhältnis Dateigröße / roher Pixel-Daten.
# > **Bedeutung für dich:**
# > - **Niedrig** (z. B. < 0.1): Bild komprimiert sich sehr gut verlustfrei → oft **Diagramm**, **Screenshot**, **Icon** oder KI-generierte Grafik mit wenigen Farben und großen, einheitlichen Flächen.
# > - **Mittel** (0.1–0.3): typischer **Fotoinhalt** mit moderatem Detailgrad → könnte direkt vom **Smartphone** stammen.
# > - **Hoch** (> 0.3): Bild lässt sich kaum komprimieren → sehr detailreich oder häufige Farbübergänge → ebenfalls Foto, evtl. mehrfach bearbeitet (achte zusätzlich auf `pixel_entropy`).

# ### 2. `pixel_entropy`
# > **Was es ist:** Shannon-Entropie aller Pixelwerte (0–8).
# > **Bedeutung für dich:**
# > - **Niedrig** (< 5): wenige Farben, große Flächen, stark quantisierte Verläufe → oft **Diagramme**, **Screenshots**, **Vector-Art** (häufig KI-Output).
# > - **Hoch** (> 7): komplexe Farbinhalte, Foto-typisches Rauschen, viele Details → **echtes Foto** (Smartphone, Kamera) oder aufwändig in Photoshop bearbeitet.

# ### 3. `idat_bytes` vs. `filesize`
# > **Was es ist:** Anteil der reinen Bilddaten im PNG-Container.
# > **Bedeutung für dich:**
# > - **Nahezu gleich**: Container enthält kaum zusätzliche Metadaten oder versteckte Layer → direktes Export-PNG.
# > - **Deutlicher Unterschied**: legt nahe, dass Text-Chunks, Transparenzmasken oder andere nicht-bildrelevante Daten eingelagert sind (z. B. Photoshop-Projekte, C2PA-Claims, Exif-Text) → nachträgliche Bearbeitung.

# ---

# ## Praktische Flags

# - **KI-Erzeugnis (PNG)**
#   - `compression_ratio < 0.1` **und** `pixel_entropy < 5`
#   - + evtl. ungewöhnliche Metadaten in EXIF/Text-Chunks

# - **Smartphone-Foto**
#   - `0.1 < compression_ratio < 0.3` **und** `pixel_entropy > 7`
#   - meist einfache IDAT-Struktur, keine großen Text-Chunks

# - **Bearbeitetes/Resaved PNG**
#   - `compression_ratio` und `pixel_entropy` wie Foto, aber
#   - `filesize – idat_bytes` relativ groß → viel zusätzlicher Container-Overhead

# Feature | Dein Wert | Typisches KI-PNG (Illustration)
# compression_ratio | 0.5376 | oft < 0.1 (flache Farbflächen)
# pixel_entropy | 7.53 | typischerweise < 5 bei KI-Artworks
# idat_bytes/filesize | ≈0.96 | sehr hoch, kaum Metadaten