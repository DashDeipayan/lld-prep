from lldtest import eq, raises
from solution import FileSystem


def sample():
    fs = FileSystem()
    fs.mkdir("/home/alice/docs")
    fs.create_file("/home/alice/docs/cv.pdf", 120)
    fs.create_file("/home/alice/docs/notes.txt", 30)
    fs.mkdir("/home/bob")
    fs.create_file("/home/bob/todo.txt", 5)
    fs.create_file("/readme.txt", 1)
    return fs


def test_mkdir_p_and_ls():
    fs = sample()
    eq(["home/", "readme.txt"], fs.ls("/"), "root listing, dirs suffixed and sorted")
    eq(["alice/", "bob/"], fs.ls("/home"), "nested")
    eq(["cv.pdf"], fs.ls("/home/alice/docs/cv.pdf"), "ls on a file")
    fs.mkdir("/home/alice")
    eq(["docs/"], fs.ls("/home/alice"), "mkdir on existing dir is a no-op")


def test_recursive_size():
    fs = sample()
    eq(156, fs.size("/"), "whole tree")
    eq(155, fs.size("/home"), "subtree")
    eq(30, fs.size("/home/alice/docs/notes.txt"), "file")
    fs.mkdir("/empty")
    eq(0, fs.size("/empty"), "empty dir")


def test_find_by_extension():
    fs = sample()
    eq(["/home/alice/docs/notes.txt", "/home/bob/todo.txt", "/readme.txt"], fs.find("/", "txt"), "all txt, sorted")
    eq(["/home/bob/todo.txt"], fs.find("/home/bob", "txt"), "scoped")
    eq([], fs.find("/home", "png"), "none")


def test_delete():
    fs = sample()
    fs.delete("/home/alice")
    eq(["bob/"], fs.ls("/home"), "subtree removed")
    eq(6, fs.size("/"), "size updated")
    fs.delete("/readme.txt")
    raises(FileNotFoundError, lambda: fs.size("/readme.txt"), "file gone")
    raises(ValueError, lambda: fs.delete("/"), "cannot delete root")
    raises(FileNotFoundError, lambda: fs.delete("/nope"), "delete missing")


def test_errors():
    fs = sample()
    raises(FileNotFoundError, lambda: fs.create_file("/missing/x.txt", 1), "parent missing")
    raises(NotADirectoryError, lambda: fs.create_file("/readme.txt/x", 1), "parent is a file")
    raises(FileExistsError, lambda: fs.create_file("/home/bob/todo.txt", 9), "file exists")
    raises(NotADirectoryError, lambda: fs.mkdir("/readme.txt/sub"), "mkdir through a file")
    raises(FileNotFoundError, lambda: fs.ls("/nope"), "ls missing")
    raises(FileNotFoundError, lambda: fs.find("/nope", "txt"), "find missing")
