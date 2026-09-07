"""Fail fast if two servers attempt to recover or run tasks against the same DB."""

import os
from contextlib import contextmanager
from pathlib import Path

from sqlalchemy import text


@contextmanager
def single_instance(engine):
    if engine.dialect.name == "postgresql":
        with engine.connect() as connection:
            key = 783429105213
            if not connection.scalar(text("SELECT pg_try_advisory_lock(:key)"), {"key": key}):
                raise RuntimeError("Only one Agent server may use this database; use --workers 1")
            try:
                yield
            finally:
                connection.execute(text("SELECT pg_advisory_unlock(:key)"), {"key": key})
        return
    path = Path(engine.url.database).resolve().with_suffix(".worker.lock")
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a+b") as lock:
        lock.seek(0)
        if os.name == "nt":
            import msvcrt

            try:
                msvcrt.locking(lock.fileno(), msvcrt.LK_NBLCK, 1)
            except OSError as exc:
                raise RuntimeError("Only one Agent server may use this database; use --workers 1") from exc
        else:
            import fcntl

            try:
                fcntl.flock(lock.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
            except OSError as exc:
                raise RuntimeError("Only one Agent server may use this database; use --workers 1") from exc
        try:
            yield
        finally:
            lock.seek(0)
            if os.name == "nt":
                msvcrt.locking(lock.fileno(), msvcrt.LK_UNLCK, 1)
            else:
                fcntl.flock(lock.fileno(), fcntl.LOCK_UN)
