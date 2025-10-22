#!/bin/bash

echo "🐳 Testing Docker build and run..."

# 构建镜像
echo "Building Docker image..."
docker build -t image-dedup:test .

if [ $? -ne 0 ]; then
    echo "❌ Docker build failed!"
    exit 1
fi

echo "✅ Docker image built successfully!"

# 测试容器运行
echo "Testing container..."
docker run --rm image-dedup:test python -c "import imagededup; print('✅ imagededup imported successfully')"

if [ $? -eq 0 ]; then
    echo "✅ Container test passed!"
    echo ""
    echo "To run with your data:"
    echo "docker run -v \$(pwd)/tests/data:/app/data image-dedup:test"
else
    echo "❌ Container test failed!"
    exit 1
fi
