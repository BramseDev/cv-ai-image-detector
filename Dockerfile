FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server/main.go

FROM python:3.11-slim
RUN apt-get update && apt-get install -y \
    libgl1-mesa-glx libglib2.0-0 libsm6 libxext6 libxrender-dev \
    libgomp1 libgtk-3-0 libjpeg-dev libpng-dev git curl \
    build-essential g++ gcc python3-dev exiftool \
    || apt-get install -y \
    libgl1 libglib2.0-0 libsm6 libxext6 libxrender-dev \
    libgomp1 libgtk-3-0 libjpeg-dev libpng-dev git curl \
    build-essential g++ gcc python3-dev exiftool \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Install Rust for c2pa binary compilation
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"

WORKDIR /app

# Install Python dependencies für beide Module
COPY pythonScripts/requirements.txt ./pythonScripts/
COPY ai-analyse/requirements.txt ./ai-analyse/
RUN pip install --no-cache-dir -r pythonScripts/requirements.txt \
    && pip install --no-cache-dir -r ai-analyse/requirements.txt \
    && pip cache purge

COPY pythonScripts/ ./pythonScripts/
COPY ai-analyse/ ./ai-analyse/
COPY pkg/ ./pkg/
COPY dashboard/ ./dashboard/
COPY --from=go-builder /app/main .

# Build c2pa-rust binary
RUN cd pkg/analyzer/c2pa-rust && cargo build --release

RUN mkdir -p /app/uploads /app/logs /app/tmp \
    && chmod +x ./main \
    && chmod -R 755 ./pythonScripts/

ENV GIN_MODE=release
ENV PYTHON_PATH=/usr/local/bin/python3
ENV SCRIPTS_PATH=/app/pythonScripts

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["./main"]