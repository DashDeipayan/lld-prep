from __future__ import annotations

from types import MappingProxyType
from typing import Dict, Mapping, Optional

METHODS = frozenset({"GET", "POST", "PUT", "PATCH", "DELETE"})


class HttpRequest:
    __slots__ = ("_method", "_url", "_headers", "_timeout_ms", "_body")

    def __init__(self, builder: "Builder") -> None:
        self._method = builder._method
        self._url = builder._url
        # Copy, then wrap read-only: later builder changes can't leak in, callers can't write.
        self._headers = MappingProxyType(dict(builder._headers))
        self._timeout_ms = builder._timeout_ms
        self._body = builder._body

    @staticmethod
    def builder(url: Optional[str]) -> "Builder":
        return Builder(url)

    @property
    def method(self) -> str:
        return self._method

    @property
    def url(self) -> str:
        return self._url

    @property
    def headers(self) -> Mapping[str, str]:
        return self._headers

    @property
    def timeout_ms(self) -> int:
        return self._timeout_ms

    @property
    def body(self) -> Optional[str]:
        return self._body


class Builder:
    def __init__(self, url: Optional[str]) -> None:
        self._url = url
        self._method = "GET"  # defaults live in exactly one place
        self._headers: Dict[str, str] = {}
        self._timeout_ms = 30_000
        self._body: Optional[str] = None

    def method(self, method: str) -> "Builder":
        m = (method or "").upper()
        if m not in METHODS:
            raise ValueError("unsupported method: %r" % method)
        self._method = m
        return self

    def header(self, name: str, value: str) -> "Builder":
        if not name:
            raise ValueError("header name is required")
        self._headers[name.lower()] = value
        return self

    def timeout_ms(self, ms: int) -> "Builder":
        if not 1 <= ms <= 60_000:
            raise ValueError("timeout must be 1..60000 ms")
        self._timeout_ms = ms
        return self

    def body(self, body: Optional[str]) -> "Builder":
        self._body = body
        return self

    def build(self) -> HttpRequest:
        # Cross-field rules are checked here, once everything is known.
        if not self._url or not self._url.startswith(("http://", "https://")):
            raise ValueError("invalid url: %r" % self._url)
        if self._body and self._method in ("GET", "DELETE"):
            raise ValueError("%s must not have a body" % self._method)
        return HttpRequest(self)
