from lldtest import eq
from solution import StockTicker, ThresholdAlert


def test_notifies_only_on_change_in_order():
    t = StockTicker()
    log = []
    t.subscribe(lambda s, p: log.append("a:%s%d" % (s, p)))
    t.subscribe(lambda s, p: log.append("b:%s%d" % (s, p)))
    t.update("X", 1)
    t.update("X", 1)
    t.update("Y", 1)
    t.update("X", 2)
    eq(["a:X1", "b:X1", "a:Y1", "b:Y1", "a:X2", "b:X2"], log, "same price twice is not a change; order = subscription order")


def test_unsubscribe_handle():
    t = StockTicker()
    got = []
    off = t.subscribe(lambda s, p: got.append(p))
    eq(1, t.listener_count(), "one listener")
    t.update("X", 1)
    off()
    off()
    t.update("X", 2)
    eq([1], got, "no events after unsubscribe")
    eq(0, t.listener_count(), "count after unsubscribe (idempotent)")


def test_faulty_listener_isolated():
    t = StockTicker()
    got = []

    def boom(s, p):
        raise RuntimeError("boom")

    t.subscribe(boom)
    t.subscribe(lambda s, p: got.append(p))
    t.update("X", 7)
    eq([7], got, "listener after the faulty one still notified")


def test_unsubscribe_during_notification():
    t = StockTicker()
    log = []
    handle = {}

    def once(s, p):
        log.append("once%d" % p)
        handle["off"]()

    handle["off"] = t.subscribe(once)
    t.subscribe(lambda s, p: log.append("always%d" % p))
    t.update("X", 1)
    t.update("X", 2)
    eq(["once1", "always1", "always2"], log, "self-unsubscribing listener gets one event, others unaffected")


def test_alert_edge_triggered_above():
    t = StockTicker()
    sink = []
    t.subscribe(ThresholdAlert("AAPL", 150, True, sink))
    for p in [149, 151, 155, 148, 150]:
        t.update("AAPL", p)
    eq(["AAPL>=150 @151", "AAPL>=150 @150"], sink, "fires on crossing, re-arms after dropping below")


def test_alert_below_and_other_symbols():
    t = StockTicker()
    sink = []
    t.subscribe(ThresholdAlert("TSLA", 100, False, sink))
    t.update("AAPL", 10)
    t.update("TSLA", 99)
    t.update("TSLA", 98)
    eq(["TSLA<=100 @99"], sink, "below-threshold alert, other symbols ignored, no repeat while still below")
