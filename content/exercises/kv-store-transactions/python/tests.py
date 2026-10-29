import time

from lldtest import check, eq, raises
from solution import KVStore, NoTransactionError


def test_basic_ops_and_count():
    s = KVStore()
    s.set("a", "10")
    s.set("b", "10")
    eq("10", s.get("a"), "get")
    eq(2, s.count("10"), "two keys with 10")
    s.set("b", "20")
    eq(1, s.count("10"), "overwrite moves the count")
    eq(1, s.count("20"), "new value counted")
    s.delete("a")
    s.delete("nope")
    eq(None, s.get("a"), "deleted")
    eq(0, s.count("10"), "count after delete")


def test_rollback_restores_innermost_block():
    s = KVStore()
    s.set("a", "10")
    s.begin()
    s.set("a", "20")
    s.set("c", "30")
    eq("20", s.get("a"), "uncommitted value visible")
    s.rollback()
    eq("10", s.get("a"), "restored")
    eq(None, s.get("c"), "key created in block is gone")
    eq(1, s.count("10"), "counts restored")
    eq(0, s.count("30"), "counts restored for removed key")


def test_nested_blocks():
    s = KVStore()
    s.set("a", "10")
    s.begin()
    s.set("a", "20")
    s.begin()
    s.delete("a")
    eq(None, s.get("a"), "deleted in inner block")
    s.rollback()
    eq("20", s.get("a"), "inner rollback only")
    s.rollback()
    eq("10", s.get("a"), "outer rollback")


def test_multiple_writes_same_key_in_block():
    s = KVStore()
    s.set("x", "1")
    s.begin()
    s.set("x", "2")
    s.set("x", "3")
    s.delete("x")
    s.set("x", "4")
    s.rollback()
    eq("1", s.get("x"), "restored to the value before the block, not an intermediate one")
    eq(1, s.count("1"), "count")
    eq(0, s.count("4"), "count")


def test_commit_closes_all_blocks():
    s = KVStore()
    s.begin()
    s.set("a", "1")
    s.begin()
    s.set("b", "1")
    s.commit()
    eq(2, s.count("1"), "committed")
    raises(NoTransactionError, s.rollback, "nothing open after commit")
    raises(NoTransactionError, s.commit, "commit with nothing open")


def test_rollback_cost_proportional_to_changes():
    s = KVStore()
    for i in range(200_000):
        s.set("k%d" % i, "v")
    t0 = time.monotonic()
    for _ in range(2_000):
        s.begin()
        s.set("k1", "w")
        s.rollback()
    ms = (time.monotonic() - t0) * 1000
    check(ms < 1_500, "2000 begin/set/rollback cycles on a 200k-key store took %dms; don't copy the whole dict per transaction" % ms)
    eq(200_000, s.count("v"), "all restored")
