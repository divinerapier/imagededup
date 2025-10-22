#!/bin/bash

# 构建 Docker 镜像
echo "Building Docker image..."
docker build -t image-dedup:latest .

# 检查构建是否成功
if [ $? -eq 0 ]; then
    echo "✅ Docker image built successfully!"
    echo ""
    echo "To run the container:"
    echo "docker run -v /path/to/your/images:/app/data image-dedup:latest"
    echo ""
    echo "Example:"
    echo "docker run -v \$(pwd)/tests/data:/app/data image-dedup:latest"
else
    echo "❌ Docker build failed!"
    exit 1
fi
