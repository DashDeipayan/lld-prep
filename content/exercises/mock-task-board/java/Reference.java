import java.util.*;

enum Status { TODO, IN_PROGRESS, IN_REVIEW, DONE }

final class UnauthorizedException extends RuntimeException {
    UnauthorizedException(String m) { super(m); }
}

final class TaskBoard {
    private enum Who { ASSIGNEE, NOT_ASSIGNEE }

    private record Transition(Status from, Status to) {}

    private static final Map<Transition, Who> RULES = Map.of(
            new Transition(Status.TODO, Status.IN_PROGRESS), Who.ASSIGNEE,
            new Transition(Status.IN_PROGRESS, Status.IN_REVIEW), Who.ASSIGNEE,
            new Transition(Status.IN_REVIEW, Status.DONE), Who.NOT_ASSIGNEE,
            new Transition(Status.IN_REVIEW, Status.IN_PROGRESS), Who.NOT_ASSIGNEE);

    private static final class Task {
        final String id, title;
        final int priority, seq;
        String assignee;
        Status status = Status.TODO;
        final Set<String> dependsOn = new LinkedHashSet<>();
        final List<String> history = new ArrayList<>();
        Task(String id, String title, int priority, int seq) { this.id = id; this.title = title; this.priority = priority; this.seq = seq; }
    }

    private final Map<String, Task> tasks = new LinkedHashMap<>();
    private int next = 1;

    synchronized String create(String title, String reporter, int priority) {
        if (priority < 1 || priority > 5) throw new IllegalArgumentException("priority must be 1..5");
        Task t = new Task("T-" + next, title, priority, next++);
        t.history.add("created by " + reporter);
        tasks.put(t.id, t);
        return t.id;
    }

    synchronized void assign(String taskId, String user) {
        Task t = task(taskId);
        if (t.status == Status.DONE) throw new IllegalStateException("task is done");
        t.assignee = user;
        t.history.add("assigned to " + user);
    }

    synchronized void addDependency(String taskId, String dependsOn) {
        Task t = task(taskId);
        task(dependsOn);
        if (taskId.equals(dependsOn) || reachable(dependsOn, taskId)) throw new IllegalStateException("dependency would create a cycle");
        t.dependsOn.add(dependsOn);
    }

    /** Is `target` reachable from `from` following depends-on edges? (iterative DFS) */
    private boolean reachable(String from, String target) {
        Deque<String> stack = new ArrayDeque<>(List.of(from));
        Set<String> seen = new HashSet<>();
        while (!stack.isEmpty()) {
            String cur = stack.pop();
            if (cur.equals(target)) return true;
            if (seen.add(cur)) stack.addAll(tasks.get(cur).dependsOn);
        }
        return false;
    }

    synchronized void move(String taskId, Status to, String actor) {
        Task t = task(taskId);
        Who rule = RULES.get(new Transition(t.status, to));
        if (rule == null) throw new IllegalStateException(t.status + " -> " + to + " is not allowed");
        if (t.assignee == null) throw new IllegalStateException("task is unassigned");
        boolean isAssignee = t.assignee.equals(actor);
        if (rule == Who.ASSIGNEE && !isAssignee) throw new UnauthorizedException(actor + " is not the assignee");
        if (rule == Who.NOT_ASSIGNEE && isAssignee) throw new UnauthorizedException("assignee cannot review own task");
        if (to == Status.IN_PROGRESS && t.status == Status.TODO && isBlocked(t)) throw new IllegalStateException("blocked by dependencies");
        t.history.add(t.status + "->" + to + " by " + actor);
        t.status = to;
    }

    synchronized Status status(String taskId) { return task(taskId).status; }

    synchronized List<String> tasksFor(String user) {
        return tasks.values().stream()
                .filter(t -> user.equals(t.assignee) && t.status != Status.DONE)
                .sorted(Comparator.comparingInt((Task t) -> -t.priority).thenComparingInt(t -> t.seq))
                .map(t -> t.id).toList();
    }

    synchronized List<String> blockedTasks() {
        return tasks.values().stream().filter(t -> t.status != Status.DONE && isBlocked(t)).map(t -> t.id).toList();
    }

    synchronized List<String> history(String taskId) { return List.copyOf(task(taskId).history); }

    private boolean isBlocked(Task t) {
        for (String d : t.dependsOn) if (tasks.get(d).status != Status.DONE) return true;
        return false;
    }

    private Task task(String id) {
        Task t = tasks.get(id);
        if (t == null) throw new IllegalArgumentException("unknown task " + id);
        return t;
    }
}
