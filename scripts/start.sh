#!/bin/bash
set -e

echo "Starting Image Analyzer with AI Detection..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Docker is not running. Please start Docker first."
    exit 1
fi

# Check if ports are available
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>/dev/null; then
    echo "Port 8080 is already in use. Stopping existing services..."
    docker-compose down 2>/dev/null || true
fi

echo "Building optimized Docker image..."
docker-compose build

echo "Starting all services..."
docker-compose up -d

echo "Waiting for services to initialize..."
sleep 20

# Comprehensive health check
echo "Running health checks..."

# Basic connectivity
if curl -f -s http://localhost:8080/health > /dev/null; then
    echo "Main service is healthy"
else
    echo "Main service health check failed"
    echo "Service logs:"
    docker-compose logs --tail=20 image-analyzer
    exit 1
fi

echo ""
echo "Image Analyzer is successfully running!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Upload images: http://localhost:8080/upload"
echo "System metrics: http://localhost:8080/metrics"
echo "Health status: http://localhost:8080/health"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Usage examples:"
echo "curl -X POST -F \"file=@your-image.jpg\" http://localhost:8080/upload"
echo "curl http://localhost:8080/health"
echo ""
echo "To stop: make docker-stop"
echo "View logs: make logs"