#!/usr/bin/env python3
"""
AI Image Detection API Test Script

This script tests all images in the testPics folder against the AI detection API
and displays the results in a formatted table.

Usage:
    python test_images.py [--limit N]
    
    --limit N: Test only the first N images from each folder (default: no limit)
"""

import os
import urllib.request
import urllib.parse
import json
from pathlib import Path
import time
import mimetypes
import argparse

# Configuration
API_URL = "http://localhost:8080/upload"
AI_FOLDER = "ai"
NONAI_FOLDER = "nonai"
TESTPICS_FOLDER = "testPics"  # Fallback for backward compatibility
SUPPORTED_EXTENSIONS = {'.jpg', '.jpeg', '.png', '.bmp', '.tiff', '.webp'}

# Detection threshold (images above this probability are considered AI-generated)
# Optimized threshold based on score distribution analysis
AI_DETECTION_THRESHOLD = 55.0  # Optimized for best accuracy (55% vs 45% at 50% threshold)

def test_image(image_path):
    """
    Send an image to the API and return the analysis results.
    
    Args:
        image_path (str): Path to the image file
        
    Returns:
        dict: API response or error information
    """
    try:
        # Create multipart form data
        boundary = '----WebKitFormBoundary7MA4YWxkTrZu0gW'
        
        with open(image_path, 'rb') as img_file:
            img_data = img_file.read()
        
        # Get MIME type
        mime_type, _ = mimetypes.guess_type(image_path)
        if not mime_type:
            mime_type = 'application/octet-stream'
        
        # Build multipart body
        body = []
        body.append(f'--{boundary}'.encode())
        body.append(f'Content-Disposition: form-data; name="image"; filename="{os.path.basename(image_path)}"'.encode())
        body.append(f'Content-Type: {mime_type}'.encode())
        body.append(b'')
        body.append(img_data)
        body.append(f'--{boundary}--'.encode())
        
        body_bytes = b'\r\n'.join(body)
        
        # Create request
        req = urllib.request.Request(
            API_URL,
            data=body_bytes,
            headers={
                'Content-Type': f'multipart/form-data; boundary={boundary}',
                'Content-Length': str(len(body_bytes))
            },
            method='POST'
        )
        
        # Send request
        with urllib.request.urlopen(req, timeout=120) as response:
            if response.status == 200:
                return json.loads(response.read().decode())
            else:
                return {
                    'error': f"HTTP {response.status}",
                    'message': response.read().decode()
                }
                
    except urllib.error.URLError as e:
        return {
            'error': 'Connection Error',
            'message': str(e.reason)
        }
    except Exception as e:
        return {
            'error': 'Unexpected Error',
            'message': str(e)
        }

def format_verdict(verdict, probability):
    """Format the verdict with color coding for terminal output."""
    if probability < 20:
        return f"🟢 {verdict}"
    elif probability < 40:
        return f"🔵 {verdict}"
    elif probability < 60:
        return f"🟡 {verdict}"
    elif probability < 80:
        return f"🟠 {verdict}"
    else:
        return f"🔴 {verdict}"

def print_header():
    """Print the table header."""
    print("\n" + "="*90)
    print("AI IMAGE DETECTION RESULTS")
    print("="*90)
    print(f"{'Image':<30} {'Label':<8} {'Verdict':<25} {'Probability':<12} {'Confidence':<10} {'Correct':<8}")
    print("-"*90)

def print_result(filename, result, true_label):
    """Print a single test result."""
    if 'error' in result:
        print(f"{filename:<30} {true_label:<8} ❌ {result['error']:<23} {'N/A':<12} {'N/A':<10} {'N/A':<8}")
        if 'message' in result:
            print(f"{' '*30} └─ {result['message'][:50]}...")
    else:
        # Extract from the 'analysis' section of the response
        analysis = result.get('analysis', {})
        verdict = analysis.get('verdict', 'Unknown')
        probability = analysis.get('probability', 0)
        confidence = analysis.get('confidence', 0)
        
        # Determine if prediction is correct
        predicted_ai = probability >= AI_DETECTION_THRESHOLD
        actual_ai = true_label == "AI"
        is_correct = predicted_ai == actual_ai
        
        formatted_verdict = format_verdict(verdict, probability)
        correct_symbol = "✅" if is_correct else "❌"
        
        print(f"{filename:<30} {true_label:<8} {formatted_verdict:<30} {probability:<8.1f}%   {confidence:<8.1f}   {correct_symbol:<8}")

