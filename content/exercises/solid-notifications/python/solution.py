from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import FrozenSet, List, Optional


# ---- Provided ----
@dataclass(frozen=True)
class User:
    id: str
    email: Optional[str]
    phone: Optional[str]
    device_token: Optional[str]
    muted: FrozenSet[str] = frozenset()


@dataclass
class Outbox:
    sent: List[str] = field(default_factory=list)

    def record(self, channel: str, to: str, message: str) -> None:
        self.sent.append("%s:%s:%s" % (channel, to, message))


# ---- The code you are replacing (read it, then delete it) ----
class LegacyNotifications:
    # Problems: new channel => edit this method; hard-wired clients; one failure stops the rest.
    def notify(self, u: User, msg: str, outbox: Outbox) -> int:
        sent = 0
        if u.email and "email" not in u.muted:
            outbox.record("email", u.email, msg)
            sent += 1
        if u.phone and "sms" not in u.muted:
            outbox.record("sms", u.phone, msg)  # if this raised, push never runs
            sent += 1
        if u.device_token and "push" not in u.muted:
            outbox.record("push", u.device_token, msg)
            sent += 1
        return sent


# ---- Your design ----
class Notifier(ABC):
    @abstractmethod
    def channel(self) -> str: ...

    @abstractmethod
    def can_reach(self, user: User) -> bool: ...

    @abstractmethod
    def send(self, user: User, message: str) -> None: ...


class EmailNotifier(Notifier):
    def __init__(self, outbox: Outbox) -> None:
        pass  # TODO

    def channel(self) -> str:
        raise NotImplementedError("TODO")

    def can_reach(self, user: User) -> bool:
        raise NotImplementedError("TODO")

    def send(self, user: User, message: str) -> None:
        raise NotImplementedError("TODO")


class SmsNotifier(Notifier):
    def __init__(self, outbox: Outbox) -> None:
        pass  # TODO

    def channel(self) -> str:
        raise NotImplementedError("TODO")

    def can_reach(self, user: User) -> bool:
        raise NotImplementedError("TODO")

    def send(self, user: User, message: str) -> None:
        raise NotImplementedError("TODO")


class PushNotifier(Notifier):
    def __init__(self, outbox: Outbox) -> None:
        pass  # TODO

    def channel(self) -> str:
        raise NotImplementedError("TODO")

    def can_reach(self, user: User) -> bool:
        raise NotImplementedError("TODO")

    def send(self, user: User, message: str) -> None:
        raise NotImplementedError("TODO")


class NotificationService:
    def __init__(self, notifiers: List[Notifier]) -> None:
        pass  # TODO

    def notify(self, user: User, message: str) -> int:
        """Returns the number of successful sends."""
        raise NotImplementedError("TODO")
