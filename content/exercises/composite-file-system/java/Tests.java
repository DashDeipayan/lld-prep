import java.util.*;

public class Tests {
    static FileSystem sample() {
        FileSystem fs = new FileSystem();
        fs.mkdir("/home/alice/docs");
        fs.createFile("/home/alice/docs/cv.pdf", 120);
        fs.createFile("/home/alice/docs/notes.txt", 30);
        fs.mkdir("/home/bob");
        fs.createFile("/home/bob/todo.txt", 5);
        fs.createFile("/readme.txt", 1);
        return fs;
    }

    static void testMkdirPAndLs() {
        FileSystem fs = sample();
        T.eq(List.of("home/", "readme.txt"), fs.ls("/"), "root listing, dirs suffixed and sorted");
        T.eq(List.of("alice/", "bob/"), fs.ls("/home"), "nested");
        T.eq(List.of("cv.pdf"), fs.ls("/home/alice/docs/cv.pdf"), "ls on a file");
        fs.mkdir("/home/alice");
        T.eq(List.of("docs/"), fs.ls("/home/alice"), "mkdir on existing dir is a no-op");
    }

    static void testRecursiveSize() {
        FileSystem fs = sample();
        T.eq(156L, fs.size("/"), "whole tree");
        T.eq(155L, fs.size("/home"), "subtree");
        T.eq(30L, fs.size("/home/alice/docs/notes.txt"), "file");
        fs.mkdir("/empty");
        T.eq(0L, fs.size("/empty"), "empty dir");
    }

    static void testFindByExtension() {
        FileSystem fs = sample();
        T.eq(List.of("/home/alice/docs/notes.txt", "/home/bob/todo.txt", "/readme.txt"), fs.find("/", "txt"), "all txt, sorted");
        T.eq(List.of("/home/bob/todo.txt"), fs.find("/home/bob", "txt"), "scoped");
        T.eq(List.of(), fs.find("/home", "png"), "none");
    }

    static void testDelete() {
        FileSystem fs = sample();
        fs.delete("/home/alice");
        T.eq(List.of("bob/"), fs.ls("/home"), "subtree removed");
        T.eq(6L, fs.size("/"), "size updated");
        fs.delete("/readme.txt");
        T.throwsEx(NoSuchElementException.class, () -> fs.size("/readme.txt"), "file gone");
        T.throwsEx(IllegalArgumentException.class, () -> fs.delete("/"), "cannot delete root");
        T.throwsEx(NoSuchElementException.class, () -> fs.delete("/nope"), "delete missing");
    }

    static void testErrors() {
        FileSystem fs = sample();
        T.throwsEx(NoSuchElementException.class, () -> fs.createFile("/missing/x.txt", 1), "parent missing");
        T.throwsEx(IllegalStateException.class, () -> fs.createFile("/readme.txt/x", 1), "parent is a file");
        T.throwsEx(IllegalStateException.class, () -> fs.createFile("/home/bob/todo.txt", 9), "file exists");
        T.throwsEx(IllegalStateException.class, () -> fs.mkdir("/readme.txt/sub"), "mkdir through a file");
        T.throwsEx(NoSuchElementException.class, () -> fs.ls("/nope"), "ls missing");
        T.throwsEx(NoSuchElementException.class, () -> fs.find("/nope", "txt"), "find missing");
    }
}
