import java.util.*;

// Suggested shape (you may change internals, not the TextEditor API):
interface Command {
    void execute();
    void undo();
}

final class TextEditor {
    TextEditor(int maxHistory) {
        // TODO
    }

    String text() { throw new UnsupportedOperationException("TODO"); }

    void insert(int pos, String s) { throw new UnsupportedOperationException("TODO"); }

    void delete(int from, int to) { throw new UnsupportedOperationException("TODO"); }

    boolean undo() { throw new UnsupportedOperationException("TODO"); }

    boolean redo() { throw new UnsupportedOperationException("TODO"); }
}
