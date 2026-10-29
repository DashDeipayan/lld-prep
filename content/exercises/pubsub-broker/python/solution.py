from __future__ import annotations

from dataclasses import dataclass
from typing import List


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


class Broker:
    def create_topic(self, topic: str) -> None:
        raise NotImplementedError("TODO")

    def publish(self, topic: str, payload: str) -> int:
        raise NotImplementedError("TODO")

    def subscribe(self, topic: str, group: str, from_beginning: bool) -> None:
        raise NotImplementedError("TODO")

    def poll(self, topic: str, group: str, max_messages: int) -> List[Message]:
        raise NotImplementedError("TODO")

    def seek(self, topic: str, group: str, offset: int) -> None:
        raise NotImplementedError("TODO")

    def lag(self, topic: str, group: str) -> int:
        raise NotImplementedError("TODO")
