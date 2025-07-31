# analyze_compression.py
import json
import sys
from PIL import Image
import numpy as np

def extract_quant_tables(path):
    im = Image.open(path)
    if not hasattr(im, "quantization") or im.quantization is None:
        return {}
    qtables = im.quantization  # dict: Komponenten → 8×8-Matrix
    stats = {}
    for comp, table in qtables.items():
        arr = np.array(table).reshape(8,8)
        stats[f"comp{comp}_mean"] = float(arr.mean())
        stats[f"comp{comp}_std"]  = float(arr.std())
    return stats

if __name__ == "__main__":
    img_path = sys.argv[1]
    result = {
        "quant_tables": extract_quant_tables(img_path)
    }
    print(json.dumps(result))

