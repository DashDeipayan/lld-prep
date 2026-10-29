from __future__ import annotations

from abc import ABC, abstractmethod


# Suggested shape (you may change internals, not the TextEditor API):
class Command(ABC):
    @abstractmethod
    def execute(self) -> None: ...

    @abstractmethod
    def undo(self) -> None: ...


class TextEditor:
    def __init__(self, max_history: int) -> None:
        pass  # TODO

    def text(self) -> str:
        raise NotImplementedError("TODO")

    def insert(self, pos: int, s: str) -> None:
        raise NotImplementedError("TODO")

    def delete(self, start: int, end: int) -> None:
        raise NotImplementedError("TODO")

    def undo(self) -> bool:
        raise NotImplementedError("TODO")

    def redo(self) -> bool:
        raise NotImplementedError("TODO")
