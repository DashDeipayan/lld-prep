import threading

from lldtest import check, eq, raises
from solution import AuctionHouse, AuctionNotEndedError, BidRejectedError, Result


class Env:
    def __init__(self):
        self.now = 0
        self.events = []
        self.house = AuctionHouse(lambda: self.now)


def reason(fn):
    try:
        fn()
        return "accepted"
    except BidRejectedError as e:
        return e.reason


def test_bid_rules():
    h = Env().house
    a = h.create("sam", "lamp", 100, 10, 1_000_000)
    eq("TOO_LOW", reason(lambda: h.place_bid(a, "asha", 99)), "below starting price")
    h.place_bid(a, "asha", 100)
    eq("LEADING", reason(lambda: h.place_bid(a, "asha", 200)), "leader can't outbid self")
    eq("TOO_LOW", reason(lambda: h.place_bid(a, "ravi", 109)), "needs highest + increment")
    h.place_bid(a, "ravi", 110)
    eq("ravi", h.leader(a), "leader")
    eq(110, h.highest(a), "highest")
    raises(ValueError, lambda: h.place_bid(a, "sam", 500), "seller can't bid")
    raises(ValueError, lambda: h.create("sam", "x", 100, 10, 0), "end must be in the future")


def test_outbid_notifications_in_order():
    env = Env()
    h = env.house
    a = h.create("sam", "lamp", 100, 10, 1_000_000)
    h.subscribe(a, lambda aid, e: env.events.append("%s %s" % (aid, e)))
    h.place_bid(a, "asha", 100)
    h.place_bid(a, "ravi", 150)
    h.place_bid(a, "asha", 160)
    eq(["A1 BID:asha:100", "A1 BID:ravi:150", "A1 OUTBID:asha", "A1 BID:asha:160", "A1 OUTBID:ravi"], env.events, "events")


def test_anti_sniping_extension():
    env = Env()
    h = env.house
    a = h.create("sam", "lamp", 100, 10, 600_000)
    env.now = 500_000
    h.place_bid(a, "asha", 100)
    eq(600_000, h.end_time(a), "100s left: no extension")
    env.now = 570_000
    h.place_bid(a, "ravi", 110)
    eq(630_000, h.end_time(a), "30s left: extended to now + 60s")
    env.now = 629_999
    h.place_bid(a, "asha", 120)
    eq(689_999, h.end_time(a), "extends again")
    env.now = 689_999
    eq("CLOSED", reason(lambda: h.place_bid(a, "ravi", 500)), "closed at end time")


def test_close_is_idempotent():
    env = Env()
    h = env.house
    a = h.create("sam", "lamp", 100, 10, 1_000_000)
    b = h.create("sam", "rug", 100, 10, 1_000_000)
    h.subscribe(a, lambda aid, e: env.events.append(e))
    h.subscribe(b, lambda aid, e: env.events.append(e))
    h.place_bid(a, "asha", 100)
    raises(AuctionNotEndedError, lambda: h.close(a), "can't close early")
    env.now = 1_000_000
    r = h.close(a)
    eq(Result("A1", "asha", 100), r, "winner")
    check(r is h.close(a), "second close returns the same result")
    eq(Result("A2", None, 0), h.close(b), "unsold")
    eq(["BID:asha:100", "CLOSED:asha:100", "CLOSED:UNSOLD"], env.events, "close notifies once")
    eq("CLOSED", reason(lambda: h.place_bid(a, "ravi", 999)), "no bids after close")


def test_concurrent_bids_stay_consistent():
    h = Env().house
    a = h.create("sam", "lamp", 100, 10, 10_000_000)
    accepted = []
    h.subscribe(a, lambda aid, e: accepted.append(int(e.rsplit(":", 1)[1])) if e.startswith("BID:") else None)
    start = threading.Barrier(40)

    def bid(n):
        start.wait()
        reason(lambda: h.place_bid(a, "u%d" % n, 100 + 10 * n))

    threads = [threading.Thread(target=bid, args=(n,)) for n in range(40)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(10)
    check(all(x < y for x, y in zip(accepted, accepted[1:])), "accepted bids strictly increasing: %s" % accepted)
    eq("u39", h.leader(a), "highest bidder wins")
    eq(490, h.highest(a), "highest amount")