def get_image_files_from_folders(limit_per_folder=None):
    """Get image files from ai and nonai folders, or fallback to testPics.
    
    Args:
        limit_per_folder (int, optional): Maximum number of images to take from each folder
    """
    image_data = []
    
    # Check for ai and nonai folders in root directory first
    if os.path.exists(AI_FOLDER) and os.path.exists(NONAI_FOLDER):
        print(f"🔍 Found labeled folders: {AI_FOLDER}/ and {NONAI_FOLDER}/")
        
        # Get AI images
        if os.path.exists(AI_FOLDER):
            ai_files = []
            for file in sorted(os.listdir(AI_FOLDER)):  # Sort for consistent ordering
                file_path = Path(file)
                if file_path.suffix.lower() in SUPPORTED_EXTENSIONS:
                    ai_files.append({
                        'filename': file,
                        'path': os.path.join(AI_FOLDER, file),
                        'label': 'AI'
                    })
            
            # Apply limit if specified
            if limit_per_folder:
                ai_files = ai_files[:limit_per_folder]
            
            image_data.extend(ai_files)
        
        # Get Non-AI images
        if os.path.exists(NONAI_FOLDER):
            nonai_files = []
            for file in sorted(os.listdir(NONAI_FOLDER)):  # Sort for consistent ordering
                file_path = Path(file)
                if file_path.suffix.lower() in SUPPORTED_EXTENSIONS:
                    nonai_files.append({
                        'filename': file,
                        'path': os.path.join(NONAI_FOLDER, file),
                        'label': 'Non-AI'
                    })
            
            # Apply limit if specified
            if limit_per_folder:
                nonai_files = nonai_files[:limit_per_folder]
            
            image_data.extend(nonai_files)
    
    # Check for ai and nonai folders inside testPics
    elif (os.path.exists(os.path.join(TESTPICS_FOLDER, AI_FOLDER)) and 
          os.path.exists(os.path.join(TESTPICS_FOLDER, NONAI_FOLDER))):
        print(f"🔍 Found labeled folders: {TESTPICS_FOLDER}/{AI_FOLDER}/ and {TESTPICS_FOLDER}/{NONAI_FOLDER}/")
        
        # Get AI images
        ai_path = os.path.join(TESTPICS_FOLDER, AI_FOLDER)
        if os.path.exists(ai_path):
            ai_files = []
            for file in sorted(os.listdir(ai_path)):  # Sort for consistent ordering
                file_path = Path(file)
                if file_path.suffix.lower() in SUPPORTED_EXTENSIONS:
                    ai_files.append({
                        'filename': file,
                        'path': os.path.join(ai_path, file),
                        'label': 'AI'
                    })
            
            # Apply limit if specified
            if limit_per_folder:
                ai_files = ai_files[:limit_per_folder]
            
            image_data.extend(ai_files)
        
        # Get Non-AI images
        nonai_path = os.path.join(TESTPICS_FOLDER, NONAI_FOLDER)
        if os.path.exists(nonai_path):
            nonai_files = []
            for file in sorted(os.listdir(nonai_path)):  # Sort for consistent ordering
                file_path = Path(file)
                if file_path.suffix.lower() in SUPPORTED_EXTENSIONS:
                    nonai_files.append({
                        'filename': file,
                        'path': os.path.join(nonai_path, file),
                        'label': 'Non-AI'
                    })
            
            # Apply limit if specified
            if limit_per_folder:
                nonai_files = nonai_files[:limit_per_folder]
            
            image_data.extend(nonai_files)
    
    # Fallback to testPics folder (unlabeled)
    elif os.path.exists(TESTPICS_FOLDER):
        print(f"🔍 Using fallback folder: {TESTPICS_FOLDER}/")
        files = []
        for file in sorted(os.listdir(TESTPICS_FOLDER)):  # Sort for consistent ordering
            file_path = Path(file)
            if file_path.suffix.lower() in SUPPORTED_EXTENSIONS:
                files.append({
                    'filename': file,
                    'path': os.path.join(TESTPICS_FOLDER, file),
                    'label': 'Unknown'
                })
        
        # Apply limit if specified (for unlabeled data, apply to total)
        if limit_per_folder:
            files = files[:limit_per_folder]
        
        image_data.extend(files)
    
    return image_data

