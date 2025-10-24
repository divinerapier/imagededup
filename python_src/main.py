import argparse
from logging import log
import logging
from pathlib import Path
import time
from typing import Any, Dict, List, Optional

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import uvicorn

from find_duplicates import find_duplicates
from find_duplicates_to_remove import find_duplicates_to_remove

# Pydantic 模型定义
class FindDuplicatesRequest(BaseModel):
    algorithm: str
    image_dir: str

class FindDuplicatesToRemoveRequest(BaseModel):
    algorithm: str
    image_dir: str

class Response(BaseModel):
    status: str
    data: Optional[Any] = None
    message: Optional[str] = None
    elapsed: Optional[float] = None

# 创建 FastAPI 应用
app = FastAPI(title="Image Deduplication API", version="1.0.0")

def validate_algorithm(algorithm: str) -> None:
    """验证算法参数"""
    valid_algorithms = ["phasher", "cnn", "dhash", "ahash", "whash"]
    if algorithm not in valid_algorithms:
        raise HTTPException(
            status_code=400, 
            detail=f"Invalid algorithm: {algorithm}. Valid options: {valid_algorithms}"
        )

def validate_image_dir(image_dir: str) -> Path:
    """验证图像目录"""
    path = Path(image_dir)
    if not path.exists():
        raise HTTPException(
            status_code=400,
            detail=f"Image directory does not exist: {path.as_posix()}"
        )
    return path

@app.get("/health")
async def health():
    """健康检查端点"""
    return {"status": "ok", "message": "Image Deduplication API is running"}

@app.post("/find-duplicates", response_model=Response)
async def find_duplicates_endpoint(request: FindDuplicatesRequest):
    """查找重复图像"""
    start_time = time.time()
    
    try:
        # 验证参数
        validate_algorithm(request.algorithm)
        image_dir = validate_image_dir(request.image_dir)
        log(level=logging.INFO, msg=f"Finding duplicates for {request.algorithm} in {image_dir}")
        # 执行查找重复图像
        results = find_duplicates(request.algorithm, image_dir)
        
        elapsed = time.time() - start_time
        
        return Response(
            status="ok",
            data=results,
            elapsed=elapsed
        )
        
    except HTTPException:
        raise
    except Exception as e:
        elapsed = time.time() - start_time
        raise HTTPException(
            status_code=500,
            detail={
                "status": "error",
                "message": str(e),
                "elapsed": elapsed
            }
        )

@app.post("/find-duplicates-to-remove", response_model=Response)
async def find_duplicates_to_remove_endpoint(request: FindDuplicatesToRemoveRequest):
    """获取需要删除的重复文件列表"""
    start_time = time.time()
    print(f"Finding duplicates to remove for {request.algorithm} in {request.image_dir}")
    try:
        # 验证参数
        validate_algorithm(request.algorithm)
        image_dir = validate_image_dir(request.image_dir)
        
        # 执行查找需要删除的文件
        results = find_duplicates_to_remove(request.algorithm, image_dir)
        
        elapsed = time.time() - start_time
        
        return Response(
            status="ok",
            data=results,
            elapsed=elapsed
        )
        
    except HTTPException:
        raise
    except Exception as e:
        elapsed = time.time() - start_time
        raise HTTPException(
            status_code=500,
            detail={
                "status": "error",
                "message": str(e),
                "elapsed": elapsed
            }
        )

def main():
    """主函数，解析命令行参数并启动服务器"""
    parser = argparse.ArgumentParser(description="Image Deduplication API Server")
    parser.add_argument("--port", "-p", type=int, default=8000, help="Port to run the server on")
    parser.add_argument("--host", type=str, default="127.0.0.1", help="Host to bind the server to")
    parser.add_argument("--reload", action="store_true", help="Enable auto-reload for development")
    
    args = parser.parse_args()
    
    print(f"Starting Image Deduplication API server on {args.host}:{args.port}")
    
    
    uvicorn.run(
        "main:app",
        host=args.host,
        port=args.port,
        reload=args.reload,
        log_level="info"
    )

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        logging.info("KeyboardInterrupt received, shutting down server")
        pass
        