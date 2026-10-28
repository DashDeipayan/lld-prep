import java.util.*;

public class Main {
    public static void main(String[] args) {
        StockTicker ticker = new StockTicker();
        List<String> alerts = new ArrayList<>();
        ticker.subscribe(new ThresholdAlert("AAPL", 150, true, alerts));
        Runnable stopLog = ticker.subscribe((s, p) -> System.out.println("tick " + s + " " + p));
        for (long p : new long[] {149, 151, 151, 155, 148, 150}) ticker.update("AAPL", p);
        stopLog.run();
        ticker.update("AAPL", 140);
        System.out.println("alerts = " + alerts + ", listeners = " + ticker.listenerCount());
    }
}
