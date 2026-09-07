"""Small transactional application store. Graph checkpoints use LangGraph's own saver."""

from datetime import UTC, datetime
from uuid import uuid4

from sqlalchemy import JSON, BigInteger, ForeignKey, Integer, String, Text, create_engine, select, update
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, sessionmaker


def uid() -> str:
    return str(uuid4())


def now() -> str:
    return datetime.now(UTC).isoformat()


class Base(DeclarativeBase):
    pass


class SessionRow(Base):
    __tablename__ = "agent_sessions"
    id: Mapped[str] = mapped_column(String(64), primary_key=True)  # SHA-256 of opaque bearer
    actor_id: Mapped[int] = mapped_column(BigInteger, index=True)
    credentials: Mapped[str] = mapped_column(Text)  # encrypted SSO access/refresh tokens
    expires_at: Mapped[str] = mapped_column(String(40))


class Conversation(Base):
    __tablename__ = "conversations"
    id: Mapped[str] = mapped_column(String(36), primary_key=True, default=uid)
    actor_id: Mapped[int] = mapped_column(BigInteger, index=True)
    title: Mapped[str] = mapped_column(String(100), default="新会话")
    created_at: Mapped[str] = mapped_column(String(40), default=now)


class Task(Base):
    __tablename__ = "tasks"
    id: Mapped[str] = mapped_column(String(36), primary_key=True, default=uid)
    conversation_id: Mapped[str] = mapped_column(ForeignKey("conversations.id"), index=True)
    actor_id: Mapped[int] = mapped_column(BigInteger, index=True)
    status: Mapped[str] = mapped_column(String(32), default="queued")
    error: Mapped[str | None] = mapped_column(Text)
    created_at: Mapped[str] = mapped_column(String(40), default=now)


class Message(Base):
    __tablename__ = "messages"
    id: Mapped[str] = mapped_column(String(36), primary_key=True, default=uid)
    conversation_id: Mapped[str] = mapped_column(ForeignKey("conversations.id"), index=True)
    role: Mapped[str] = mapped_column(String(16))
    content: Mapped[str] = mapped_column(Text)
    created_at: Mapped[str] = mapped_column(String(40), default=now)


class Plan(Base):
    __tablename__ = "operation_plans"
    id: Mapped[str] = mapped_column(String(36), primary_key=True, default=uid)
    task_id: Mapped[str] = mapped_column(ForeignKey("tasks.id"), index=True)
    actor_id: Mapped[int] = mapped_column(BigInteger)
    tool_call_id: Mapped[str] = mapped_column(String(200), unique=True)
    payload: Mapped[dict] = mapped_column(JSON)
    status: Mapped[str] = mapped_column(String(32), default="pending")
    expires_at: Mapped[str] = mapped_column(String(40))
    result: Mapped[dict | None] = mapped_column(JSON)


class Event(Base):
    __tablename__ = "task_events"
    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    task_id: Mapped[str] = mapped_column(ForeignKey("tasks.id"), index=True)
    kind: Mapped[str] = mapped_column(String(64))
    payload: Mapped[dict] = mapped_column(JSON)
    created_at: Mapped[str] = mapped_column(String(40), default=now)


class Audit(Base):
    __tablename__ = "agent_audit_logs"
    id: Mapped[str] = mapped_column(String(36), primary_key=True, default=uid)
    task_id: Mapped[str] = mapped_column(String(36), index=True)
    actor_id: Mapped[int] = mapped_column(BigInteger)
    action: Mapped[str] = mapped_column(String(80))
    details: Mapped[dict] = mapped_column(JSON)
    created_at: Mapped[str] = mapped_column(String(40), default=now)


def as_dict(row) -> dict:
    return {column.name: getattr(row, column.name) for column in row.__table__.columns}


class Store:
    def __init__(self, url: str):
        self.engine = create_engine(url, pool_pre_ping=True)
        self.session = sessionmaker(self.engine, expire_on_commit=False)

    def emit(self, task_id: str, kind: str, payload: dict):
        with self.session.begin() as db:
            db.add(Event(task_id=task_id, kind=kind, payload=payload))

    def set_task(self, task_id: str, status: str, error: str | None = None):
        with self.session.begin() as db:
            db.execute(update(Task).where(Task.id == task_id).values(status=status, error=error))

    def audit(self, task_id: str, actor_id: int, action: str, details: dict):
        with self.session.begin() as db:
            db.add(Audit(task_id=task_id, actor_id=actor_id, action=action, details=details))

    def recover_interrupted_runs(self):
        # No automatic write replay: the OA may have committed before this process stopped.
        with self.session.begin() as db:
            for task in db.scalars(select(Task).where(Task.status.in_(["queued", "running"]))):
                task.status = "failed"
                task.error = "服务重启中断了任务，请核查 OA 当前状态后重新提交。"
                db.add(Event(task_id=task.id, kind="task.failed", payload={"message": task.error}))
                db.execute(
                    update(Plan)
                    .where(Plan.task_id == task.id, Plan.status.in_(["pending", "approved"]))
                    .values(status="cancelled")
                )
            db.execute(update(Plan).where(Plan.status == "executing").values(status="unknown"))
