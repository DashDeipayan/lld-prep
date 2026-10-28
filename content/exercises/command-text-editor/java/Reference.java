import java.util.*;

interface Command {
    void execute();
    void undo();
}

/** The receiver: knows how to edit text, nothing about history. */
final class Document {
    private final StringBuilder sb = new StringBuilder();
    String text() { return sb.toString(); }
    int length() { return sb.length(); }
    void insert(int pos, String s) { sb.insert(pos, s); }
    String delete(int from, int to) {
        String removed = sb.substring(from, to);
        sb.delete(from, to);
        return removed;
    }
}

final class InsertCommand implements Command {
    private final Document doc; private final int pos; private final String text;
    InsertCommand(Document doc, int pos, String text) { this.doc = doc; this.pos = pos; this.text = text; }
    public void execute() { doc.insert(pos, text); }
    public void undo() { doc.delete(pos, pos + text.length()); }
}

final class DeleteCommand implements Command {
    private final Document doc; private final int from, to;
    private String removed;
    DeleteCommand(Document doc, int from, int to) { this.doc = doc; this.from = from; this.to = to; }
    public void execute() { removed = doc.delete(from, to); }
    public void undo() { doc.insert(from, removed); }
}

final class TextEditor {
    private final Document doc = new Document();
    private final Deque<Command> undo = new ArrayDeque<>();
    private final Deque<Command> redo = new ArrayDeque<>();
    private final int maxHistory;

    TextEditor(int maxHistory) {
        if (maxHistory < 0) throw new IllegalArgumentException("maxHistory must be >= 0");
        this.maxHistory = maxHistory;
    }

    String text() { return doc.text(); }

    void insert(int pos, String s) {
        if (pos < 0 || pos > doc.length()) throw new IndexOutOfBoundsException("pos " + pos);
        if (s == null || s.isEmpty()) return;
        run(new InsertCommand(doc, pos, s));
    }

    void delete(int from, int to) {
        if (from < 0 || to > doc.length() || from > to) throw new IndexOutOfBoundsException("[" + from + "," + to + ")");
        if (from == to) return;
        run(new DeleteCommand(doc, from, to));
    }

    boolean undo() {
        if (undo.isEmpty()) return false;
        Command c = undo.pop();
        c.undo();
        redo.push(c);
        return true;
    }

    boolean redo() {
        if (redo.isEmpty()) return false;
        Command c = redo.pop();
        c.execute();
        undo.push(c);
        return true;
    }

    private void run(Command c) {
        c.execute();
        undo.push(c);
        redo.clear();
        while (undo.size() > maxHistory) undo.removeLast(); // drop the oldest
    }
}
