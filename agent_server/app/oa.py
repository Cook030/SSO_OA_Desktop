"""Typed, allow-listed OA access. Tokens never enter model inputs or checkpoints."""

from typing import Any

import httpx
from pydantic import BaseModel, ConfigDict, Field


class ServiceError(Exception):
    def __init__(self, message: str, status: int = 502):
        super().__init__(message)
        self.status = status


class DTO(BaseModel):
    model_config = ConfigDict(populate_by_name=True)


class Actor(DTO):
    user_id: int = Field(alias="userId", gt=0)
    is_admin: bool = Field(alias="isAdmin", default=False)
    permissions: list[str] = Field(default_factory=list)
    roles: list[str] = Field(default_factory=list)

    def require(self, code: str):
        if not self.is_admin and code not in self.permissions:
            raise ServiceError("当前账号没有该操作权限", 403)


class Role(DTO):
    id: int
    name: str
    code: str
    is_builtin: int = Field(alias="isBuiltin", default=0)
    status: int = 1


class Employee(DTO):
    id: int
    name: str
    account: str
    department: str = ""
    roles: list[Role] = Field(default_factory=list)
    platform_permissions: list[dict] = Field(alias="platformPermissions", default_factory=list)
    # Phone/email are deliberately excluded from model input.


async def request_json(client: httpx.AsyncClient, method: str, path: str, **kwargs) -> Any:
    try:
        response = await client.request(method, path, **kwargs)
        response.raise_for_status()
        body = response.json()
        code = body.get("code")
        if code != 200:
            status = code if code in (400, 401, 403, 404, 409, 429) else 502
            # Never echo upstream bodies, cookies, URLs, or credentials to model/logs.
            messages = {
                401: "登录已过期，请重新登录",
                403: "OA 拒绝了当前账号的操作",
                400: "OA 拒绝了操作参数",
                429: "请求过于频繁，请稍后重试",
            }
            raise ServiceError(messages.get(status, "上游服务未完成请求"), status)
        return body["data"]
    except httpx.HTTPStatusError as exc:
        if exc.response.status_code in (400, 401, 403, 404, 409, 429):
            status = exc.response.status_code
            raise ServiceError(
                "登录已过期，请重新登录" if status == 401 else "上游服务拒绝了请求", status
            ) from exc
        raise ServiceError("上游服务不可用") from exc
    except (httpx.HTTPError, ValueError, KeyError, AttributeError) as exc:
        raise ServiceError("上游服务不可用或响应格式异常") from exc


class OAClient:
    def __init__(self, client: httpx.AsyncClient, token: str):
        self.client = client
        self.token = token

    async def call(self, method: str, path: str, **kwargs):
        return await request_json(
            self.client, method, path, headers={"Cookie": f"mh_sso2_access_token={self.token}"}, **kwargs
        )

    async def me(self) -> Actor:
        return Actor.model_validate(await self.call("GET", "/api/me/permissions"))

    async def search(self, keyword: str, page: int = 1) -> dict:
        data = await self.call(
            "GET", "/api/employees", params={"keyword": keyword, "page": page, "pageSize": 20}
        )
        return {
            "total": data["total"],
            "page": page,
            "items": [Employee.model_validate(x).model_dump() for x in data["list"]],
        }

    async def roles(self) -> list[dict]:
        return [Role.model_validate(x).model_dump() for x in await self.call("GET", "/api/roles")]

    async def user_roles(self, user_id: int) -> list[dict]:
        return [
            Role.model_validate(x).model_dump() for x in await self.call("GET", f"/api/users/{user_id}/roles")
        ]

    async def add_roles(self, user_id: int, role_ids: list[int]):
        # Existing OA endpoint is incremental, transactional, and skips existing assignments.
        return await self.call("POST", "/api/roles/users", json={"userIds": [user_id], "roleIds": role_ids})
