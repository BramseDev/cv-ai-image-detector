.PHONY: docker-build docker-run docker-stop docker-clean setup test deploy logs

deploy: docker-build docker-run
	@echo "Waiting for service to start..."
	@sleep 15
	@echo "Image Analyzer is running!"
	@echo "Service: http://localhost:8080"

docker-build:
	@echo "Building Docker image..."
	docker-compose build --no-cache

docker-run:
	@echo "Starting services..."
	docker-compose up -d

docker-stop:
	@echo "Stopping services..."
	docker-compose down

test-health:
	@echo "Testing health endpoint..."
	@curl -f http://localhost:8080/health

logs:
	docker-compose logs -f image-analyzer

status:
	@docker-compose ps
