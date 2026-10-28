import java.util.*;

interface Node {
    String name();
    long size();
}

record FileNode(String name, long size) implements Node {}

final class DirNode implements Node {
    private final String name;
    final SortedMap<String, Node> children = new TreeMap<>();

    DirNode(String name) { this.name = name; }
    public String name() { return name; }
    public long size() {
        long total = 0;
        for (Node c : children.values()) total += c.size();   // the Composite recursion
        return total;
    }
}

final class FileSystem {
    private final DirNode root = new DirNode("");

    void mkdir(String path) {
        DirNode cur = root;
        for (String part : split(path)) {
            Node next = cur.children.get(part);
            if (next == null) {
                DirNode d = new DirNode(part);
                cur.children.put(part, d);
                cur = d;
            } else if (next instanceof DirNode d) {
                cur = d;
            } else {
                throw new IllegalStateException("file exists at " + part);
            }
        }
    }

    void createFile(String path, long size) {
        List<String> parts = split(path);
        if (parts.isEmpty()) throw new IllegalArgumentException("invalid path");
        DirNode parent = dir(parts.subList(0, parts.size() - 1));
        String name = parts.get(parts.size() - 1);
        if (parent.children.containsKey(name)) throw new IllegalStateException("exists: " + path);
        parent.children.put(name, new FileNode(name, size));
    }

    long size(String path) { return resolve(split(path)).size(); }

    List<String> ls(String path) {
        Node n = resolve(split(path));
        if (n instanceof FileNode f) return List.of(f.name());
        List<String> out = new ArrayList<>();
        for (Node c : ((DirNode) n).children.values()) out.add(c instanceof DirNode ? c.name() + "/" : c.name());
        return out;
    }

    List<String> find(String path, String ext) {
        List<String> parts = split(path);
        Node start = resolve(parts);
        List<String> out = new ArrayList<>();
        walk(start, parts.isEmpty() ? "" : "/" + String.join("/", parts), "." + ext, out);
        Collections.sort(out);
        return out;
    }

    void delete(String path) {
        List<String> parts = split(path);
        if (parts.isEmpty()) throw new IllegalArgumentException("cannot delete root");
        DirNode parent = dir(parts.subList(0, parts.size() - 1));
        if (parent.children.remove(parts.get(parts.size() - 1)) == null) throw new NoSuchElementException(path);
    }

    // ---- helpers ----
    private void walk(Node n, String fullPath, String suffix, List<String> out) {
        if (n instanceof FileNode) {
            if (fullPath.endsWith(suffix)) out.add(fullPath);
            return;
        }
        for (Node c : ((DirNode) n).children.values()) walk(c, fullPath + "/" + c.name(), suffix, out);
    }

    private Node resolve(List<String> parts) {
        Node cur = root;
        for (String p : parts) {
            if (!(cur instanceof DirNode d)) throw new NoSuchElementException(p);
            cur = d.children.get(p);
            if (cur == null) throw new NoSuchElementException(p);
        }
        return cur;
    }

    private DirNode dir(List<String> parts) {
        Node n = resolve(parts);
        if (!(n instanceof DirNode d)) throw new IllegalStateException("not a directory");
        return d;
    }

    private static List<String> split(String path) {
        if (path == null || !path.startsWith("/")) throw new IllegalArgumentException("absolute path required: " + path);
        List<String> parts = new ArrayList<>();
        for (String s : path.split("/")) if (!s.isEmpty()) parts.add(s);
        return parts;
    }
}
