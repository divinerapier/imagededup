#!/bin/bash

# 启动 Image Deduplication API 服务器
# 用法: ./start_server.sh [port] [host]

PORT=${1:-8000}
HOST=${2:-127.0.0.1}

echo "Starting Image Deduplication API server..."
echo "Host: $HOST"
echo "Port: $PORT"
echo "API Documentation: http://$HOST:$PORT/docs"
echo "Health Check: http://$HOST:$PORT/health"
echo ""

python main.py --host $HOST --port $PORT
