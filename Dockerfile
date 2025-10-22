# 使用 Ubuntu 22.04 作为基础镜像
FROM divinerapier/uv:v2025102201

# 设置工作目录
WORKDIR /app

# 复制项目文件
COPY pyproject.toml uv.lock ./
COPY .python-version ./
COPY main.py ./

# 创建虚拟环境并安装依赖

RUN uv venv && \
    uv sync --frozen

# 设置环境变量，使用虚拟环境
ENV PATH="/app/.venv/bin:$PATH"

# 创建数据目录
RUN mkdir -p /app/data

ENTRYPOINT ["/bin/bash", "-c"]

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import imagededup; print('OK')" || exit 1
