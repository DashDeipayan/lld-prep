public class Tests {
    static void testInsertAndDelete() {
        TextEditor e = new TextEditor(10);
        e.insert(0, "hello");
        e.insert(5, " world");
        e.insert(0, ">");
        T.eq(">hello world", e.text(), "inserts");
        e.delete(0, 1);
        e.delete(5, 11);
        T.eq("hello", e.text(), "deletes");
    }

    static void testUndoRedoSequence() {
        TextEditor e = new TextEditor(10);
        e.insert(0, "hello");
        e.insert(5, " world");
        e.delete(0, 6);
        T.eq("world", e.text(), "after edits");
        T.check(e.undo(), "undo 1");
        T.eq("hello world", e.text(), "undo delete restores removed text");
        T.check(e.undo(), "undo 2");
        T.eq("hello", e.text(), "undo insert");
        T.check(e.redo(), "redo");
        T.eq("hello world", e.text(), "redo insert");
    }

    static void testNewEditClearsRedo() {
        TextEditor e = new TextEditor(10);
        e.insert(0, "ab");
        e.undo();
        e.insert(0, "x");
        T.check(!e.redo(), "redo must be cleared by a new edit");
        T.eq("x", e.text(), "text");
    }

    static void testEmptyStacks() {
        TextEditor e = new TextEditor(10);
        T.check(!e.undo(), "nothing to undo");
        T.check(!e.redo(), "nothing to redo");
        e.insert(0, "a");
        e.undo();
        T.check(!e.undo(), "undo past beginning");
        T.eq("", e.text(), "empty");
    }

    static void testInvalidOpsChangeNothing() {
        TextEditor e = new TextEditor(10);
        e.insert(0, "abc");
        T.throwsEx(IndexOutOfBoundsException.class, () -> e.insert(4, "x"), "insert past end");
        T.throwsEx(IndexOutOfBoundsException.class, () -> e.insert(-1, "x"), "negative insert");
        T.throwsEx(IndexOutOfBoundsException.class, () -> e.delete(2, 1), "from > to");
        T.throwsEx(IndexOutOfBoundsException.class, () -> e.delete(0, 4), "to past end");
        T.eq("abc", e.text(), "text unchanged");
        T.check(e.undo(), "only the valid insert is in history");
        T.check(!e.undo(), "invalid ops were not recorded");
    }

    static void testNoOpsNotRecorded() {
        TextEditor e = new TextEditor(10);
        e.insert(0, "abc");
        e.insert(1, "");
        e.delete(2, 2);
        e.undo();
        T.eq("", e.text(), "single undo reverts the only real edit");
        T.check(!e.undo(), "no-ops not in history");
    }

    static void testBoundedHistoryDropsOldest() {
        TextEditor e = new TextEditor(2);
        e.insert(0, "a");
        e.insert(1, "b");
        e.insert(2, "c");
        T.check(e.undo(), "undo c");
        T.check(e.undo(), "undo b");
        T.check(!e.undo(), "a dropped from history");
        T.eq("a", e.text(), "oldest edit can no longer be undone");
    }
}
