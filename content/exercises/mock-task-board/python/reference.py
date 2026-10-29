from __future__ import annotations

import threading
from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, List, Optional, Tuple


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


# ---- Model ----
ASSIGNEE, NOT_ASSIGNEE = "assignee", "not assignee"

# (from, to) -> who may make the move. Anything not listed is illegal.
RULES: Dict[Tuple[Status, Status], str] = {
    (Status.TODO, Status.IN_PROGRESS): ASSIGNEE,
    (Status.IN_PROGRESS, Status.IN_REVIEW): ASSIGNEE,
    (Status.IN_REVIEW, Status.DONE): NOT_ASSIGNEE,
    (Status.IN_REVIEW, Status.IN_PROGRESS): NOT_ASSIGNEE,
}


@dataclass
class _Task:
    id: str
    title: str
    priority: int
    seq: int
    assignee: Optional[str] = None
    status: Status = Status.TODO
    depends_on: List[str] = field(default_factory=list)
    history: List[str] = field(default_factory=list)


class TaskBoard:
    def __init__(self) -> None:
        self._tasks: Dict[str, _Task] = {}  # insertion order = creation order
        self._next = 1
        self._lock = threading.Lock()

    def create(self, title: str, reporter: str, priority: int) -> str:
        if not 1 <= priority <= 5:
            raise ValueError("priority must be 1..5")
        with self._lock:
            t = _Task("T-%d" % self._next, title, priority, self._next)
            self._next += 1
            t.history.append("created by " + reporter)
            self._tasks[t.id] = t
            return t.id

    def assign(self, task_id: str, user: str) -> None:
        with self._lock:
            t = self._task(task_id)
            if t.status is Status.DONE:
                raise WorkflowError("task is done")
            t.assignee = user
            t.history.append("assigned to " + user)

    def add_dependency(self, task_id: str, depends_on: str) -> None:
        with self._lock:
            t = self._task(task_id)
            self._task(depends_on)
            if task_id == depends_on or self._reachable(depends_on, task_id):
                raise WorkflowError("dependency would create a cycle")
            if depends_on not in t.depends_on:
                t.depends_on.append(depends_on)

    def _reachable(self, start: str, target: str) -> bool:
        """Is `target` reachable from `start` following depends-on edges? (iterative DFS)"""
        stack, seen = [start], set()
        while stack:
            cur = stack.pop()
            if cur == target:
                return True
            if cur not in seen:
                seen.add(cur)
                stack.extend(self._tasks[cur].depends_on)
        return False

    def move(self, task_id: str, to: Status, actor: str) -> None:
        with self._lock:
            t = self._task(task_id)
            rule = RULES.get((t.status, to))
            if rule is None:
                raise WorkflowError("%s -> %s is not allowed" % (t.status.value, to.value))
            if t.assignee is None:
                raise WorkflowError("task is unassigned")
            is_assignee = t.assignee == actor
            if rule == ASSIGNEE and not is_assignee:
                raise UnauthorizedError(actor + " is not the assignee")
            if rule == NOT_ASSIGNEE and is_assignee:
                raise UnauthorizedError("assignee cannot review own task")
            if t.status is Status.TODO and to is Status.IN_PROGRESS and self._blocked(t):
                raise WorkflowError("blocked by dependencies")
            t.history.append("%s->%s by %s" % (t.status.value, to.value, actor))
            t.status = to

    def status(self, task_id: str) -> Status:
        with self._lock:
            return self._task(task_id).status

    def tasks_for(self, user: str) -> List[str]:
        with self._lock:
            mine = [t for t in self._tasks.values() if t.assignee == user and t.status is not Status.DONE]
            return [t.id for t in sorted(mine, key=lambda t: (-t.priority, t.seq))]

    def blocked_tasks(self) -> List[str]:
        with self._lock:
            return [t.id for t in self._tasks.values() if t.status is not Status.DONE and self._blocked(t)]

    def history(self, task_id: str) -> List[str]:
        with self._lock:
            return list(self._task(task_id).history)

    def _blocked(self, t: _Task) -> bool:
        return any(self._tasks[d].status is not Status.DONE for d in t.depends_on)

    def _task(self, task_id: str) -> _Task:
        t = self._tasks.get(task_id)
        if t is None:
            raise ValueError("unknown task " + task_id)
        return t