def calculate_accuracy_metrics(results):
    """Calculate accuracy metrics for labeled data."""
    if not results:
        return None
    
    # Only calculate metrics for labeled data (not 'Unknown')
    labeled_results = [(filename, data) for filename, data in results.items() 
                      if data.get('true_label') in ['AI', 'Non-AI'] and 'error' not in data]
    
    if not labeled_results:
        return None
    
    total = len(labeled_results)
    correct = 0
    true_positives = 0  # Correctly identified AI images
    false_positives = 0  # Non-AI images incorrectly identified as AI
    true_negatives = 0  # Correctly identified Non-AI images
    false_negatives = 0  # AI images incorrectly identified as Non-AI
    
    for filename, data in labeled_results:
        analysis = data.get('analysis', {})
        probability = analysis.get('probability', 0)
        true_label = data.get('true_label')
        
        predicted_ai = probability >= AI_DETECTION_THRESHOLD
        actual_ai = true_label == "AI"
        
        if predicted_ai == actual_ai:
            correct += 1
        
        if actual_ai and predicted_ai:
            true_positives += 1
        elif actual_ai and not predicted_ai:
            false_negatives += 1
        elif not actual_ai and predicted_ai:
            false_positives += 1
        elif not actual_ai and not predicted_ai:
            true_negatives += 1
    
    accuracy = correct / total
    
    # Calculate precision, recall, and F1-score
    precision = true_positives / (true_positives + false_positives) if (true_positives + false_positives) > 0 else 0
    recall = true_positives / (true_positives + false_negatives) if (true_positives + false_negatives) > 0 else 0
    f1_score = 2 * (precision * recall) / (precision + recall) if (precision + recall) > 0 else 0
    
    return {
        'total': total,
        'correct': correct,
        'accuracy': accuracy,
        'true_positives': true_positives,
        'false_positives': false_positives,
        'true_negatives': true_negatives,
        'false_negatives': false_negatives,
        'precision': precision,
        'recall': recall,
        'f1_score': f1_score
    }
