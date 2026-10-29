from __future__ import annotations

import dataclasses
import itertools
import string
import threading
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from enum import Enum
from typing import Callable, Dict, List, Optional

# ======================= Provided =======================
Clock = Callable[[], int]  # milliseconds


class PaymentGateway(ABC):
    @abstractmethod
    def charge(self, user: str, amount: int) -> bool: ...


@dataclass(frozen=True)
class Booking:
    id: str
    show_id: str
    user: str
    seats: List[str]  # sorted
    amount: int
    status: str  # "CONFIRMED" or "CANCELLED"


class SeatUnavailableError(Exception):
    pass


class PaymentFailedError(Exception):
    pass


class HoldNotActiveError(Exception):
    pass


class BookingStateError(Exception):
    """Show overlap, show already started, or a double cancel."""


# ======================= Catalogue =======================
class SeatType(Enum):
    REGULAR = "REGULAR"
    PREMIUM = "PREMIUM"


@dataclass(frozen=True)
class Theatre:
    id: str
    city: str


@dataclass(frozen=True)
class Screen:
    id: str
    theatre: Theatre
    layout: Dict[str, SeatType]


@dataclass(frozen=True)
class Movie:
    id: str
    title: str
    duration_minutes: int


# ======================= Inventory =======================
class SeatStatus(Enum):
    AVAILABLE = "AVAILABLE"
    HELD = "HELD"
    BOOKED = "BOOKED"


@dataclass
class _SeatState:
    status: SeatStatus = SeatStatus.AVAILABLE
    hold_id: Optional[str] = None
    expires_at: int = 0


@dataclass
class Show:
    id: str
    movie: Movie
    screen: Screen
    start: int
    prices: Dict[SeatType, int]
    seats: Dict[str, _SeatState] = field(default_factory=dict)
    lock: threading.Lock = field(default_factory=threading.Lock)  # one lock per show

    def __post_init__(self) -> None:
        self.seats = {s: _SeatState() for s in sorted(self.screen.layout)}

    @property
    def end(self) -> int:
        return self.start + self.movie.duration_minutes * 60_000

    def available(self, st: _SeatState, now: int) -> bool:
        return st.status is SeatStatus.AVAILABLE or (st.status is SeatStatus.HELD and st.expires_at <= now)

    def price(self, seat_id: str) -> int:
        return self.prices[self.screen.layout[seat_id]]

    def set_status(self, seat_ids: List[str], status: SeatStatus, hold_id: Optional[str]) -> None:
        for sid in seat_ids:
            st = self.seats[sid]
            st.status, st.hold_id = status, hold_id


@dataclass(frozen=True)
class _Hold:
    id: str
    show: Show
    user: str
    seats: List[str]
    expires_at: int


