from lldtest import check, eq, raises
from solution import TextEditor


def test_insert_and_delete():
    e = TextEditor(10)
    e.insert(0, "hello")
    e.insert(5, " world")
    e.insert(0, ">")
    eq(">hello world", e.text(), "inserts")
    e.delete(0, 1)
    e.delete(5, 11)
    eq("hello", e.text(), "deletes")


def test_undo_redo_sequence():
    e = TextEditor(10)
    e.insert(0, "hello")
    e.insert(5, " world")
    e.delete(0, 6)
    eq("world", e.text(), "after edits")
    check(e.undo(), "undo 1")
    eq("hello world", e.text(), "undo delete restores removed text")
    check(e.undo(), "undo 2")
    eq("hello", e.text(), "undo insert")
    check(e.redo(), "redo")
    eq("hello world", e.text(), "redo insert")


def test_new_edit_clears_redo():
    e = TextEditor(10)
    e.insert(0, "ab")
    e.undo()
    e.insert(0, "x")
    check(not e.redo(), "redo must be cleared by a new edit")
    eq("x", e.text(), "text")


def test_empty_stacks():
    e = TextEditor(10)
    check(not e.undo(), "nothing to undo")
    check(not e.redo(), "nothing to redo")
    e.insert(0, "a")
    e.undo()
    check(not e.undo(), "undo past beginning")
    eq("", e.text(), "empty")


def test_invalid_ops_change_nothing():
    e = TextEditor(10)
    e.insert(0, "abc")
    raises(IndexError, lambda: e.insert(4, "x"), "insert past end")
    raises(IndexError, lambda: e.insert(-1, "x"), "negative insert")
    raises(IndexError, lambda: e.delete(2, 1), "start > end")
    raises(IndexError, lambda: e.delete(0, 4), "end past end")
    eq("abc", e.text(), "text unchanged")
    check(e.undo(), "only the valid insert is in history")
    check(not e.undo(), "invalid ops were not recorded")


def test_no_ops_not_recorded():
    e = TextEditor(10)
    e.insert(0, "abc")
    e.insert(1, "")
    e.delete(2, 2)
    e.undo()
    eq("", e.text(), "single undo reverts the only real edit")
    check(not e.undo(), "no-ops not in history")


def test_bounded_history_drops_oldest():
    e = TextEditor(2)
    e.insert(0, "a")
    e.insert(1, "b")
    e.insert(2, "c")
    check(e.undo(), "undo c")
    check(e.undo(), "undo b")
    check(not e.undo(), "a dropped from history")
    eq("a", e.text(), "oldest edit can no longer be undone")
