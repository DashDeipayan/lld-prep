from __future__ import annotations

import threading
from dataclasses import dataclass, field
from typing import Dict, List


@dataclass(frozen=True)
class Message:
    offset: int
    payload: str


# ---- Provided errors ----
class TopicExistsError(Exception):
    pass


class UnknownTopicError(Exception):
    pass


class UnknownGroupError(Exception):
    pass


@dataclass
class _Topic:
    """One lock per topic: different topics never contend."""

    log: List[str] = field(default_factory=list)
    offsets: Dict[str, int] = field(default_factory=dict)
    lock: threading.Lock = field(default_factory=threading.Lock)

    def offset(self, group: str) -> int:
        if group not in self.offsets:
            raise UnknownGroupError("group not subscribed: " + group)
        return self.offsets[group]


class Broker:
    def __init__(self) -> None:
        self._topics: Dict[str, _Topic] = {}
        self._lock = threading.Lock()  # guards the topics map only

    def create_topic(self, topic: str) -> None:
        with self._lock:
            if topic in self._topics:
                raise TopicExistsError(topic)
            self._topics[topic] = _Topic()

    def publish(self, topic: str, payload: str) -> int:
        t = self._topic(topic)
        with t.lock:
            t.log.append(payload)
            return len(t.log) - 1

    def subscribe(self, topic: str, group: str, from_beginning: bool) -> None:
        t = self._topic(topic)
        with t.lock:
            t.offsets.setdefault(group, 0 if from_beginning else len(t.log))  # existing group: keep its offset

    def poll(self, topic: str, group: str, max_messages: int) -> List[Message]:
        t = self._topic(topic)
        with t.lock:  # read and advance atomically, or two consumers get the same batch
            start = t.offset(group)
            end = min(len(t.log), start + max(0, max_messages))
            t.offsets[group] = end
            return [Message(i, t.log[i]) for i in range(start, end)]

    def seek(self, topic: str, group: str, offset: int) -> None:
        t = self._topic(topic)
        with t.lock:
            t.offset(group)
            if not 0 <= offset <= len(t.log):
                raise ValueError("offset out of range: %d" % offset)
            t.offsets[group] = offset

    def lag(self, topic: str, group: str) -> int:
        t = self._topic(topic)
        with t.lock:
            return len(t.log) - t.offset(group)

    def _topic(self, name: str) -> _Topic:
        with self._lock:
            t = self._topics.get(name)
        if t is None:
            raise UnknownTopicError(name)
        return t
