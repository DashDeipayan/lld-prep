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


class Notifier(ABC):
    @abstractmethod
    def channel(self) -> str: ...

    @abstractmethod
    def can_reach(self, user: User) -> bool: ...

    @abstractmethod
    def send(self, user: User, message: str) -> None: ...


class OutboxNotifier(Notifier):
    """Shared plumbing: each concrete notifier only says which address it uses."""

    def __init__(self, outbox: Outbox) -> None:
        if outbox is None:
            raise ValueError("outbox is required")
        self._outbox = outbox

    @abstractmethod
    def address(self, user: User) -> Optional[str]: ...

    def can_reach(self, user: User) -> bool:
        a = self.address(user)
        return bool(a and a.strip())

    def send(self, user: User, message: str) -> None:
        self._outbox.record(self.channel(), self.address(user), message)


class EmailNotifier(OutboxNotifier):
    def channel(self) -> str:
        return "email"

    def address(self, user: User) -> Optional[str]:
        return user.email


class SmsNotifier(OutboxNotifier):
    def channel(self) -> str:
        return "sms"

    def address(self, user: User) -> Optional[str]:
        return user.phone


class PushNotifier(OutboxNotifier):
    def channel(self) -> str:
        return "push"

    def address(self, user: User) -> Optional[str]:
        return user.device_token


class NotificationService:
    def __init__(self, notifiers: List[Notifier]) -> None:
        self._notifiers = tuple(notifiers)

    def notify(self, user: User, message: str) -> int:
        if not message or not message.strip():
            raise ValueError("message must not be blank")
        sent = 0
        for n in self._notifiers:
            if n.channel() in user.muted or not n.can_reach(user):
                continue
            try:
                n.send(user, message)
                sent += 1
            except Exception:
                # isolate the failure; a real system would log + retry/dead-letter here
                pass
        return sent
