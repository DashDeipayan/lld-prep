import java.util.*;

enum Status { TODO, IN_PROGRESS, IN_REVIEW, DONE }

final class UnauthorizedException extends RuntimeException {
    UnauthorizedException(String m) { super(m); }
}

final class TaskBoard {
    String create(String title, String reporter, int priority) { throw new UnsupportedOperationException("TODO"); }
    void assign(String taskId, String user) { throw new UnsupportedOperationException("TODO"); }
    void addDependency(String taskId, String dependsOn) { throw new UnsupportedOperationException("TODO"); }
    void move(String taskId, Status to, String actor) { throw new UnsupportedOperationException("TODO"); }
    Status status(String taskId) { throw new UnsupportedOperationException("TODO"); }
    List<String> tasksFor(String user) { throw new UnsupportedOperationException("TODO"); }
    List<String> blockedTasks() { throw new UnsupportedOperationException("TODO"); }
    List<String> history(String taskId) { throw new UnsupportedOperationException("TODO"); }
}
