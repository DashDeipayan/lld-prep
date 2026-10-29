from lldtest import eq, raises
from solution import Status, TaskBoard, UnauthorizedError, WorkflowError


def test_workflow_and_actors():
    b = TaskBoard()
    t = b.create("Fix login", "pm", 3)
    eq("T-1", t, "id")
    raises(WorkflowError, lambda: b.move(t, Status.IN_PROGRESS, "bob"), "unassigned task can't start")
    b.assign(t, "bob")
    raises(UnauthorizedError, lambda: b.move(t, Status.IN_PROGRESS, "eve"), "only assignee starts")
    b.move(t, Status.IN_PROGRESS, "bob")
    raises(WorkflowError, lambda: b.move(t, Status.DONE, "alice"), "can't skip review")
    b.move(t, Status.IN_REVIEW, "bob")
    raises(UnauthorizedError, lambda: b.move(t, Status.DONE, "bob"), "no self-approval")
    b.move(t, Status.IN_PROGRESS, "alice")
    b.move(t, Status.IN_REVIEW, "bob")
    b.move(t, Status.DONE, "alice")
    eq(Status.DONE, b.status(t), "done")
    raises(WorkflowError, lambda: b.move(t, Status.TODO, "bob"), "done is terminal")
    raises(WorkflowError, lambda: b.assign(t, "eve"), "can't reassign done task")


def test_history():
    b = TaskBoard()
    t = b.create("X", "pm", 1)
    b.assign(t, "bob")
    b.move(t, Status.IN_PROGRESS, "bob")
    b.move(t, Status.IN_REVIEW, "bob")
    eq(["created by pm", "assigned to bob", "TODO->IN_PROGRESS by bob", "IN_PROGRESS->IN_REVIEW by bob"], b.history(t), "audit trail")


def test_dependencies_block_start():
    b = TaskBoard()
    api, ui = b.create("API", "pm", 2), b.create("UI", "pm", 2)
    b.add_dependency(ui, api)
    b.assign(api, "bob")
    b.assign(ui, "carol")
    eq([ui], b.blocked_tasks(), "ui blocked")
    raises(WorkflowError, lambda: b.move(ui, Status.IN_PROGRESS, "carol"), "blocked start")
    b.move(api, Status.IN_PROGRESS, "bob")
    b.move(api, Status.IN_REVIEW, "bob")
    b.move(api, Status.DONE, "alice")
    eq([], b.blocked_tasks(), "unblocked")
    b.move(ui, Status.IN_PROGRESS, "carol")
    eq(Status.IN_PROGRESS, b.status(ui), "started")


def test_cycle_detection():
    b = TaskBoard()
    a, c, d = b.create("A", "pm", 1), b.create("B", "pm", 1), b.create("C", "pm", 1)
    b.add_dependency(a, c)
    b.add_dependency(c, d)
    raises(WorkflowError, lambda: b.add_dependency(d, a), "A->B->C->A cycle")
    raises(WorkflowError, lambda: b.add_dependency(a, a), "self dependency")
    b.add_dependency(a, d)
    raises(ValueError, lambda: b.add_dependency(a, "T-99"), "unknown task")


def test_tasks_for_user_sorted():
    b = TaskBoard()
    low, high, mid1, mid2 = b.create("low", "pm", 1), b.create("high", "pm", 5), b.create("m1", "pm", 3), b.create("m2", "pm", 3)
    for t in (low, high, mid1, mid2):
        b.assign(t, "bob")
    b.assign(mid2, "carol")
    eq([high, mid1, low], b.tasks_for("bob"), "priority desc, then creation order")
    raises(ValueError, lambda: b.create("bad", "pm", 6), "priority range")
