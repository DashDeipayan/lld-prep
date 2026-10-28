public class Main {
    public static void main(String[] args) {
        TextEditor ed = new TextEditor(100);
        ed.insert(0, "hello");
        ed.insert(5, " world");
        ed.delete(0, 6);
        System.out.println("'" + ed.text() + "'");
        ed.undo();
        System.out.println("undo  -> '" + ed.text() + "'");
        ed.undo();
        System.out.println("undo  -> '" + ed.text() + "'");
        ed.redo();
        System.out.println("redo  -> '" + ed.text() + "'");
        ed.insert(11, "!");
        System.out.println("edit  -> '" + ed.text() + "', redo possible? " + ed.redo());
    }
}
