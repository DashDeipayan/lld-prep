from __future__ import annotations

from typing import Mapping, Optional


class HttpRequest:
    def __init__(self, builder: "Builder") -> None:
        pass  # TODO: copy from the builder (defensively!)

    @staticmethod
    def builder(url: Optional[str]) -> "Builder":
        return Builder(url)

    @property
    def method(self) -> str:
        raise NotImplementedError("TODO")

    @property
    def url(self) -> str:
        raise NotImplementedError("TODO")

    @property
    def headers(self) -> Mapping[str, str]:
        raise NotImplementedError("TODO")

    @property
    def timeout_ms(self) -> int:
        raise NotImplementedError("TODO")

    @property
    def body(self) -> Optional[str]:
        raise NotImplementedError("TODO")


class Builder:
    def __init__(self, url: Optional[str]) -> None:
        pass  # TODO

    def method(self, method: str) -> "Builder":
        raise NotImplementedError("TODO")

    def header(self, name: str, value: str) -> "Builder":
        raise NotImplementedError("TODO")

    def timeout_ms(self, ms: int) -> "Builder":
        raise NotImplementedError("TODO")

    def body(self, body: Optional[str]) -> "Builder":
        raise NotImplementedError("TODO")

    def build(self) -> HttpRequest:
        raise NotImplementedError("TODO")
