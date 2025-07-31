import sys
import json
import numpy as np
import jpegio as jio

from quant_tables import extract_quant_tables


def dct_histogram(path, bins=100, rang=(-100,100)):
    jpg = jio.read(path)
    # alle Komponenten-Koeffizienten zusammenfassen
    coeffs = np.hstack([c.ravel() for c in jpg.coef_arrays])
    hist, edges = np.histogram(coeffs, bins=bins, range=rang)
    return {
        "histogram": hist.tolist(),
        "bin_edges": edges.tolist()
    }

if __name__ == "__main__":
    img_path = sys.argv[1]
    out = extract_quant_tables(img_path)
    hist = dct_histogram(img_path)
    out.update(hist)
    print(json.dumps(out))
