import random
import threading

from lldtest import check, eq, raises
from solution import InsufficientFundsError, Txn, Wallet


def wallet(*ids):
    w = Wallet()
    for i in ids:
        w.open_account(i)
    return w


def run_threads(targets, timeout):
    threads = [threading.Thread(target=t, daemon=True) for t in targets]
    for t in threads:
        t.start()
    for t in threads:
        t.join(timeout)
    return all(not t.is_alive() for t in threads)


def test_deposit_transfer_and_history():
    w = wallet("a", "b")
    eq("T1", w.deposit("a", 500), "deposit id")
    eq("T2", w.transfer("k1", "a", "b", 200), "transfer id")
    eq(300, w.balance("a"), "a debited")
    eq(200, w.balance("b"), "b credited")
    eq([Txn("T1", "EXTERNAL", "a", 500), Txn("T2", "a", "b", 200)], w.history("a"), "a history")
    eq([Txn("T2", "a", "b", 200)], w.history("b"), "b history")


def test_validation():
    w = wallet("a", "b")
    w.deposit("a", 100)
    raises(ValueError, lambda: w.open_account("a"), "duplicate account")
    raises(ValueError, lambda: w.deposit("a", 0), "zero deposit")
    raises(ValueError, lambda: w.transfer("k1", "a", "zz", 10), "unknown account")
    raises(ValueError, lambda: w.transfer("k2", "a", "a", 10), "self transfer")
    raises(ValueError, lambda: w.transfer("k3", "a", "b", -1), "negative")
    raises(InsufficientFundsError, lambda: w.transfer("k4", "a", "b", 101), "insufficient")
    eq(100, w.balance("a"), "nothing moved")
    eq(1, len(w.history("a")), "failed transfers not recorded")


def test_idempotent_retries():
    w = wallet("a", "b")
    w.deposit("a", 100)
    t = w.transfer("pay-1", "a", "b", 60)
    eq(t, w.transfer("pay-1", "a", "b", 60), "same key -> same txn id")
    eq(40, w.balance("a"), "charged once")
    raises(InsufficientFundsError, lambda: w.transfer("pay-2", "a", "b", 60), "second payment fails")
    w.deposit("a", 1000)
    raises(InsufficientFundsError, lambda: w.transfer("pay-2", "a", "b", 60), "retry of a failed key returns the same failure")
    eq(1040, w.balance("a"), "failed key not re-executed")


def test_concurrent_random_transfers_conserve_money():
    ids = ["acc%d" % i for i in range(10)]
    w = Wallet()
    for i in ids:
        w.open_account(i)
        w.deposit(i, 1_000)

    def worker(seed):
        r = random.Random(seed)
        for i in range(125):
            a = r.randrange(10)
            b = (a + 1 + r.randrange(9)) % 10
            try:
                w.transfer("t%d-%d" % (seed, i), ids[a], ids[b], 1 + r.randrange(300))
            except InsufficientFundsError:
                pass

    check(run_threads([lambda s=s: worker(s) for s in range(8)], 10), "transfers finished")
    total = 0
    for i in ids:
        b = w.balance(i)
        check(b >= 0, "%s went negative: %d" % (i, b))
        total += b
    eq(10_000, total, "money conserved")


def test_opposite_direction_transfers_do_not_deadlock():
    w = wallet("A", "B")
    w.deposit("A", 1_000_000)
    w.deposit("B", 1_000_000)
    start = threading.Barrier(2)

    def ab():
        start.wait()
        for i in range(200):
            w.transfer("ab%d" % i, "A", "B", 1)

    def ba():
        start.wait()
        for i in range(200):
            w.transfer("ba%d" % i, "B", "A", 1)

    check(run_threads([ab, ba], 6), "A->B and B->A at once must not deadlock")
    eq(2_000_000, w.balance("A") + w.balance("B"), "conserved")


def test_concurrent_same_key_executes_once():
    w = wallet("a", "b")
    w.deposit("a", 1_000)
    start = threading.Barrier(50)
    got = []

    def retry():
        start.wait()
        got.append(w.transfer("same", "a", "b", 10))

    check(run_threads([retry] * 50, 5), "all retries returned")
    eq(1, len(set(got)), "every retry sees the same txn id")
    eq(990, w.balance("a"), "money moved exactly once")
