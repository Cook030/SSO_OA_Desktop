from pydantic import SecretStr
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_prefix="AGENT_", extra="ignore")

    database_url: str = "sqlite:///./data/agent.db"
    checkpoint_url: str = "./data/checkpoints.db"
    encryption_key: SecretStr  # Fernet key, stable across restarts; no insecure default.
    oa_url: str = "http://127.0.0.1:8080"
    sso_url: str = "http://127.0.0.1:8081"
    cors_origins: list[str] = [
        "http://localhost:5173",
        "http://127.0.0.1:5173",
        "http://wails.localhost",
        "wails://wails",
    ]
    model_base_url: str = "https://api.openai.com/v1"
    model_api_key: SecretStr = SecretStr("")
    model_name: str = ""
    model_timeout: float = 60
    max_steps: int = 16
    task_timeout: float = 300
    session_seconds: int = 28800
    plan_seconds: int = 600
    max_concurrent_tasks: int = 8
