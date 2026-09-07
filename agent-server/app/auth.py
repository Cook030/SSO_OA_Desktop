import asyncio
import hashlib
import json
import secrets
from datetime import UTC, datetime, timedelta

import httpx
from cryptography.fernet import Fernet

from app.config import Settings
from app.oa import OAClient, ServiceError, request_json
from app.store import SessionRow, Store


class Auth:
    def __init__(self, settings: Settings, store: Store, oa_http: httpx.AsyncClient):
        self.settings = settings
        self.store = store
        self.oa_http = oa_http
        self.cipher = Fernet(settings.encryption_key.get_secret_value().encode())
        # Single-process deployment: stripe locks bound memory and serialize refresh rotation.
        self.locks = [asyncio.Lock() for _ in range(64)]

    async def sso(self, path: str, **kwargs):
        # Per-call client prevents Cookie jar sharing across users.
        async with httpx.AsyncClient(
            base_url=self.settings.sso_url, timeout=15, follow_redirects=False
        ) as client:
            return await request_json(client, "POST", "/api/v1/auth/" + path, **kwargs)

    def encode(self, credentials: dict) -> str:
        return self.cipher.encrypt(json.dumps(credentials).encode()).decode()

    @staticmethod
    def key(bearer: str) -> str:
        return hashlib.sha256(bearer.encode()).hexdigest()

    async def login(self, account: str, password: str):
        data = await self.sso("login", json={"account": account, "password": password})
        oa = OAClient(self.oa_http, data["accessToken"])
        actor = await oa.me()
        bearer = secrets.token_urlsafe(32)
        expiry = datetime.now(UTC) + timedelta(seconds=self.settings.session_seconds)
        with self.store.session.begin() as db:
            db.add(
                SessionRow(
                    id=self.key(bearer),
                    actor_id=actor.user_id,
                    credentials=self.encode(
                        {"accessToken": data["accessToken"], "refreshToken": data["refreshToken"]}
                    ),
                    expires_at=expiry.isoformat(),
                )
            )
        return {"token": bearer, "actor": actor.model_dump(), "expires_at": expiry.isoformat()}

    async def resolve(self, bearer: str):
        key = self.key(bearer)
        async with self.locks[int(key[:4], 16) % len(self.locks)]:
            with self.store.session() as db:
                row = db.get(SessionRow, key)
                if row is None or datetime.fromisoformat(row.expires_at) <= datetime.now(UTC):
                    raise ServiceError("请登录企业账号", 401)
                credentials = json.loads(self.cipher.decrypt(row.credentials.encode()))
                actor_id = row.actor_id
            oa = OAClient(self.oa_http, credentials["accessToken"])
            try:
                actor = await oa.me()
            except ServiceError as exc:
                if exc.status != 401:
                    raise
                data = await self.sso("refresh", json={"refreshToken": credentials["refreshToken"]})
                credentials = {"accessToken": data["accessToken"], "refreshToken": data["refreshToken"]}
                oa = OAClient(self.oa_http, credentials["accessToken"])
                # Refresh tokens rotate. Persist the replacement before an OA call that may fail.
                # actor_id remains unchanged; subsequent verification still fails closed.
                with self.store.session.begin() as db:
                    row = db.get(SessionRow, key)
                    if row is None:
                        raise ServiceError("会话已退出", 401)
                    row.credentials = self.encode(credentials)
                actor = await oa.me()
            if actor.user_id != actor_id:
                raise ServiceError("SSO 身份与会话不符", 401)
            return actor, oa

    async def logout(self, bearer: str):
        key = self.key(bearer)
        async with self.locks[int(key[:4], 16) % len(self.locks)]:
            with self.store.session.begin() as db:
                row = db.get(SessionRow, key)
                if row is None:
                    return
                credentials = json.loads(self.cipher.decrypt(row.credentials.encode()))
                db.delete(row)
            try:
                await self.sso(
                    "logout",
                    json={"refreshToken": credentials["refreshToken"]},
                    headers={"Cookie": f"mh_sso2_access_token={credentials['accessToken']}"},
                )
            except ServiceError:
                # Local session is already revoked; caller is informed of remote uncertainty.
                return {"warning": "工作台已退出；SSO 暂不可用，远端会话撤销未确认。"}