class BookingService:
    def __init__(self, gateway: PaymentGateway, clock: Clock, hold_ttl_ms: int) -> None:
        self._gateway = gateway
        self._clock = clock
        self._ttl = hold_ttl_ms
        self._theatres: Dict[str, Theatre] = {}
        self._screens: Dict[str, Screen] = {}
        self._movies: Dict[str, Movie] = {}
        self._shows: Dict[str, Show] = {}
        self._holds: Dict[str, _Hold] = {}
        self._bookings: Dict[str, Booking] = {}
        self._ids = {k: itertools.count(1) for k in "SHB"}
        self._lock = threading.Lock()  # catalogue, id counters and the holds/bookings maps

    # ---------- catalogue ----------
    def add_theatre(self, theatre_id: str, city: str) -> None:
        with self._lock:
            if theatre_id in self._theatres:
                raise ValueError("duplicate theatre")
            self._theatres[theatre_id] = Theatre(theatre_id, city)

    def add_screen(self, theatre_id: str, screen_id: str, regular_rows: int, premium_rows: int, seats_per_row: int) -> None:
        with self._lock:
            t = self._get(self._theatres, theatre_id, "theatre")
            rows = regular_rows + premium_rows
            if regular_rows < 0 or premium_rows < 0 or not 0 < rows <= 26 or seats_per_row < 1:
                raise ValueError("invalid layout")
            if screen_id in self._screens:
                raise ValueError("duplicate screen")
            layout = {
                "%s%d" % (string.ascii_uppercase[r], n): SeatType.REGULAR if r < regular_rows else SeatType.PREMIUM
                for r in range(rows) for n in range(1, seats_per_row + 1)
            }
            self._screens[screen_id] = Screen(screen_id, t, layout)

    def add_movie(self, movie_id: str, title: str, duration_minutes: int) -> None:
        if duration_minutes <= 0:
            raise ValueError("duration must be positive")
        with self._lock:
            if movie_id in self._movies:
                raise ValueError("duplicate movie")
            self._movies[movie_id] = Movie(movie_id, title, duration_minutes)

    def add_show(self, movie_id: str, screen_id: str, start_ms: int, regular_price: int, premium_price: int) -> str:
        with self._lock:
            m = self._get(self._movies, movie_id, "movie")
            sc = self._get(self._screens, screen_id, "screen")
            end = start_ms + m.duration_minutes * 60_000
            for s in self._shows.values():
                if s.screen is sc and start_ms < s.end and s.start < end:  # half-open overlap
                    raise BookingStateError("overlaps show " + s.id)
            show_id = "S%d" % next(self._ids["S"])
            self._shows[show_id] = Show(show_id, m, sc, start_ms, {SeatType.REGULAR: regular_price, SeatType.PREMIUM: premium_price})
            return show_id

    def find_shows(self, city: str, movie_id: str) -> List[str]:
        now = self._clock()
        with self._lock:
            hits = [s for s in self._shows.values()
                    if s.movie.id == movie_id and s.screen.theatre.city.lower() == city.lower() and s.start > now]
        return [s.id for s in sorted(hits, key=lambda s: (s.start, s.id))]

    def available_seats(self, show_id: str) -> List[str]:
        s = self._show(show_id)
        with s.lock:
            now = self._clock()
            return [sid for sid, st in s.seats.items() if s.available(st, now)]

    # ---------- booking ----------
    def hold(self, show_id: str, user: str, seat_ids: List[str]) -> str:
        s = self._show(show_id)
        if not seat_ids or len(set(seat_ids)) != len(seat_ids):
            raise ValueError("invalid seat list")
        with s.lock:
            now = self._clock()
            if now >= s.start:
                raise BookingStateError("show already started")
            for sid in seat_ids:  # validate all, then mutate all
                st = s.seats.get(sid)
                if st is None:
                    raise ValueError("unknown seat " + sid)
                if not s.available(st, now):
                    raise SeatUnavailableError(sid)
            with self._lock:
                hold_id = "H%d" % next(self._ids["H"])
            for sid in seat_ids:
                st = s.seats[sid]
                st.status, st.hold_id, st.expires_at = SeatStatus.HELD, hold_id, now + self._ttl
            h = _Hold(hold_id, s, user, sorted(seat_ids), now + self._ttl)
            with self._lock:
                self._holds[hold_id] = h
            return hold_id

    def confirm(self, hold_id: str) -> Booking:
        with self._lock:
            h = self._holds.pop(hold_id, None)  # a hold can be used once
        if h is None:
            raise HoldNotActiveError(hold_id)
        s = h.show
        with s.lock:
            now = self._clock()
            active = h.expires_at > now and all(
                s.seats[sid].status is SeatStatus.HELD and s.seats[sid].hold_id == hold_id for sid in h.seats)
            if not active:
                raise HoldNotActiveError(hold_id)
            amount = sum(s.price(sid) for sid in h.seats)
            # NOTE: in production, call the gateway outside the show lock (see lesson).
            if not self._gateway.charge(h.user, amount):
                s.set_status(h.seats, SeatStatus.AVAILABLE, None)
                raise PaymentFailedError("payment declined for " + h.user)
            s.set_status(h.seats, SeatStatus.BOOKED, hold_id)
            with self._lock:
                b = Booking("B%d" % next(self._ids["B"]), s.id, h.user, h.seats, amount, "CONFIRMED")
                self._bookings[b.id] = b
            return b

    def cancel(self, booking_id: str) -> Booking:
        with self._lock:
            b = self._get(self._bookings, booking_id, "booking")
        s = self._show(b.show_id)
        with s.lock:
            with self._lock:
                cur = self._bookings[booking_id]
            if cur.status != "CONFIRMED":
                raise BookingStateError("already cancelled")
            if self._clock() >= s.start:
                raise BookingStateError("show already started")
            s.set_status(cur.seats, SeatStatus.AVAILABLE, None)
            cancelled = dataclasses.replace(cur, status="CANCELLED")
            with self._lock:
                self._bookings[booking_id] = cancelled
            return cancelled

    # ---------- helpers ----------
    def _show(self, show_id: str) -> Show:
        with self._lock:
            return self._get(self._shows, show_id, "show")

    @staticmethod
    def _get(m: Dict[str, object], key: str, what: str):
        v = m.get(key)
        if v is None:
            raise ValueError("unknown %s %s" % (what, key))
        return v
