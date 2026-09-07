import asyncio
import json
from contextlib import AsyncExitStack, asynccontextmanager
from datetime import UTC, datetime
from pathlib import Path

import httpx
from fastapi import Depends, FastAPI, Header, HTTPException, Request
from fastapi.exceptions import RequestValidationError
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse, StreamingResponse
from langchain_openai import ChatOpenAI
from langgraph.checkpoint.postgres.aio import AsyncPostgresSaver
from langgraph.checkpoint.sqlite.aio import AsyncSqliteSaver
from pydantic import BaseModel, ConfigDict, Field, SecretStr
from sqlalchemy import select

from app.auth import Auth
from app.config import Settings
from app.lease import single_instance
from app.oa import ServiceError
from app.runtime import Runtime
from app.store import Audit, Base, Conversation, Event, Message, Plan, Store, Task, as_dict


class LoginInput(BaseModel):
    account: str = Field(min_length=1, max_length=100)
    password: SecretStr = Field(min_length=1, max_length=512)


class MessageInput(BaseModel):
    content: str = Field(min_length=1, max_length=8000)


class DecisionInput(BaseModel):
    model_config = ConfigDict(extra="forbid")
    approved: bool


def create_app(settings: Settings | None = None, supplied_runtime: Runtime | None = None) -> FastAPI:
    @asynccontextmanager
    async def lifespan(app):
        if supplied_runtime:
            app.state.runtime = supplied_runtime
            try:
                yield
            finally:
                await supplied_runtime.close()
            return
        cfg = settings or Settings()
        if cfg.database_url.startswith("sqlite"):
            Path("data").mkdir(exist_ok=True)
        store = Store(cfg.database_url)
        async with AsyncExitStack() as stack:
            stack.enter_context(single_instance(store.engine))
            # Idempotent: existing tables are left untouched, so restarts keep data.
            Base.metadata.create_all(store.engine)
            if cfg.checkpoint_url.startswith("postgres"):
                saver = await stack.enter_async_context(
                    AsyncPostgresSaver.from_conn_string(cfg.checkpoint_url)
                )
                await saver.setup()
            else:
                Path(cfg.checkpoint_url).parent.mkdir(parents=True, exist_ok=True)
                saver = await stack.enter_async_context(AsyncSqliteSaver.from_conn_string(cfg.checkpoint_url))
                await saver.setup()
            oa_http = await stack.enter_async_context(
                httpx.AsyncClient(base_url=cfg.oa_url, timeout=15, follow_redirects=False)
            )
            auth = Auth(cfg, store, oa_http)
            model = ChatOpenAI(
                model=cfg.model_name or "not-configured",
                api_key=cfg.model_api_key.get_secret_value() or "not-configured",
                base_url=cfg.model_base_url,
                timeout=cfg.model_timeout,
                max_retries=0,
            )
            runtime = Runtime(cfg, store, auth, model, saver)
            app.state.runtime = runtime
            store.recover_interrupted_runs()
            try:
                yield
            finally:
                await runtime.close()
                store.engine.dispose()

    app = FastAPI(title="企业 OA Agent", version="0.1.0", lifespan=lifespan)
    cfg = settings or (supplied_runtime.settings if supplied_runtime else None)
    # Production loads .env on startup, including exact allowed browser origins.
    origins = cfg.cors_origins if cfg else Settings().cors_origins
    app.add_middleware(
        CORSMiddleware,
        allow_origins=origins,
        allow_credentials=False,
        allow_methods=["GET", "POST"],
        allow_headers=["Authorization", "Content-Type", "Last-Event-ID"],
    )

    @app.exception_handler(ServiceError)
    async def service_error(request, exc):
        return JSONResponse({"detail": str(exc)}, status_code=exc.status)

    @app.exception_handler(RequestValidationError)
    async def validation_error(request, exc):
        # Default validation errors echo input, which may contain a login password.
        return JSONResponse({"detail": "请求参数格式无效"}, status_code=422)

    @app.exception_handler(Exception)
    async def internal_error(request, exc):
        return JSONResponse({"detail": "服务内部错误，请检查服务端运行状态。"}, status_code=500)

    def runtime(request: Request) -> Runtime:
        return request.app.state.runtime

    async def identity(rt=Depends(runtime), authorization: str = Header(default="")):
        scheme, _, bearer = authorization.partition(" ")
        if scheme.lower() != "bearer" or not bearer or len(bearer) > 256:
            raise HTTPException(401, "请先登录")
        actor, _ = await rt.auth.resolve(bearer)
        return actor, bearer

    def own(db, cls, row_id, actor):
        row = db.get(cls, row_id)
        if row is None or row.actor_id != actor.user_id:
            raise HTTPException(404, "记录不存在")
        return row

    @app.get("/healthz")
    async def health():
        return {"status": "ok"}

    @app.post("/api/auth/login")
    async def login(data: LoginInput, rt=Depends(runtime)):
        return await rt.auth.login(data.account.strip(), data.password.get_secret_value())

    @app.post("/api/auth/logout")
    async def logout(rt=Depends(runtime), authorization: str = Header(default="")):
        scheme, _, bearer = authorization.partition(" ")
        if scheme.lower() != "bearer" or not bearer:
            raise HTTPException(401, "请先登录")
        return await rt.auth.logout(bearer) or {"ok": True}

    @app.get("/api/me")
    async def me(who=Depends(identity), rt=Depends(runtime)):
        return {
            "actor": who[0].model_dump(),
            "model_configured": bool(rt.settings.model_name and rt.settings.model_api_key.get_secret_value()),
        }

    @app.post("/api/conversations", status_code=201)
    async def new_conversation(who=Depends(identity), rt=Depends(runtime)):
        with rt.store.session.begin() as db:
            row = Conversation(actor_id=who[0].user_id)
            db.add(row)
            db.flush()
            return as_dict(row)

    @app.get("/api/conversations")
    async def conversations(who=Depends(identity), rt=Depends(runtime)):
        with rt.store.session() as db:
            return [
                as_dict(x)
                for x in db.scalars(
                    select(Conversation)
                    .where(Conversation.actor_id == who[0].user_id)
                    .order_by(Conversation.created_at.desc())
                    .limit(100)
                )
            ]

    @app.get("/api/conversations/{conversation_id}")
    async def conversation(conversation_id: str, who=Depends(identity), rt=Depends(runtime)):
        with rt.store.session() as db:
            row = own(db, Conversation, conversation_id, who[0])
            messages = db.scalars(
                select(Message)
                .where(Message.conversation_id == row.id)
                .order_by(Message.created_at, Message.id)
            )
            tasks = db.scalars(select(Task).where(Task.conversation_id == row.id).order_by(Task.created_at))
            return {
                **as_dict(row),
                "messages": [as_dict(x) for x in messages],
                "tasks": [as_dict(x) for x in tasks],
            }

    @app.post("/api/conversations/{conversation_id}/messages", status_code=202)
    async def send(conversation_id: str, data: MessageInput, who=Depends(identity), rt=Depends(runtime)):
        if not data.content.strip():
            raise HTTPException(422, "请输入任务")
        if not rt.settings.model_name or not rt.settings.model_api_key.get_secret_value():
            raise HTTPException(503, "管理员尚未配置模型服务")
        async with rt.lock(conversation_id):
            with rt.store.session.begin() as db:
                row = own(db, Conversation, conversation_id, who[0])
                active = db.scalar(
                    select(Task).where(
                        Task.conversation_id == row.id,
                        Task.status.in_(["queued", "running", "waiting_confirmation"]),
                    )
                )
                if active:
                    raise HTTPException(409, "请先完成或取消当前任务")
                if row.title == "新会话":
                    row.title = data.content.strip()[:40]
                task = Task(conversation_id=row.id, actor_id=who[0].user_id)
                db.add(task)
                db.add(Message(conversation_id=row.id, role="user", content=data.content.strip()))
                db.flush()
                result = as_dict(task)
            rt.launch(task.id, conversation_id, who[1], content=data.content.strip())
            return result

    @app.get("/api/tasks/{task_id}")
    async def task_detail(task_id: str, who=Depends(identity), rt=Depends(runtime)):
        with rt.store.session() as db:
            task = own(db, Task, task_id, who[0])
            plans = db.scalars(select(Plan).where(Plan.task_id == task.id))
            audits = db.scalars(select(Audit).where(Audit.task_id == task.id).order_by(Audit.created_at))
            return {
                **as_dict(task),
                "plans": [as_dict(x) for x in plans],
                "audit": [as_dict(x) for x in audits],
            }

    @app.post("/api/operation-plans/{plan_id}/decision", status_code=202)
    async def decision(plan_id: str, data: DecisionInput, who=Depends(identity), rt=Depends(runtime)):
        with rt.store.session() as db:
            plan = own(db, Plan, plan_id, who[0])
            task = own(db, Task, plan.task_id, who[0])
            conversation_id = task.conversation_id
        async with rt.lock(conversation_id):
            with rt.store.session.begin() as db:
                plan = own(db, Plan, plan_id, who[0])
                task = own(db, Task, plan.task_id, who[0])
                if task.status != "waiting_confirmation" or plan.status != "pending" or task.id in rt.jobs:
                    raise HTTPException(409, "该计划已处理或任务尚未就绪")
                if data.approved and datetime.fromisoformat(plan.expires_at) <= datetime.now(UTC):
                    raise HTTPException(409, "计划已过期，请取消任务后重新生成")
                plan.status = "approved" if data.approved else "rejected"
                task.status = "queued"
                task_id = task.id
            rt.store.audit(
                task_id,
                who[0].user_id,
                "plan.confirmed" if data.approved else "plan.rejected",
                {"plan_id": plan_id},
            )
            rt.launch(
                task_id, conversation_id, who[1], decision={"plan_id": plan_id, "approved": data.approved}
            )
            return {"task_id": task_id}

    @app.post("/api/tasks/{task_id}/cancel")
    async def cancel(task_id: str, who=Depends(identity), rt=Depends(runtime)):
        with rt.store.session() as db:
            task = own(db, Task, task_id, who[0])
            conversation_id = task.conversation_id
        async with rt.lock(conversation_id):
            with rt.store.session() as db:
                task = own(db, Task, task_id, who[0])
                if task.status not in ("queued", "running", "waiting_confirmation"):
                    return {"status": task.status}
            if task_id in rt.jobs:
                rt.jobs[task_id].cancel()
                await rt.jobs[task_id]
            else:
                rt.fail(task_id, "用户取消了任务。", "cancelled")
            return {"status": "cancelled"}

    @app.get("/api/tasks/{task_id}/events")
    async def events(
        task_id: str,
        request: Request,
        after: int = 0,
        last_event_id: str = Header(default=""),
        who=Depends(identity),
        rt=Depends(runtime),
    ):
        with rt.store.session() as db:
            own(db, Task, task_id, who[0])
        try:
            cursor = max(0, after, int(last_event_id or "0"))
        except ValueError as exc:
            raise HTTPException(400, "事件游标无效") from exc

        async def stream():
            nonlocal cursor
            # Bound each stream lifetime so authorization is revalidated on reconnect.
            deadline = asyncio.get_running_loop().time() + 60
            while not await request.is_disconnected():
                with rt.store.session() as db:
                    batch = list(
                        db.scalars(
                            select(Event)
                            .where(Event.task_id == task_id, Event.id > cursor)
                            .order_by(Event.id)
                            .limit(200)
                        )
                    )
                    task = db.get(Task, task_id)
                    status = task.status
                for event in batch:
                    cursor = event.id
                    yield f"id: {event.id}\nevent: {event.kind}\ndata: {json.dumps(event.payload, ensure_ascii=False)}\n\n"
                if len(batch) == 200:
                    continue
                if status not in ("queued", "running") or asyncio.get_running_loop().time() >= deadline:
                    break
                if not batch:
                    yield ": heartbeat\n\n"
                await asyncio.sleep(0.5)

        return StreamingResponse(
            stream(),
            media_type="text/event-stream",
            headers={"Cache-Control": "no-cache", "X-Accel-Buffering": "no"},
        )

    return app
