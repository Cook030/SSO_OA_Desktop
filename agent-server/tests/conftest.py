import httpx
import pytest
from cryptography.fernet import Fernet
from langchain_core.messages import AIMessageChunk
from langgraph.checkpoint.sqlite.aio import AsyncSqliteSaver

from app.auth import Auth
from app.config import Settings
from app.main import create_app
from app.runtime import Runtime
from app.store import Base, Store


class ScriptedModel:
    """Deterministic provider boundary; the real LangGraph/checkpointer still run."""

    def __init__(self):
        self.script = []
        self.inputs = []

    def bind_tools(self, tools):
        assert any(t.name == "plan_add_user_roles" for t in tools)
        return self

    async def astream(self, messages):
        self.inputs.append(messages)
        assert self.script, "Unexpected additional model call"
        item = self.script.pop(0)
        if isinstance(item, str):
            yield AIMessageChunk(content=item)
        else:
            yield AIMessageChunk(content="", tool_calls=[item])

    def flow(self):
        self.script.extend(
            [
                {"id": "c1", "name": "search_employees", "args": {"keyword": "张三"}},
                {"id": "c2", "name": "list_roles", "args": {}},
                {"id": "c3", "name": "plan_add_user_roles", "args": {"user_id": 18, "role_ids": [2]}},
            ]
        )


class FakeOA:
    def __init__(self):
        self.assigned = [1]
        self.roles = [
            {"id": 1, "name": "员工", "code": "employee", "isBuiltin": 1, "status": 1},
            {"id": 2, "name": "运营只读", "code": "ops_read", "isBuiltin": 0, "status": 1},
        ]
        self.permissions = {1: [10], 2: [20]}
        self.calls = []
        self.write_count = 0
        self.deny_write = False
        self.timeout_write = False
        self.expired_access = False
        self.forbid_actor = False
        self.deny_readback = False
        self.refreshed_outage = False

    def respond(self, data=None, code=200):
        return httpx.Response(200, json={"code": code, "message": "do not leak upstream text", "data": data})

    def handle(self, request):
        self.calls.append((request.method, request.url.path, request.headers.get("cookie", "")))
        cookie = request.headers.get("cookie", "")
        assert "mh_sso2_access_token=" in cookie
        actor_id = 2 if "access-other" in cookie else 1
        path = request.url.path
        if path == "/api/me/permissions":
            if self.refreshed_outage and "access-refreshed" in cookie:
                raise httpx.ConnectError("OA temporarily unavailable")
            if self.expired_access and "access-current" in cookie:
                return self.respond(code=401)
            return self.respond(
                {"userId": actor_id, "isAdmin": not self.forbid_actor, "roles": [], "permissions": []}
            )
        if path == "/api/employees":
            return self.respond(
                {
                    "total": 1,
                    "list": [
                        {
                            "id": 18,
                            "name": "张三",
                            "account": "zhangsan",
                            "department": "运营部",
                            "phone": "SECRET-PHONE",
                            "roles": [],
                            "platformPermissions": [],
                        }
                    ],
                }
            )
        if path == "/api/roles":
            return self.respond(self.roles)
        if path == "/api/users/18/roles":
            if self.deny_readback and self.write_count:
                return self.respond(code=403)
            return self.respond([r for r in self.roles if r["id"] in self.assigned])
        if path.startswith("/api/roles/") and path.endswith("/permissions"):
            return self.respond(self.permissions[int(path.split("/")[3])])
        if path == "/api/roles/users":
            import json

            self.write_count += 1
            if self.deny_write:
                return self.respond(code=403)
            payload = json.loads(request.content)
            assert payload == {"userIds": [18], "roleIds": [2]}
            self.assigned = sorted(set(self.assigned + payload["roleIds"]))
            if self.timeout_write:
                raise httpx.ReadTimeout("secret internal URL")
            return self.respond({"affectedCount": 1})
        raise AssertionError(f"Unexpected OA request: {request.method} {path}")


@pytest.fixture
async def env(tmp_path):
    settings = Settings(
        _env_file=None,
        encryption_key=Fernet.generate_key().decode(),
        database_url=f"sqlite:///{tmp_path / 'app.db'}",
        model_name="test",
        model_api_key="test",
    )
    store = Store(settings.database_url)
    Base.metadata.create_all(store.engine)
    oa = FakeOA()
    async with httpx.AsyncClient(base_url="http://oa", transport=httpx.MockTransport(oa.handle)) as oa_http:
        auth = Auth(settings, store, oa_http)

        async def sso(path, **kwargs):
            if path == "login":
                token = "access-other" if kwargs["json"]["account"] == "other" else "access-current"
                return {"accessToken": token, "refreshToken": "refresh-secret"}
            if path == "refresh":
                return {"accessToken": "access-refreshed", "refreshToken": "refresh-rotated"}
            return None

        auth.sso = sso
        async with AsyncSqliteSaver.from_conn_string(str(tmp_path / "checkpoints.db")) as saver:
            model = ScriptedModel()
            rt = Runtime(settings, store, auth, model, saver)
            app = create_app(settings, rt)
            async with app.router.lifespan_context(app):
                async with httpx.AsyncClient(
                    transport=httpx.ASGITransport(app=app), base_url="http://agent"
                ) as client:
                    result = await client.post(
                        "/api/auth/login", json={"account": "admin", "password": "secret"}
                    )
                    assert result.status_code == 200
                    client.headers["Authorization"] = "Bearer " + result.json()["token"]
                    yield {
                        "client": client,
                        "runtime": rt,
                        "store": store,
                        "oa": oa,
                        "model": model,
                        "auth": auth,
                        "tmp_path": tmp_path,
                    }
    store.engine.dispose()


async def run_to_plan(env):
    client = env["client"]
    env["model"].flow()
    conversation = (await client.post("/api/conversations", json={})).json()
    result = await client.post(
        f"/api/conversations/{conversation['id']}/messages", json={"content": "给张三增加运营只读角色"}
    )
    assert result.status_code == 202
    task_id = result.json()["id"]
    await env["runtime"].jobs[task_id]
    task = (await client.get(f"/api/tasks/{task_id}")).json()
    assert task["status"] == "waiting_confirmation", task
    assert env["oa"].write_count == 0
    return conversation["id"], task_id, task["plans"][0]


async def decide(env, plan, approved=True):
    result = await env["client"].post(
        f"/api/operation-plans/{plan['id']}/decision", json={"approved": approved}
    )
    assert result.status_code == 202, result.text
    task_id = result.json()["task_id"]
    await env["runtime"].jobs[task_id]
    return (await env["client"].get(f"/api/tasks/{task_id}")).json()
