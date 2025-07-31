import sys
import json
import numpy as np
from scipy.signal import correlate

from quant_tables import extract_quant_tables
from dct_histogramm import dct_histogram

def double_compression_score(hist):
    # Autokorrelation des Histogramms
    ac = correlate(hist, hist, mode="full")
    mid = len(ac) // 2
    # Suche stärksten Neben-Peak (abstand > 1)
    main = ac[mid]
    side_peaks = ac[mid+2:mid+50]  # überspringe Nullversatz
    peak = float(np.max(side_peaks))
    score = peak / main if main != 0 else 0.0
    return score

if __name__ == "__main__":
    img_path = sys.argv[1]
    qstats = extract_quant_tables(img_path)
    hist_data = dct_histogram(img_path)["histogram"]
    score = double_compression_score(np.array(hist_data))
    out = {
        "quant_tables": qstats,
        "double_compression_score": score
    }
    print(json.dumps(out))
