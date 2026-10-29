from solution import JsonFormatter, Level, Logger, MemoryHandler, SimpleFormatter

console, alerts = [], []
log = (Logger("orders", Level.INFO)
       .add_handler(MemoryHandler(Level.DEBUG, console, SimpleFormatter()))
       .add_handler(MemoryHandler(Level.ERROR, alerts, JsonFormatter())))
log.info("created #1")
log.debug("cache miss")
log.error('charge failed for "card"')
print("console:", console)
print("alerts: ", alerts)
