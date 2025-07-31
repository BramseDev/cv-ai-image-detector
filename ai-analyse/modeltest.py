import torch
from torchvision import models, transforms
from PIL import Image
import argparse
import os


def load_model(checkpoint_path, device):
    model = models.resnet18(pretrained=False)
    model.fc = torch.nn.Linear(model.fc.in_features, 2)  # 2 Klassen: fake vs. real
    state = torch.load(checkpoint_path, map_location=device)
    if 'model_state_dict' in state:
        state = state['model_state_dict']
    model.load_state_dict(state)
    model.eval()
    return model.to(device)


def predict(image_path, model, device):
    preprocess = transforms.Compose([
        transforms.Resize((224, 224)),
        transforms.ToTensor(),
    ])
    img = Image.open(image_path).convert('RGB')
    inp = preprocess(img).unsqueeze(0).to(device)
    with torch.no_grad():
        out = model(inp)
        probs = torch.softmax(out, dim=1)[0]
        idx = probs.argmax().item()
    classes = ['fake', 'real']
    return classes[idx], probs[idx].item()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Fake-vs-Real Detector")
    parser.add_argument("--model", "-m", required=True,
                        help="Pfad zur .pth-Modelldatei")
    parser.add_argument("--image", "-i", required=True,
                        help="Pfad zum einzeln zu klassifizierenden Bild")
    args = parser.parse_args()

    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    model = load_model(args.model, device)

    label, prob = predict(args.image, model, device)


    import json
    result = {
        "prediction": label,
        "probability": float(prob),
        "confidence": float(prob),
        "model_type": "resnet18",
        "ai_model_analysis": {
            "predicted_class": label,
            "confidence_score": float(prob),
            "is_ai_generated": label == "fake",
            "authenticity_score": float(1.0 - prob) if label == "fake" else float(prob)
        }
    }

    print(json.dumps(result))