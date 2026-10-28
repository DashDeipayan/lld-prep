public class Main {
    public static void main(String[] args) {
        TaskBoard board = new TaskBoard();
        String api = board.create("Design API", "pm", 4);
        String ui = board.create("Build UI", "pm", 3);
        board.addDependency(ui, api);
        board.assign(api, "bob");
        board.assign(ui, "carol");
        System.out.println("blocked: " + board.blockedTasks());
        board.move(api, Status.IN_PROGRESS, "bob");
        board.move(api, Status.IN_REVIEW, "bob");
        board.move(api, Status.DONE, "alice");
        board.move(ui, Status.IN_PROGRESS, "carol");
        System.out.println("blocked: " + board.blockedTasks());
        System.out.println("history " + api + ": " + board.history(api));
    }
}
