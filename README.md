# AI Image Detection System

This Bachelor project explores the integration of computer vision and artificial intelligence to automatically analyze and evaluate AI-generated images. 

## Overview
Our system processes images through advanced computer vision techniques to extract relevant features, then applies AI models to assess and assign weighted scores based on predefined detection criteria.

## Acknowledgments

This bachelor project was a collaborative effort. Special thanks to [@lawi2022](https://github.com/lawi2022), [@derfisch98](https://github.com/derfisch98), [@fabian-thies](https://github.com/fabian-thies), [@iakdis](https://github.com/iakdis), and [@RuneMolzen](https://github.com/RuneMolzen) for their significant joint contributions to the development of this AI detection system, along with all other team members and contributors.



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

### Direct access to inference python script

`./ai-analyse/new_analysis/classify-v6.py --models ensemble1 --live`

Just start, and either drag and drop an image into shell, or pass an image path with surrounding `''`.

## Deployment

### Prerequisites

- Docker & Docker Compose

```bash
# Build the image
docker compose build --no-cache
# or directly
docker compose up

# Start services
docker compose up -d

# Check status
docker ps

# for the frontend
cd frontend
echo -e "POSTGRES_PASSWORD=secure_password_123\nDATABASE_URL=postgresql://root:secure_password_123@localhost:5434/local\nVITE_API_BASE_URL=http://localhost:8080/nBODY_SIZE_LIMIT=25M/nORIGIN=http://localhost:3000" > .env
npm install
npm run dev
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
curl -X POST -F "image=@your-image.jpg" http://localhost:8080/upload
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
Ensemble of 5 separate models
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

## Development Setup

### Local Development

```bash
# Setup dependencies
make setup

# Install Python requirements
cd pythonScripts && pip install -r requirements.txt

# Build Go modules
go mod tidy

# Run the server
go run cmd/server/main.go
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
2. **Deep Analysis** (~200 milliseconds): AI model inference


## Troubleshooting

### Common Issues

#### Service Won't Start
```bash
# Check Docker status
docker-compose ps

# View detailed logs
docker-compose logs image-analyzer

# Clean up Docker resources (dangerous)
docker system prune -a --volumes -f

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

### Metrics Integration

Export metrics to external systems:

```bash
# Prometheus format
curl http://localhost:8080/metrics

# JSON format for custom integrations
curl http://localhost:8080/metrics | jq '.business.ai_detection_rate'
```

## License

This project is licensed under the PolyForm-Noncommercial License. See LICENSE file for details.

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

**Built with**: Go, Python, Rust, Docker, OpenCV, scikit-image, C2PA
