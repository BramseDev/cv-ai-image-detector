import pandas as pd
import numpy as np
import seaborn as sns
import matplotlib.pyplot as plt
from sklearn.model_selection import train_test_split
from sklearn.metrics import confusion_matrix, classification_report, roc_curve, auc
from sklearn.linear_model import LogisticRegression

# "Double Compression Score" als Unterscheidungsmerkmal
# zwischen KI-generierten und normalen Bildern

# style
#sns.set_theme(style="whitegrid")
#custom_palette = ["#E0D4E7", "#CCDCEB"]  # lavender for real, blue for AI
#sns.set_palette(custom_palette)


#############################
# 1. DATENLADUNG UND ANALYSE
#############################

# Daten laden
df = pd.read_csv('compression_results.csv')

# Basis-Daten Übersicht
print("First five rows of the dataset:")
print(df.head())

print("\nDataset information:")
print(df.info())

# Null-Werte prüfen
print("\nMissing values check:")
print(df.isnull().sum())

# Statistische Kennzahlen
print("\nBasic statistics:")
print(df.describe())

# Sind Daten ausgewogen?
print("\nClass distribution:")
class_distribution = df['gitCategory'].value_counts()
print(class_distribution)

# Klassenverteilung visualisieren
plt.figure(figsize=(8, 5))
ax = sns.countplot(x='gitCategory', data=df)
plt.title('Distribution of AI vs Non-AI Images')
plt.xlabel('gitCategory')
plt.ylabel('Count')

# Werte über den Balken anzeigen
for p in ax.patches:
    ax.annotate(f'{p.get_height()}',
                (p.get_x() + p.get_width() / 2., p.get_height()),
                ha='center', va='bottom', fontsize=12)

plt.savefig('class_distribution.png')
plt.close()

# Gruppierung nach Kategorie mit Statistiken
group_stats = df.groupby('gitCategory')['Double Compression Score'].describe()
print("\nGruppierte Statistiken nach Kategorie:")
print(group_stats)

# Visualisierung für:
    # Boxplots (zeigen Median, Quartile, Ausreißer)
    # Histogramme (für Häufigkeitsverteilung)
    # Violin-Plots (kombinieren Boxplot und Dichteverteilung)
    # Strip-Plots (zeigen individuelle Datenpunkte)
    # Barplot (für Durchschnittswerten)

# Boxplot
plt.figure(figsize=(10, 6))
sns.boxplot(x='gitCategory', y='Double Compression Score', data=df)
plt.title('Double Compression Score Distribution by gitCategory')
plt.savefig('score_boxplot.png')
plt.close()

# Histogramm der Gesamtdaten
plt.figure(figsize=(10, 6))
df.hist(column='Double Compression Score', bins=15, grid=True)
plt.suptitle('Histogramm aller Double Compression Scores')
plt.savefig('histogram_all.png')
plt.close()

# Boxplot nach Kategorie
plt.figure(figsize=(10, 6))
df.boxplot(column='Double Compression Score', by='gitCategory')
plt.title('Boxplot der Double Compression Scores nach Kategorie')
plt.suptitle('')  # Entfernt den automatischen Suptitle
plt.savefig('boxplot_by_category.png')
plt.close()

# Verteilung der Double Compression Scores für ki und nonKi
plt.figure(figsize=(10, 6))
sns.boxplot(x='gitCategory', y='Double Compression Score', data=df)
plt.title('Double Compression Score Verteilung nach Kategorie')
plt.savefig('boxplot_distribution.png')
plt.close()

# Histogramm: Häufigkeitsverteilung der Scores für beide Kategorien
plt.figure(figsize=(10, 6))
sns.histplot(data=df, x='Double Compression Score', hue='gitCategory', bins=20, multiple='stack')
plt.title('Histogramm der Double Compression Scores')
plt.savefig('histogram_by_category.png')
plt.close()

# Violin-Plot
plt.figure(figsize=(10, 6))
sns.violinplot(x='gitCategory', y='Double Compression Score', data=df)
plt.title('Violin-Plot der Double Compression Scores')
plt.savefig('violin_plot.png')
plt.close()

# Strip-Plot
plt.figure(figsize=(10, 6))
sns.stripplot(x='gitCategory', y='Double Compression Score', data=df, jitter=True)
plt.title('Strip-Plot der Double Compression Scores')
plt.savefig('strip_plot.png')
plt.close()

# Vergleich: durchschnittliche Scores zwischen ki und nonki
plt.figure(figsize=(10, 6))
ax = sns.barplot(x='gitCategory', y='Double Compression Score', data=df, estimator='mean')
plt.title('Durchschnittlicher Double Compression Score pro Kategorie')

# Werte über den Balken anzeigen
for p in ax.patches:
    ax.annotate(f'{p.get_height():.4f}',
                (p.get_x() + p.get_width() / 2., p.get_height() + 0.002),
                ha='center', va='bottom', fontsize=11)

plt.savefig('average_scores.png')
plt.close()

# Optional: Einfaches Klassifikationsmodell hinzufügen
print("\nTraining eines einfachen Klassifikationsmodells:")

# Kategorie in numerische Werte konvertieren (falls nötig)
if df['gitCategory'].dtype == 'object':
    df['target'] = df['gitCategory'].map({'ki': 1, 'nonKi': 0})  # Anpassen je nach genauen Kategorienamen
else:
    df['target'] = df['gitCategory']  # Falls bereits numerisch

# Features und Target
X = df[['Double Compression Score']]
y = df['target']

# Train-Test-Split
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.3, random_state=42)

# Modell trainieren
model = LogisticRegression()
model.fit(X_train, y_train)

# Modellbewertung
y_pred = model.predict(X_test)
print("\nConfusion Matrix:")
print(confusion_matrix(y_test, y_pred))

print("\nClassification Report:")
print(classification_report(y_test, y_pred))

print("\nModel coefficients:")
print(f"Intercept: {model.intercept_[0]:.4f}")
print(f"Coefficient: {model.coef_[0][0]:.4f}")

# ROC Kurve: Zeigt die Leistung des Modells bei verschiedenen Schwellenwerten
# Je höher die AUC (Area Under Curve), desto besser ist das Modell (1.0 = perfekt)

y_pred_proba = model.predict_proba(X_test)[:, 1]
fpr, tpr, _ = roc_curve(y_test, y_pred_proba)
roc_auc = auc(fpr, tpr)

plt.figure(figsize=(8, 6))
plt.plot(fpr, tpr, color='darkorange', lw=2, label=f'ROC Curve (AUC = {roc_auc:.2f})')
plt.plot([0, 1], [0, 1], color='navy', lw=2, linestyle='--')
plt.xlim([0.0, 1.0])
plt.ylim([0.0, 1.05])
plt.xlabel('False Positive Rate')
plt.ylabel('True Positive Rate')
plt.title('Receiver Operating Characteristic (ROC)')
plt.legend(loc="lower right")
plt.savefig('roc_curve.png')
plt.close()

print("\nAnalyse abgeschlossen. Alle Diagramme wurden gespeichert.")