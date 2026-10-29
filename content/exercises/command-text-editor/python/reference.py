from __future__ import annotations

from abc import ABC, abstractmethod
from collections import deque
from typing import Deque, List


class Command(ABC):
    @abstractmethod
    def execute(self) -> None: ...

    @abstractmethod
    def undo(self) -> None: ...


class Document:
    """The receiver: knows how to edit text, nothing about history."""

    def __init__(self) -> None:
        self.text = ""

    def insert(self, pos: int, s: str) -> None:
        self.text = self.text[:pos] + s + self.text[pos:]

    def delete(self, start: int, end: int) -> str:
        removed = self.text[start:end]
        self.text = self.text[:start] + self.text[end:]
        return removed


class InsertCommand(Command):
    def __init__(self, doc: Document, pos: int, s: str) -> None:
        self._doc, self._pos, self._s = doc, pos, s

    def execute(self) -> None:
        self._doc.insert(self._pos, self._s)

    def undo(self) -> None:
        self._doc.delete(self._pos, self._pos + len(self._s))


class DeleteCommand(Command):
    def __init__(self, doc: Document, start: int, end: int) -> None:
        self._doc, self._start, self._end = doc, start, end
        self._removed = ""

    def execute(self) -> None:
        self._removed = self._doc.delete(self._start, self._end)

    def undo(self) -> None:
        self._doc.insert(self._start, self._removed)


class TextEditor:
    def __init__(self, max_history: int) -> None:
        if max_history < 0:
            raise ValueError("max_history must be >= 0")
        self._doc = Document()
        # deque(maxlen=n) silently drops the oldest entry when a new one is appended.
        self._undo: Deque[Command] = deque(maxlen=max_history)
        self._redo: List[Command] = []

    def text(self) -> str:
        return self._doc.text

    def insert(self, pos: int, s: str) -> None:
        if not 0 <= pos <= len(self._doc.text):
            raise IndexError("pos %d" % pos)
        if not s:
            return  # no-op: accepted, not recorded
        self._run(InsertCommand(self._doc, pos, s))

    def delete(self, start: int, end: int) -> None:
        if not 0 <= start <= end <= len(self._doc.text):
            raise IndexError("[%d, %d)" % (start, end))
        if start == end:
            return
        self._run(DeleteCommand(self._doc, start, end))

    def undo(self) -> bool:
        if not self._undo:
            return False
        c = self._undo.pop()
        c.undo()
        self._redo.append(c)
        return True

    def redo(self) -> bool:
        if not self._redo:
            return False
        c = self._redo.pop()
        c.execute()
        self._undo.append(c)
        return True

    def _run(self, c: Command) -> None:
        c.execute()
        self._undo.append(c)
        self._redo.clear()  # a new edit invalidates the redo history
