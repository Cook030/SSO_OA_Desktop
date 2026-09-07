"""Local UI verification only. No external model/SSO/OA calls. Never deploy this module."""

from contextlib import asynccontextmanager
from tempfile import TemporaryDirectory

import httpx
from cryptography.fernet import Fernet
from langgraph.checkpoint.sqlite.aio import AsyncSqliteSaver

from app.auth import Auth
from app.config import Settings
from app.main import create_app
from app.runtime import Runtime
from app.store import Base, Store
from tests.conftest import FakeOA, ScriptedModel


def create_demo():
    settings = Settings(
        _env_file=None,
        encryption_key=Fernet.generate_key().decode(),
        model_name="test-only",
        model_api_key="test-only",
    )
    app = create_app(settings)

    @asynccontextmanager
    async def lifespan(app):
        with TemporaryDirectory(prefix="oa-agent-ui-") as folder:
            store = Store(f"sqlite:///{folder}/agent.db")
            Base.metadata.create_all(store.engine)
            oa = FakeOA()
            async with httpx.AsyncClient(
                base_url="http://test-oa", transport=httpx.MockTransport(oa.handle)
            ) as client:
                auth = Auth(settings, store, client)

                async def sso(path, **kwargs):
                    return {"accessToken": "access-current", "refreshToken": "test-only"}

                auth.sso = sso
                model = ScriptedModel()
                model.flow()
                model.script.extend(
                    ["这是隔离测试会话，后续消息会保留历史。真实员工状态需要连接实际 OA 后查询。"] * 5
                )
                async with AsyncSqliteSaver.from_conn_string(f"{folder}/checkpoints.db") as saver:
                    runtime = Runtime(settings, store, auth, model, saver)
                    app.state.runtime = runtime
                    try:
                        yield
                    finally:
                        await runtime.close()
            store.engine.dispose()

    app.router.lifespan_context = lifespan
    return app
