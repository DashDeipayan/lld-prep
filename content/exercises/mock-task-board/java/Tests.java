import java.util.*;

public class Tests {
    static void testWorkflowAndActors() {
        TaskBoard b = new TaskBoard();
        String t = b.create("Fix login", "pm", 3);
        T.eq("T-1", t, "id");
        T.throwsEx(IllegalStateException.class, () -> b.move(t, Status.IN_PROGRESS, "bob"), "unassigned task can't start");
        b.assign(t, "bob");
        T.throwsEx(UnauthorizedException.class, () -> b.move(t, Status.IN_PROGRESS, "eve"), "only assignee starts");
        b.move(t, Status.IN_PROGRESS, "bob");
        T.throwsEx(IllegalStateException.class, () -> b.move(t, Status.DONE, "alice"), "can't skip review");
        b.move(t, Status.IN_REVIEW, "bob");
        T.throwsEx(UnauthorizedException.class, () -> b.move(t, Status.DONE, "bob"), "no self-approval");
        b.move(t, Status.IN_PROGRESS, "alice");
        b.move(t, Status.IN_REVIEW, "bob");
        b.move(t, Status.DONE, "alice");
        T.eq(Status.DONE, b.status(t), "done");
        T.throwsEx(IllegalStateException.class, () -> b.move(t, Status.TODO, "bob"), "done is terminal");
        T.throwsEx(IllegalStateException.class, () -> b.assign(t, "eve"), "can't reassign done task");
    }

    static void testHistory() {
        TaskBoard b = new TaskBoard();
        String t = b.create("X", "pm", 1);
        b.assign(t, "bob");
        b.move(t, Status.IN_PROGRESS, "bob");
        b.move(t, Status.IN_REVIEW, "bob");
        T.eq(List.of("created by pm", "assigned to bob", "TODO->IN_PROGRESS by bob", "IN_PROGRESS->IN_REVIEW by bob"), b.history(t), "audit trail");
    }

    static void testDependenciesBlockStart() {
        TaskBoard b = new TaskBoard();
        String api = b.create("API", "pm", 2), ui = b.create("UI", "pm", 2);
        b.addDependency(ui, api);
        b.assign(api, "bob");
        b.assign(ui, "carol");
        T.eq(List.of(ui), b.blockedTasks(), "ui blocked");
        T.throwsEx(IllegalStateException.class, () -> b.move(ui, Status.IN_PROGRESS, "carol"), "blocked start");
        b.move(api, Status.IN_PROGRESS, "bob");
        b.move(api, Status.IN_REVIEW, "bob");
        b.move(api, Status.DONE, "alice");
        T.eq(List.of(), b.blockedTasks(), "unblocked");
        b.move(ui, Status.IN_PROGRESS, "carol");
        T.eq(Status.IN_PROGRESS, b.status(ui), "started");
    }

    static void testCycleDetection() {
        TaskBoard b = new TaskBoard();
        String a = b.create("A", "pm", 1), c = b.create("B", "pm", 1), d = b.create("C", "pm", 1);
        b.addDependency(a, c);
        b.addDependency(c, d);
        T.throwsEx(IllegalStateException.class, () -> b.addDependency(d, a), "A->B->C->A cycle");
        T.throwsEx(IllegalStateException.class, () -> b.addDependency(a, a), "self dependency");
        b.addDependency(a, d);
        T.throwsEx(IllegalArgumentException.class, () -> b.addDependency(a, "T-99"), "unknown task");
    }

    static void testTasksForUserSorted() {
        TaskBoard b = new TaskBoard();
        String low = b.create("low", "pm", 1), high = b.create("high", "pm", 5), mid1 = b.create("m1", "pm", 3), mid2 = b.create("m2", "pm", 3);
        for (String t : List.of(low, high, mid1, mid2)) b.assign(t, "bob");
        b.assign(mid2, "carol");
        T.eq(List.of(high, mid1, low), b.tasksFor("bob"), "priority desc, then creation order");
        T.throwsEx(IllegalArgumentException.class, () -> b.create("bad", "pm", 6), "priority range");
    }
}
