from __future__ import annotations

from abc import ABC, abstractmethod
from enum import IntEnum
from typing import List


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
        raise NotImplementedError("TODO")


class JsonFormatter(Formatter):
    def format(self, logger: str, level: Level, msg: str) -> str:
        raise NotImplementedError("TODO")


class MemoryHandler(Handler):
    def __init__(self, min_level: Level, sink: List[str], formatter: Formatter) -> None:
        pass  # TODO

    def handle(self, logger: str, level: Level, msg: str) -> None:
        raise NotImplementedError("TODO")


class Logger:
    def __init__(self, name: str, min_level: Level) -> None:
        pass  # TODO

    def add_handler(self, handler: Handler) -> "Logger":
        raise NotImplementedError("TODO")

    def set_level(self, level: Level) -> None:
        raise NotImplementedError("TODO")

    def debug(self, msg: str) -> None:
        raise NotImplementedError("TODO")

    def info(self, msg: str) -> None:
        raise NotImplementedError("TODO")

    def warn(self, msg: str) -> None:
        raise NotImplementedError("TODO")

    def error(self, msg: str) -> None:
        raise NotImplementedError("TODO")
