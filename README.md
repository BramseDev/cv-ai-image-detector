# AI Image Detection System

This Bachelor project explores the integration of computer vision and artificial intelligence to automatically analyze and evaluate AI-generated images. 

## Overview
Our system processes images through advanced computer vision techniques to extract relevant features, then applies AI models to assess and assign weighted scores based on predefined detection criteria.

## Acknowledgments
Special thanks to [@lawi2022](https://github.com/lawi2022) and [@derfisch98](https://github.com/derfisch98) for their significant contributions to the development of this AI detection system, along with all other contributors.



## Features

- **Multi-Modal Analysis**: Combines traditional computer vision, metadata analysis, and AI models
- **Comprehensive Dashboard**: Web-based monitoring and metrics dashboard
- **Format Support**: JPEG, PNG with specialized analysis for each format
- **Forensic Analysis**: EXIF data, C2PA certificates, compression artifacts
- **Advanced Algorithms**: Fourier transforms, pattern detection, lighting physics analysis

## Architecture

The system utilizes a modular pipeline architecture consisting of the following components:

- **Go Backend**: Implements core backend logic for certain tests and processing pipelines.
- **Python Scripts**: Provide specialized analysis algorithms for feature extraction and image evaluation.
- **Rust Components**: Leverage the `c2pa-rust` library for C2PA certificate validation, enhancing image authenticity checks.
- **Web Dashboard**: Offers a real-time monitoring interface for system status and results visualization.
- **Docker Deployment**: All components are containerized for easy and reproducible deployment.


## Quick Start

### Prerequisites

- Docker & Docker Compose
- Make (optional, for convenience commands)

### One-Command Deployment

```bash
make deploy
```

This will:
1. Build the Docker image
2. Start all services
3. Run health checks
4. Display access URLs

### Manual Deployment

```bash
# Build the image
docker-compose build --no-cache

# Start services
docker-compose up -d

# Check status
docker-compose ps
```

### Access Points

After deployment, access these URLs:

- **Main API**: http://localhost:8080/upload
- **Health Check**: http://localhost:8080/health
- **Metrics API**: http://localhost:8080/metrics
- **Metrics Dashboard**: http://localhost:8080/dashboard/metrics
- **Health Dashboard**: http://localhost:8080/dashboard/health

## API Usage

### Upload Image for Analysis

```bash
curl -X POST -F "file=@your-image.jpg" http://localhost:8080/upload
```

### Example Response

```json
{
  "verdict": "Likely AI Generated",
  "probability": 75.3,
  "confidence": 0.87,
  "analysis_summary": {
    "traditional_computer_vision": {
      "verdict": "Strong AI Indicators",
      "explanation": "Computer Vision detected AI patterns in: Visual Artifacts, Mathematical Patterns"
    },
    "metadata_forensics": {
      "verdict": "Authenticity Indicators", 
      "explanation": "Rich EXIF data suggests camera origin"
    },
    "ai_deep_learning": {
      "verdict": "Strong AI Indicators",
      "explanation": "AI model detects 89% probability of synthetic generation"
    }
  },
  "detailed_analysis": {
    "artifacts": {
      "ai_probability_score": 0.82,
      "positive_indicators": 7,
      "total_indicators": 10
    },
    "lighting-analysis": {
      "ai_lighting_score": 0.71,
      "anomalies": ["Inconsistent shadow directions"]
    }
  }
}
```

## Analysis Components

### Traditional Computer Vision
- **Visual Artifacts**: 

detect-artifacts.py

 - JPEG compression, ringing effects
  
- **Pixel Analysis**: Mathematical frequency analysis and noise characterization  

- **Color Balance**: 
analyze_color_balance.py
 - Unnatural color distributions

- **Lighting Physics**: 
analyze_lighting.py
 - Shadow consistency, light source analysis

- **Object Coherence**: 
analyze_coherence.py
 - Perspective and edge consistency

- **Advanced Patterns**: 
advanced-artifacts.py
 - Fourier transforms, AI-specific signatures

### Metadata Forensics
- **EXIF Analysis**: 
exif_analyzer
 - Camera metadata validation

- **C2PA Certificates**: 
c2pa-rust
 - Content authenticity verification
- **Compression Analysis**: 

analyze_compression.py
 - Multi-compression detection

### AI Deep Learning
- **Neural Network**: 
modeltest.py
 - Trained detection model

- **Pattern Recognition**: Synthetic noise and generation artifacts

## System Monitoring

### Real-time Dashboard

The system includes comprehensive monitoring:

```bash
# View live metrics
open http://localhost:8080/dashboard/metrics

# View system health  
open http://localhost:8080/dashboard/health
```

### Key Metrics

- **AI Detection Rate**: Percentage of images flagged as AI-generated
- **Analysis Performance**: Average processing time per image
- **Cache Efficiency**: Hit/miss ratios for performance optimization
- **Error Tracking**: Failed analysis detection

### Logs

```bash
# View application logs
make logs

# View all service logs
docker-compose logs -f
```

## Configuration

### Environment Variables

Create `.env` file:

```env
# Analysis Configuration
EARLY_EXIT_ENABLED=true
CACHE_TTL=30m
MAX_FILE_SIZE=50MB

# Performance Tuning
ANALYSIS_TIMEOUT=120s
PIPELINE_WORKERS=4

# Monitoring
METRICS_ENABLED=true
LOG_LEVEL=info
```

## Development Setup

### Local Development

```bash
# Setup dependencies
make setup

# Install Python requirements
cd pythonScripts && pip install -r requirements.txt

# Build Go modules
go mod tidy

# Run tests
make test
```

### File Structure

```
├── cmd/server/           # Main application entry
├── internal/handlers/    # HTTP request handlers
├── pkg/analyzer/         # Analysis pipeline components
├── pythonScripts/        # Computer vision algorithms
├── dashboard/           # Web monitoring interface
├── ai-analyse/          # Deep learning models
├── cache/              # Caching implementation
├── monitoring/         # Metrics collection
└── docker-compose.yml  # Container orchestration
```

## Performance Optimization

### Caching Strategy

The system implements intelligent caching in 

pipelines.go:

- **Image Fingerprinting**: SHA256-based cache keys
- **Result Persistence**: 30-minute cache TTL
- **Early Exit**: Skip expensive analysis when definitive results found

### Analysis Pipeline

1. **Core Analysis** (10-20 seconds): Computer vision algorithms  
2. **Deep Analysis** (20-30 seconds): AI model inference


## Troubleshooting

### Common Issues

#### Service Won't Start
```bash
# Check Docker status
docker-compose ps

# View detailed logs
docker-compose logs image-analyzer

# Restart services
make restart

# when nothing helps
source venv/bin/activate
cd pythonscripts
source venv/bin/activate
cd ..
go run cmd/server/main.go
```

#### Analysis Failures
```bash
# Check Python dependencies
docker-compose exec image-analyzer python3 -c "import cv2, numpy, scipy"

# Test individual scripts
docker-compose exec image-analyzer python3 pythonScripts/detect-artifacts.py /path/to/test.jpg
```

#### Performance Issues
```bash
# Monitor resource usage
docker stats

# Check cache performance
curl http://localhost:8080/metrics | jq '.system.cache_hit_rate'
```

### Health Checks

```bash
# System health
curl http://localhost:8080/health

# Specific component test
curl -X POST -F "file=@test-image.jpg" http://localhost:8080/upload
```

## Advanced Features

### Custom Analysis Scripts

Add new detection algorithms in 

pythonScripts:

```python
def analyze_custom_feature(image_path):
    """Custom analysis implementation"""
    # Your detection logic here
    return {
        "custom_score": 0.75,
        "indicators": ["pattern1", "pattern2"]
    }
```

Register in 

pipelines.go:

```go
{
    Name:         "custom-analysis",
    Priority:     3,
    Timeout:      15 * time.Second,
    Analyzer:     pythonrunner.RunCustomAnalysis,
}
```

### Metrics Integration

Export metrics to external systems:

```bash
# Prometheus format
curl http://localhost:8080/metrics

# JSON format for custom integrations
curl http://localhost:8080/metrics | jq '.business.ai_detection_rate'
```

## Production Deployment

### Security Considerations

- Configure file upload limits
- Implement rate limiting
- Use HTTPS in production
- Secure temporary file handling

### Scaling

- Use container orchestration (Kubernetes)
- Implement horizontal scaling for analysis workers
- Configure load balancing for API endpoints
- Monitor resource usage and auto-scaling triggers

### Backup Strategy

```bash
# Export analysis history
curl http://localhost:8080/dashboard/metrics > metrics-backup.json

# Backup configuration
docker-compose config > docker-config-backup.yml
```

## License

This project is licensed under the MIT License. See LICENSE file for details.

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## Support

- **Documentation**: Check this README and inline code comments
- **Issues**: Report bugs via GitHub Issues
- **Performance**: Monitor via built-in dashboard
- **Logs**: Access via `make logs` or Docker commands

---

## Scoring Interpretation

| Score Range | Interpretation           | Description                                                                              |
|-------------|--------------------------|------------------------------------------------------------------------------------------|
| 0 - 19      | Very Likely Authentic    | Strong indicators of genuine photography. Natural compression, EXIF data present. Realistic lighting and color distribution. |
| 20 - 39     | Likely Authentic         | Mostly authentic features. Minor anomalies within normal range. Probably a real photo.    |
| 40 - 59     | Possibly AI Generated    | Mixed signals detected. Some suspicious characteristics present. Further analysis recommended. |
| 60 - 79     | Very Likely AI Generated | Multiple AI indicators found. Unnatural artifacts or patterns. High probability of AI generation. |
| 80 - 100    | Almost Certainly AI Generated | Strong AI signals in multiple areas. Characteristic AI artifacts. Very high confidence in AI detection. |


**Built with**: Go, Python, Rust, Docker, OpenCV, scikit-image, C2PA
