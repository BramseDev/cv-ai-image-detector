# Stage 1: Go Builder
FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY pkg/ ./pkg/
COPY internal/ ./internal/
COPY dashboard/ ./dashboard/
COPY logging/ ./logging/
COPY cache/ ./cache/
COPY monitoring/ ./monitoring/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server/main.go

# Stage 2: Rust Builder (Updated to latest version with build tools)
FROM rust:1.82-slim AS rust-builder
RUN apt-get update && apt-get install -y \
    perl \
    make \
    gcc \
    g++ \
    libc6-dev \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY pkg/analyzer/c2pa-rust/ ./
RUN cargo build --release && strip target/release/c2pa-rust

# Stage 3: Final image
FROM python:3.11-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl git exiftool build-essential python3-dev libgl1 libglib2.0-0 && \
    apt-get clean && rm -rf /var/lib/apt/lists/* /var/cache/apt/archives/*

WORKDIR /app

# Copy requirements first for better caching
COPY pythonScripts/requirements.txt ./pythonScripts/
COPY ai-analyse/requirements.txt ./ai-analyse/
RUN pip install --no-cache-dir -r pythonScripts/requirements.txt && \
    pip install --no-cache-dir -r ai-analyse/requirements.txt && \
    pip cache purge

# Copy application files
COPY pythonScripts/ ./pythonScripts/
COPY ai-analyse/ ./ai-analyse/
COPY dashboard/ ./dashboard/
COPY --from=go-builder /app/main .
COPY --from=rust-builder /app/target/release/c2pa-rust ./pkg/analyzer/c2pa-rust/target/release/

RUN chmod +x ./main && mkdir -p uploads logs tmp

EXPOSE 8080
CMD ["./main"]