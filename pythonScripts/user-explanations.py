import json
import sys

def generate_user_explanations(artifacts_data, advanced_data=None):
    """Generiert benutzerfreundliche Erklärungen aus Analysedaten"""

    explanations = {
        'analysis_summary': {
            'verdict': 'inconclusive',
            'confidence_level': 'low',
            'reasoning': []
        },
        'detailed_findings': [],
        'recommendations': [],
        'what_to_look_for': []
    }

    # Artefakt-Analysen auswerten
    if 'overall_assessment' in artifacts_data:
        overall = artifacts_data['overall_assessment']
        ai_score = overall.get('ai_probability_score', 0)

        # Verdachtsgrad bestimmen
        if ai_score >= 0.7:
            explanations['analysis_summary']['verdict'] = 'likely_ai_generated'
            explanations['analysis_summary']['confidence_level'] = 'high'
            explanations['recommendations'].extend([
                "🚨 Mehrere starke AI-Indikatoren gefunden",
                "🔬 Weitere Analysen empfohlen"
            ])
        elif ai_score >= 0.4:
            explanations['analysis_summary']['verdict'] = 'possibly_ai_generated'
            explanations['analysis_summary']['confidence_level'] = 'medium'
            explanations['recommendations'].extend([
                "⚠️ Einige verdächtige Merkmale gefunden",
                "🔄 Zusätzliche Tests könnten hilfreich sein"
            ])
        else:
            explanations['analysis_summary']['verdict'] = 'likely_authentic'
            explanations['analysis_summary']['confidence_level'] = 'medium'
            explanations['recommendations'].append("✅ Keine starken AI-Indikatoren gefunden")

    # Spezifische Befunde erklären
    if artifacts_data.get('ringing_artifacts', {}).get('ai_ringing_indicator', False):
        intensity = artifacts_data['ringing_artifacts']['ringing_intensity']
        explanations['detailed_findings'].append({
            'finding': 'Edge Ringing Effects',
            'severity': 'high',
            'explanation': f"Starke Oszillationen um Kanten erkannt (Intensity: {intensity:.1f})",
            'user_tip': "Achten Sie auf 'Geisterlinien' neben scharfen Konturen."
        })

    if artifacts_data.get('png_artifacts', {}).get('ai_png_indicator', False):
        colors = artifacts_data['png_artifacts']['unique_colors']
        explanations['detailed_findings'].append({
            'finding': 'PNG Anomalies',
            'severity': 'medium',
            'explanation': f"Ungewöhnliche Farbverteilung: {colors:,} einzigartige Farben",
            'user_tip': "AI-Bilder haben oft unnatürliche Farbpaletten."
        })

    # Erweiterte Analysen (falls vorhanden)
    if advanced_data and 'ai_patterns' in advanced_data:
        patterns = advanced_data['ai_patterns']
        if patterns.get('ai_pattern_indicator', False):
            explanations['detailed_findings'].append({
                'finding': 'AI Pattern Detection',
                'severity': 'high',
                'explanation': f"Unnatürliche Frequenzmuster erkannt (Balance: {patterns['frequency_balance']:.3f})",
                'user_tip': "Moderne AI-Modelle hinterlassen charakteristische Frequenz-Signaturen."
            })

    return explanations

def main():
    if len(sys.argv) < 2:
        print("Usage: python3 user-explanations.py <artifacts_json> [advanced_json]", file=sys.stderr)
        sys.exit(1)

    # Lade Artefakt-Daten
    with open(sys.argv[1], 'r') as f:
        artifacts_data = json.load(f)

    # Lade erweiterte Daten (optional)
    advanced_data = None
    if len(sys.argv) > 2:
        with open(sys.argv[2], 'r') as f:
            advanced_data = json.load(f)

    explanations = generate_user_explanations(artifacts_data, advanced_data)
    print(json.dumps(explanations, indent=2))

if __name__ == "__main__":
    main()