from __future__ import annotations

import threading
from dataclasses import dataclass, field
from typing import Callable, Dict, List, Optional

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


# ---- Model ----
SNIPE_WINDOW_MS = 60_000


@dataclass
class _Auction:
    id: str
    seller: str
    item: str
    starting_price: int
    min_increment: int
    end: int
    leader: Optional[str] = None
    highest: int = 0
    result: Optional[Result] = None
    listeners: List[AuctionListener] = field(default_factory=list)
    lock: threading.Lock = field(default_factory=threading.Lock)  # one lock per auction

    def emit(self, event: str) -> None:
        for listener in list(self.listeners):
            listener(self.id, event)


class AuctionHouse:
    def __init__(self, clock: Clock) -> None:
        self._clock = clock
        self._auctions: Dict[str, _Auction] = {}
        self._next = 1
        self._lock = threading.Lock()

    def create(self, seller: str, item: str, starting_price: int, min_increment: int, end_ms: int) -> str:
        if end_ms <= self._clock() or starting_price <= 0 or min_increment <= 0:
            raise ValueError("invalid auction")
        with self._lock:
            a = _Auction("A%d" % self._next, seller, item, starting_price, min_increment, end_ms)
            self._next += 1
            self._auctions[a.id] = a
            return a.id

    def subscribe(self, auction_id: str, listener: AuctionListener) -> None:
        a = self._auction(auction_id)
        with a.lock:
            a.listeners.append(listener)

    def place_bid(self, auction_id: str, bidder: str, amount: int) -> None:
        a = self._auction(auction_id)
        with a.lock:
            now = self._clock()
            if a.result is not None or now >= a.end:  # rules checked in the specified order
                raise BidRejectedError("CLOSED")
            if bidder == a.seller:
                raise ValueError("seller cannot bid")
            if bidder == a.leader:
                raise BidRejectedError("LEADING")
            minimum = a.starting_price if a.leader is None else a.highest + a.min_increment
            if amount < minimum:
                raise BidRejectedError("TOO_LOW")
            previous = a.leader
            a.leader, a.highest = bidder, amount
            if a.end - now < SNIPE_WINDOW_MS:  # anti-sniping
                a.end = now + SNIPE_WINDOW_MS
            # Emitting under the auction's lock keeps every listener's view in bid order.
            a.emit("BID:%s:%d" % (bidder, amount))
            if previous is not None:
                a.emit("OUTBID:" + previous)

    def leader(self, auction_id: str) -> Optional[str]:
        a = self._auction(auction_id)
        with a.lock:
            return a.leader

    def highest(self, auction_id: str) -> int:
        a = self._auction(auction_id)
        with a.lock:
            return a.highest

    def end_time(self, auction_id: str) -> int:
        a = self._auction(auction_id)
        with a.lock:
            return a.end

    def close(self, auction_id: str) -> Result:
        a = self._auction(auction_id)
        with a.lock:
            if a.result is not None:
                return a.result  # idempotent: same result, no new event
            if self._clock() < a.end:
                raise AuctionNotEndedError("auction still running")
            a.result = Result(a.id, a.leader, 0 if a.leader is None else a.highest)
            a.emit("CLOSED:UNSOLD" if a.leader is None else "CLOSED:%s:%d" % (a.leader, a.highest))
            return a.result

    def _auction(self, auction_id: str) -> _Auction:
        with self._lock:
            a = self._auctions.get(auction_id)
        if a is None:
            raise ValueError("unknown auction " + auction_id)
        return a