def main():
    """Main function to test all images in the ai and nonai folders."""
    # Parse command line arguments
    parser = argparse.ArgumentParser(description='Test AI image detection API')
    parser.add_argument('--limit', type=int, help='Limit number of images tested from each folder')
    args = parser.parse_args()
    
    # Get image files from folders
    image_data = get_image_files_from_folders(limit_per_folder=args.limit)
    
    if not image_data:
        print(f"❌ No supported image files found!")
        print(f"Create '{AI_FOLDER}' and '{NONAI_FOLDER}' folders with images, or use '{TESTPICS_FOLDER}' folder")
        print(f"Supported formats: {', '.join(SUPPORTED_EXTENSIONS)}")
        return
    
    print(f"🔍 Found {len(image_data)} image(s) to test")
    if args.limit:
        print(f"📊 Limited to {args.limit} image(s) per folder")
    
    # Count by label
    label_counts = {}
    for img in image_data:
        label = img['label']
        label_counts[label] = label_counts.get(label, 0) + 1
    
    for label, count in label_counts.items():
        print(f"   • {label}: {count} images")
    
    print(f"🎯 Detection threshold: {AI_DETECTION_THRESHOLD}% (adjustable in script)")
    
    # Test API connection first
    try:
        with urllib.request.urlopen("http://localhost:8080/health", timeout=5) as response:
            if response.status != 200:
                print("❌ API health check failed. Make sure the server is running!")
                return
    except:
        print("❌ Cannot connect to API. Make sure the server is running on localhost:8080")
        return
    
    print("✅ API connection successful")
    
    print_header()
    
    # Test each image
    results = {}
    for i, img_info in enumerate(sorted(image_data, key=lambda x: (x['label'], x['filename'])), 1):
        filename = img_info['filename']
        image_path = img_info['path']
        true_label = img_info['label']
        
        print(f"Testing {i}/{len(image_data)}: {filename}...", end=" ")
        
        start_time = time.time()
        result = test_image(image_path)
        duration = time.time() - start_time
        
        # Add true label to result for accuracy calculation
        result['true_label'] = true_label
        
        print(f"({duration:.1f}s)")
        print_result(filename, result, true_label)
        results[filename] = result
    
    # Summary
    print("-"*90)
    successful_tests = [r for r in results.values() if 'error' not in r]
    if successful_tests:
        # Extract probabilities and confidences from the analysis section
        probabilities = [r.get('analysis', {}).get('probability', 0) for r in successful_tests]
        confidences = [r.get('analysis', {}).get('confidence', 0) for r in successful_tests]
        
        avg_probability = sum(probabilities) / len(probabilities)
        avg_confidence = sum(confidences) / len(confidences)
        
        print(f"📊 SUMMARY:")
        print(f"   • Total images tested: {len(image_data)}")
        print(f"   • Successful analyses: {len(successful_tests)}")
        print(f"   • Average AI probability: {avg_probability:.1f}%")
        print(f"   • Average confidence: {avg_confidence:.1f}")
        
        # Calculate accuracy metrics for labeled data
        accuracy_metrics = calculate_accuracy_metrics(results)
        if accuracy_metrics:
            print(f"\n🎯 ACCURACY METRICS (Threshold: {AI_DETECTION_THRESHOLD}%):")
            print(f"   • Overall Accuracy: {accuracy_metrics['accuracy']:.1%} ({accuracy_metrics['correct']}/{accuracy_metrics['total']})")
            print(f"   • Precision: {accuracy_metrics['precision']:.1%}")
            print(f"   • Recall (Sensitivity): {accuracy_metrics['recall']:.1%}")
            print(f"   • F1-Score: {accuracy_metrics['f1_score']:.3f}")
            
            print(f"\n📈 CONFUSION MATRIX:")
            print(f"   • True Positives (AI → AI): {accuracy_metrics['true_positives']}")
            print(f"   • False Positives (Non-AI → AI): {accuracy_metrics['false_positives']}")
            print(f"   • True Negatives (Non-AI → Non-AI): {accuracy_metrics['true_negatives']}")
            print(f"   • False Negatives (AI → Non-AI): {accuracy_metrics['false_negatives']}")
        
        # Count by verdict ranges
        very_authentic = sum(1 for p in probabilities if p < 20)
        likely_authentic = sum(1 for p in probabilities if 20 <= p < 40)
        possibly_ai = sum(1 for p in probabilities if 40 <= p < 60)
        likely_ai = sum(1 for p in probabilities if 60 <= p < 80)
        almost_certainly_ai = sum(1 for p in probabilities if p >= 80)
        
        print(f"\n📈 VERDICT DISTRIBUTION:")
        print(f"   🟢 Very Likely Authentic (0-19%): {very_authentic}")
        print(f"   🔵 Likely Authentic (20-39%): {likely_authentic}")
        print(f"   🟡 Possibly AI Generated (40-59%): {possibly_ai}")
        print(f"   🟠 Very Likely AI Generated (60-79%): {likely_ai}")
        print(f"   🔴 Almost Certainly AI Generated (80-100%): {almost_certainly_ai}")
    
    print("="*90)
    
    # Save detailed results to JSON file
    with open('test_results.json', 'w') as f:
        json.dump(results, f, indent=2)
    print(f"💾 Detailed results saved to test_results.json")

if __name__ == "__main__":
    main()
