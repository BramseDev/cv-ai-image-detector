import os
import csv
import json
import analyze_compression

def run_csv_analysis():
    input_dirs = ["../testPictures/ki", "../testPictures/nonKi"]
    output_file = "compression_results.csv"
    results = []

    for input_dir in input_dirs:
        category = os.path.basename(input_dir)
        for root, _, files in os.walk(input_dir):
            for file in files:
                if file.lower().endswith(('.jpg', '.jpeg')):
                    file_path = os.path.join(root, file)
                    analysis = analyze_compression.analyze_image(file_path)
                    score = analysis.get("double_compression_score", 0.0)
                    results.append([category, score])

    # CSV schreiben
    with open(output_file, mode='w', newline='', encoding='utf-8') as csvfile:
        writer = csv.writer(csvfile)
        writer.writerow(["Category", "Double Compression Score"])
        writer.writerows(results)

if __name__ == '__main__':
    run_csv_analysis()
