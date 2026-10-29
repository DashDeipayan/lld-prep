from solution import LoggingSender, RateLimitedSender, RetryingSender, Sender

HOUR = 3_600_000


class FlakySms(Sender):
    def __init__(self):
        self.calls = 0

    def send(self, to, msg):
        self.calls += 1
        if self.calls % 3 != 0:
            raise RuntimeError("gateway timeout")
        print("  SMS ->", to + ":", msg)


now = [0]
log = []
sender = LoggingSender(RateLimitedSender(RetryingSender(FlakySms(), 3), lambda: now[0], 2, HOUR), log)
for i in range(1, 4):
    try:
        sender.send("+1555", "promo %d" % i)
    except Exception as e:
        print("  rejected:", e)
now[0] += HOUR
sender.send("+1555", "promo 4")
for line in log:
    print(line)
