FROM python:3.12-slim
COPY --from=ghcr.io/astral-sh/uv:0.11.16 /uv /usr/local/bin/uv
WORKDIR /app
COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-dev
COPY app ./app
RUN useradd --create-home agent && chown -R agent:agent /app
USER agent
EXPOSE 8090
CMD ["uv", "run", "--no-sync", "uvicorn", "app.main:create_app", "--factory", "--host", "0.0.0.0", "--port", "8090", "--workers", "1"]
