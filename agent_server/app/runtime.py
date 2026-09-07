import asyncio
import hashlib
import logging

from langchain_core.messages import AIMessage, HumanMessage, ToolMessage
from langgraph.types import Command
from sqlalchemy import select

from app.graph import AgentGraph
from app.oa import ServiceError
from app.store import Plan

log = logging.getLogger(__name__)


class Runtime:
    def __init__(self, settings, store, auth, model, checkpointer):
        self.settings, self.store, self.auth = settings, store, auth
        self.model, self.checkpointer = model, checkpointer
        self.jobs: dict[str, asyncio.Task] = {}
        self.locks = [asyncio.Lock() for _ in range(64)]
        self.slots = asyncio.Semaphore(settings.max_concurrent_tasks)

    def lock(self, conversation_id: str):
        return self.locks[int(hashlib.sha256(conversation_id.encode()).hexdigest()[:8], 16) % len(self.locks)]

    def launch(
        self,
        task_id: str,
        conversation_id: str,
        bearer: str,
        content: str | None = None,
        decision: dict | None = None,
    ):
        if task_id in self.jobs:
            raise ServiceError("任务已在执行", 409)
        job = asyncio.create_task(self.run(task_id, conversation_id, bearer, content, decision))
        self.jobs[task_id] = job
        job.add_done_callback(lambda _: self.jobs.pop(task_id, None))

    async def run(self, task_id, conversation_id, bearer, content, decision):
        try:
            async with self.slots, asyncio.timeout(self.settings.task_timeout):
                self.store.set_task(task_id, "running")
                self.store.emit(task_id, "task.running", {})
                agent = AgentGraph(self, task_id, conversation_id, bearer)
                graph = agent.compile()
                config = {
                    "configurable": {"thread_id": conversation_id},
                    "recursion_limit": self.settings.max_steps * 3 + 10,
                }
                recovered_messages = []
                if decision is None:
                    snapshot = await graph.aget_state(config)
                    history = snapshot.values.get("messages", [])
                    if history and isinstance(history[-1], AIMessage) and history[-1].tool_calls:
                        # A cancelled read node can leave unpaired tool calls in a checkpoint.
                        recovered_messages = [
                            ToolMessage(
                                content="上次任务已中断，此工具调用没有确认结果。", tool_call_id=c["id"]
                            )
                            for c in history[-1].tool_calls
                        ]
                data = (
                    Command(resume=decision)
                    if decision is not None
                    else {
                        "messages": [*recovered_messages, HumanMessage(content=content)],
                        "steps": 0,
                        "plan_id": None,
                        "approved": False,
                    }
                )
                result = await graph.ainvoke(data, config=config)
                if result.get("__interrupt__"):
                    preview = result["__interrupt__"][0].value["plan"]
                    employee = preview["payload"]["employee"]
                    agent.save_answer(
                        f"已为 {employee['name']}（{employee['account']}）生成角色变更预览。请在任务详情中核对并确认，当前尚未修改 OA。"
                    )
                    self.store.set_task(task_id, "waiting_confirmation")
                    self.store.emit(
                        task_id, "operation.confirmation_required", result["__interrupt__"][0].value
                    )
                else:
                    with self.store.session() as db:
                        plans = list(db.scalars(select(Plan).where(Plan.task_id == task_id)))
                    problematic = any(p.status in ("failed", "unknown", "conflict", "expired") for p in plans)
                    rejected = any(p.status == "rejected" for p in plans)
                    status = "failed" if problematic else "cancelled" if rejected else "succeeded"
                    self.store.set_task(task_id, status)
                    self.store.emit(task_id, "task.completed", {"status": status})
        except asyncio.CancelledError:
            self.fail(task_id, "任务已停止；已发出的 OA 请求可能完成，请核查当前状态。", "cancelled")
        except (ServiceError, TimeoutError) as exc:
            self.fail(
                task_id, str(exc) if isinstance(exc, ServiceError) else "任务超时，请检查 OA 状态后重试。"
            )
        except Exception as exc:
            # Provider exceptions can include request bodies. Log only the exception type.
            log.error("task_failed task=%s type=%s", task_id, type(exc).__name__)
            self.fail(task_id, "任务执行失败，请检查模型配置、服务状态及服务端日志。")

    def fail(self, task_id, message, status="failed"):
        with self.store.session.begin() as db:
            for plan in db.scalars(select(Plan).where(Plan.task_id == task_id)):
                if plan.status == "executing":
                    plan.status = "unknown"
                elif plan.status in ("pending", "approved"):
                    plan.status = "cancelled"
        self.store.set_task(task_id, status, message)
        self.store.emit(task_id, "task.failed", {"message": message, "status": status})

    async def close(self):
        jobs = list(self.jobs.values())
        for job in jobs:
            job.cancel()
        await asyncio.gather(*jobs, return_exceptions=True)
