**Interviewer:** *"Let's build a mini Jira. Tasks move through a workflow, some tasks depend on others, and we need an audit trail."*

## Tasks

`create(title, reporter, priority)` → `"T-1"`, `"T-2"`, … (`priority` 1–5, where 5 is most urgent). New tasks are `TODO`.
`assign(taskId, user)` reassigns (not allowed once `DONE`).

## Workflow

| From → To | Who may do it | Extra rule |
|---|---|---|
| TODO → IN_PROGRESS | the **assignee** | every dependency must be DONE (otherwise *blocked*) |
| IN_PROGRESS → IN_REVIEW | the **assignee** | |
| IN_REVIEW → DONE | anyone **except** the assignee (a reviewer) | |
| IN_REVIEW → IN_PROGRESS | anyone **except** the assignee (changes requested) | |

Any other transition is illegal. A task with no assignee can't start. The wrong actor is *unauthorized*.

## Dependencies

`addDependency(task, dependsOn)`: `task` can't start until `dependsOn` is DONE. Self-dependencies and anything that would create a **cycle** (A→B→C→A) are rejected.

`blockedTasks()` returns the ids of tasks that aren't DONE and have at least one dependency that isn't DONE, in creation order.

## Queries

- `status(task)`
- `tasksFor(user)`: tasks assigned to the user that aren't DONE, sorted by **priority desc**, then creation order
- `history(task)`: audit lines, oldest first, in exactly these formats:
  - `created by <reporter>`
  - `assigned to <user>`
  - `<FROM>-><TO> by <actor>` (for example `TODO->IN_PROGRESS by bob`)

<div class="lang-note" data-lang="java">

Provided: `enum Status { TODO, IN_PROGRESS, IN_REVIEW, DONE }`, `UnauthorizedException`. Unknown ids or an invalid priority throw `IllegalArgumentException`. Illegal transitions, cycles, blocked starts and changes to DONE tasks throw `IllegalStateException`.

</div>

<div class="lang-note" data-lang="go">

Provided: `Status` (`Todo`, `InProgress`, `InReview`, `Done`) with `String()` returning the names above. `NewTaskBoard()`. Errors: `ErrInvalid`, `ErrNotFound`, `ErrIllegalTransition`, `ErrUnauthorized`, `ErrBlocked`, `ErrCycle`.

</div>

<div class="lang-note" data-lang="python">

Provided: the `Status` enum (its `.value` is the name shown above), plus `UnauthorizedError` and `WorkflowError` (for illegal transitions, cycles, blocked starts and changes to DONE tasks). Unknown ids or an invalid priority raise `ValueError`. Methods are snake_case: `add_dependency`, `move(task_id, to, actor)`, `tasks_for`, `blocked_tasks`, `history`.

</div>
