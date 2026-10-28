import java.util.*;

public class Tests {
    static void testNotifiesOnlyOnChangeInOrder() {
        StockTicker t = new StockTicker();
        List<String> log = new ArrayList<>();
        t.subscribe((s, p) -> log.add("a:" + s + p));
        t.subscribe((s, p) -> log.add("b:" + s + p));
        t.update("X", 1);
        t.update("X", 1);
        t.update("Y", 1);
        t.update("X", 2);
        T.eq(List.of("a:X1", "b:X1", "a:Y1", "b:Y1", "a:X2", "b:X2"), log, "same price twice is not a change; order = subscription order");
    }

    static void testUnsubscribeHandle() {
        StockTicker t = new StockTicker();
        List<Long> got = new ArrayList<>();
        Runnable off = t.subscribe((s, p) -> got.add(p));
        T.eq(1, t.listenerCount(), "one listener");
        t.update("X", 1);
        off.run();
        off.run();
        t.update("X", 2);
        T.eq(List.of(1L), got, "no events after unsubscribe");
        T.eq(0, t.listenerCount(), "count after unsubscribe (idempotent)");
    }

    static void testFaultyListenerIsolated() {
        StockTicker t = new StockTicker();
        List<Long> got = new ArrayList<>();
        t.subscribe((s, p) -> { throw new IllegalStateException("boom"); });
        t.subscribe((s, p) -> got.add(p));
        t.update("X", 7);
        T.eq(List.of(7L), got, "listener after the faulty one still notified");
    }

    static void testUnsubscribeDuringNotification() {
        StockTicker t = new StockTicker();
        List<String> log = new ArrayList<>();
        Runnable[] selfOff = new Runnable[1];
        selfOff[0] = t.subscribe((s, p) -> { log.add("once" + p); selfOff[0].run(); });
        t.subscribe((s, p) -> log.add("always" + p));
        t.update("X", 1);
        t.update("X", 2);
        T.eq(List.of("once1", "always1", "always2"), log, "self-unsubscribing listener gets one event, others unaffected");
    }

    static void testAlertEdgeTriggeredAbove() {
        StockTicker t = new StockTicker();
        List<String> sink = new ArrayList<>();
        t.subscribe(new ThresholdAlert("AAPL", 150, true, sink));
        for (long p : new long[] {149, 151, 155, 148, 150}) t.update("AAPL", p);
        T.eq(List.of("AAPL>=150 @151", "AAPL>=150 @150"), sink, "fires on crossing, re-arms after dropping below");
    }

    static void testAlertBelowAndOtherSymbols() {
        StockTicker t = new StockTicker();
        List<String> sink = new ArrayList<>();
        t.subscribe(new ThresholdAlert("TSLA", 100, false, sink));
        t.update("AAPL", 10);
        t.update("TSLA", 99);
        t.update("TSLA", 98);
        T.eq(List.of("TSLA<=100 @99"), sink, "below-threshold alert, other symbols ignored, no repeat while still below");
    }
}
