#!/bin/bash
echo "Testing Image Analyzer System..."

# Test health endpoint
echo "1. Testing health endpoint..."
health_response=$(curl -s http://localhost:8080/health)
if echo "$health_response" | jq -e '.healthy == true' > /dev/null 2>&1; then
    echo "Health check passed"
else
    echo "Health check failed"
    echo "   Response: $health_response"
fi

# Test metrics endpoint
echo "2. Testing metrics endpoint..."
metrics_response=$(curl -s http://localhost:8080/metrics)
if echo "$metrics_response" | jq -e '.metrics' > /dev/null 2>&1; then
    echo "Metrics endpoint working"
else
    echo "  Metrics endpoint failed"
fi

echo "🏁 System test completed!"