from __future__ import annotations

import json
import threading
from abc import ABC, abstractmethod
from enum import IntEnum
from typing import List, Tuple


class Level(IntEnum):  # IntEnum, so levels compare: Level.DEBUG < Level.INFO
    DEBUG = 10
    INFO = 20
    WARN = 30
    ERROR = 40


class Formatter(ABC):
    @abstractmethod
    def format(self, logger: str, level: Level, msg: str) -> str: ...


class Handler(ABC):
    @abstractmethod
    def handle(self, logger: str, level: Level, msg: str) -> None: ...


class SimpleFormatter(Formatter):
    def format(self, logger: str, level: Level, msg: str) -> str:
        return "%s [%s] %s" % (level.name, logger, msg)


class JsonFormatter(Formatter):
    def format(self, logger: str, level: Level, msg: str) -> str:
        # json.dumps does the escaping; compact separators keep it on one line.
        return json.dumps({"level": level.name, "logger": logger, "msg": msg}, separators=(",", ":"), ensure_ascii=False)


class MemoryHandler(Handler):
    def __init__(self, min_level: Level, sink: List[str], formatter: Formatter) -> None:
        self._min = min_level
        self._sink = sink
        self._formatter = formatter
        self._lock = threading.Lock()

    def handle(self, logger: str, level: Level, msg: str) -> None:
        if level < self._min:
            return
        line = self._formatter.format(logger, level, msg)
        with self._lock:
            self._sink.append(line)


class Logger:
    def __init__(self, name: str, min_level: Level) -> None:
        self._name = name
        self._min = min_level
        self._handlers: Tuple[Handler, ...] = ()
        self._lock = threading.Lock()

    def add_handler(self, handler: Handler) -> "Logger":
        if handler is None:
            raise ValueError("handler is required")
        with self._lock:
            # Copy-on-write: log() reads the tuple without locking.
            self._handlers = self._handlers + (handler,)
        return self

    def set_level(self, level: Level) -> None:
        self._min = level  # a single assignment is atomic, so readers see old or new

    def debug(self, msg: str) -> None:
        self._log(Level.DEBUG, msg)

    def info(self, msg: str) -> None:
        self._log(Level.INFO, msg)

    def warn(self, msg: str) -> None:
        self._log(Level.WARN, msg)

    def error(self, msg: str) -> None:
        self._log(Level.ERROR, msg)

    def _log(self, level: Level, msg: str) -> None:
        if level < self._min:  # cheap check first
            return
        for h in self._handlers:  # every handler gets a look
            try:
                h.handle(self._name, level, msg)
            except Exception:
                pass  # a logger must never take the application down
