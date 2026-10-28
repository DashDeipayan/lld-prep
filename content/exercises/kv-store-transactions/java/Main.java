public class Main {
    public static void main(String[] args) {
        KVStore db = new KVStore();
        db.set("a", "10");
        System.out.println("count 10 = " + db.count("10"));
        db.begin();
        db.set("a", "20");
        db.begin();
        db.delete("a");
        System.out.println("inner: a = " + db.get("a"));
        db.rollback();
        System.out.println("after inner rollback: a = " + db.get("a"));
        db.rollback();
        System.out.println("after outer rollback: a = " + db.get("a"));
        db.begin();
        db.set("b", "10");
        db.commit();
        System.out.println("count 10 = " + db.count("10"));
        try { db.rollback(); } catch (IllegalStateException e) { System.out.println("rollback: " + e.getMessage()); }
    }
}
