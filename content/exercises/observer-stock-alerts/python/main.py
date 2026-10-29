from solution import StockTicker, ThresholdAlert

ticker = StockTicker()
alerts = []
ticker.subscribe(ThresholdAlert("AAPL", 150, True, alerts))
stop_log = ticker.subscribe(lambda s, p: print("tick", s, p))
for p in [149, 151, 151, 155, 148, 150]:
    ticker.update("AAPL", p)
stop_log()
ticker.update("AAPL", 140)
print("alerts =", alerts, ", listeners =", ticker.listener_count())
