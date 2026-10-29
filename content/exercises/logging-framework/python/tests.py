import threading

from lldtest import eq
from solution import Handler, JsonFormatter, Level, Logger, MemoryHandler, SimpleFormatter


def test_formatters():
    eq("WARN [orders] payment slow", SimpleFormatter().format("orders", Level.WARN, "payment slow"), "simple")
    eq('{"level":"INFO","logger":"app","msg":"say \\"hi\\" \\\\o/"}',
       JsonFormatter().format("app", Level.INFO, 'say "hi" \\o/'), "json with escaping")


def test_logger_and_handler_levels():
    console, alerts = [], []
    log = (Logger("orders", Level.INFO)
           .add_handler(MemoryHandler(Level.DEBUG, console, SimpleFormatter()))
           .add_handler(MemoryHandler(Level.ERROR, alerts, SimpleFormatter())))
    log.debug("d")
    log.info("i")
    log.warn("w")
    log.error("e")
    eq(["INFO [orders] i", "WARN [orders] w", "ERROR [orders] e"], console, "logger drops DEBUG")
    eq(["ERROR [orders] e"], alerts, "handler keeps only ERROR")


def test_set_level_at_runtime():
    sink = []
    log = Logger("x", Level.ERROR).add_handler(MemoryHandler(Level.DEBUG, sink, SimpleFormatter()))
    log.info("before")
    log.set_level(Level.DEBUG)
    log.debug("after")
    eq(["DEBUG [x] after"], sink, "level change applies immediately")


def test_order_and_failure_isolation():
    order = []

    class Record(Handler):
        def __init__(self, name):
            self.name = name

        def handle(self, logger, level, msg):
            order.append(self.name)

    class DiskFull(Handler):
        def handle(self, logger, level, msg):
            raise OSError("disk full")

    log = Logger("x", Level.DEBUG).add_handler(Record("first")).add_handler(DiskFull()).add_handler(Record("third"))
    log.info("m")
    eq(["first", "third"], order, "handlers in order, failure isolated")


def test_concurrent_logging():
    sink = []
    log = Logger("x", Level.DEBUG).add_handler(MemoryHandler(Level.DEBUG, sink, SimpleFormatter()))
    start = threading.Barrier(8)

    def worker(tid):
        start.wait()
        for i in range(500):
            log.info("%d-%d" % (tid, i))

    threads = [threading.Thread(target=worker, args=(t,)) for t in range(8)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(10)
    eq(4000, len(sink), "no lost lines under concurrency")
