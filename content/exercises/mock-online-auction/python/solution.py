from __future__ import annotations

from dataclasses import dataclass
from typing import Callable, Optional

# ---- Provided ----
Clock = Callable[[], int]  # milliseconds
AuctionListener = Callable[[str, str], None]  # listener(auction_id, event)


@dataclass(frozen=True)
class Result:
    auction_id: str
    winner: Optional[str]  # None when unsold
    price: int


class BidRejectedError(Exception):
    def __init__(self, reason: str) -> None:  # "CLOSED", "LEADING" or "TOO_LOW"
        super().__init__(reason)
        self.reason = reason


class AuctionNotEndedError(Exception):
    """close() before the end time."""


# ---- Your design ----
class AuctionHouse:
    def __init__(self, clock: Clock) -> None:
        pass  # TODO

    def create(self, seller: str, item: str, starting_price: int, min_increment: int, end_ms: int) -> str:
        raise NotImplementedError("TODO")

    def subscribe(self, auction_id: str, listener: AuctionListener) -> None:
        raise NotImplementedError("TODO")

    def place_bid(self, auction_id: str, bidder: str, amount: int) -> None:
        raise NotImplementedError("TODO")

    def leader(self, auction_id: str) -> Optional[str]:
        raise NotImplementedError("TODO")

    def highest(self, auction_id: str) -> int:
        raise NotImplementedError("TODO")

    def end_time(self, auction_id: str) -> int:
        raise NotImplementedError("TODO")

    def close(self, auction_id: str) -> Result:
        raise NotImplementedError("TODO")
