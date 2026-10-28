import java.util.*;

public class Tests {
    static void testBasicOpsAndCount() {
        KVStore s = new KVStore();
        s.set("a", "10");
        s.set("b", "10");
        T.eq(Optional.of("10"), s.get("a"), "get");
        T.eq(2, s.count("10"), "two keys with 10");
        s.set("b", "20");
        T.eq(1, s.count("10"), "overwrite moves the count");
        T.eq(1, s.count("20"), "new value counted");
        s.delete("a");
        s.delete("nope");
        T.eq(Optional.empty(), s.get("a"), "deleted");
        T.eq(0, s.count("10"), "count after delete");
    }

    static void testRollbackRestoresInnermostBlock() {
        KVStore s = new KVStore();
        s.set("a", "10");
        s.begin();
        s.set("a", "20");
        s.set("c", "30");
        T.eq(Optional.of("20"), s.get("a"), "uncommitted value visible");
        s.rollback();
        T.eq(Optional.of("10"), s.get("a"), "restored");
        T.eq(Optional.empty(), s.get("c"), "key created in block is gone");
        T.eq(1, s.count("10"), "counts restored");
        T.eq(0, s.count("30"), "counts restored for removed key");
    }

    static void testNestedBlocks() {
        KVStore s = new KVStore();
        s.set("a", "10");
        s.begin();
        s.set("a", "20");
        s.begin();
        s.delete("a");
        T.eq(Optional.empty(), s.get("a"), "deleted in inner block");
        s.rollback();
        T.eq(Optional.of("20"), s.get("a"), "inner rollback only");
        s.rollback();
        T.eq(Optional.of("10"), s.get("a"), "outer rollback");
    }

    static void testMultipleWritesSameKeyInBlock() {
        KVStore s = new KVStore();
        s.set("x", "1");
        s.begin();
        s.set("x", "2");
        s.set("x", "3");
        s.delete("x");
        s.set("x", "4");
        s.rollback();
        T.eq(Optional.of("1"), s.get("x"), "restored to the value before the block, not an intermediate one");
        T.eq(1, s.count("1"), "count");
        T.eq(0, s.count("4"), "count");
    }

    static void testCommitClosesAllBlocks() {
        KVStore s = new KVStore();
        s.begin();
        s.set("a", "1");
        s.begin();
        s.set("b", "1");
        s.commit();
        T.eq(2, s.count("1"), "committed");
        T.throwsEx(IllegalStateException.class, s::rollback, "nothing open after commit");
        T.throwsEx(IllegalStateException.class, s::commit, "commit with nothing open");
    }

    static void testRollbackCostProportionalToChanges() {
        KVStore s = new KVStore();
        for (int i = 0; i < 200_000; i++) s.set("k" + i, "v");
        long t0 = System.nanoTime();
        for (int i = 0; i < 2_000; i++) { s.begin(); s.set("k1", "w"); s.rollback(); }
        long ms = (System.nanoTime() - t0) / 1_000_000;
        T.check(ms < 1_500, "2000 begin/set/rollback cycles on a 200k-key store took " + ms + "ms; don't copy the whole map per transaction");
        T.eq(200_000, s.count("v"), "all restored");
    }
}
