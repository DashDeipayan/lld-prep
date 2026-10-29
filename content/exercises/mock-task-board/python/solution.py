from __future__ import annotations

from enum import Enum
from typing import List


# ---- Provided ----
class Status(Enum):
    TODO = "TODO"
    IN_PROGRESS = "IN_PROGRESS"
    IN_REVIEW = "IN_REVIEW"
    DONE = "DONE"


class UnauthorizedError(Exception):
    """The actor isn't allowed to make this transition."""


class WorkflowError(Exception):
    """Illegal transition, blocked start, dependency cycle, or a change to a DONE task."""


# ---- Your design ----
class TaskBoard:
    def create(self, title: str, reporter: str, priority: int) -> str:
        raise NotImplementedError("TODO")

    def assign(self, task_id: str, user: str) -> None:
        raise NotImplementedError("TODO")

    def add_dependency(self, task_id: str, depends_on: str) -> None:
        raise NotImplementedError("TODO")

    def move(self, task_id: str, to: Status, actor: str) -> None:
        raise NotImplementedError("TODO")

    def status(self, task_id: str) -> Status:
        raise NotImplementedError("TODO")

    def tasks_for(self, user: str) -> List[str]:
        raise NotImplementedError("TODO")

    def blocked_tasks(self) -> List[str]:
        raise NotImplementedError("TODO")

    def history(self, task_id: str) -> List[str]:
        raise NotImplementedError("TODO")
