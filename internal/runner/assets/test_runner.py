"""Runs the functions in tests.py whose names match a regex (default ^test).

Each test runs in its own daemon thread with a time limit, and its outcome is
printed as an "@@TEST|name|PASS|message" line that the server parses.
"""
import inspect
import os
import re
import sys
import threading
import traceback

TIMEOUT = 8.0
HERE = os.path.basename(__file__)


def report(name, passed, msg=""):
    safe = (msg or "").replace("\r", "").replace("\n", "\\n")
    sys.stdout.write("\n@@TEST|%s|%s|%s\n" % (name, "PASS" if passed else "FAIL", safe))
    sys.stdout.flush()


def frames(exc):
    """The traceback without this runner's own frames, newest last."""
    return [f for f in traceback.extract_tb(exc.__traceback__) if os.path.basename(f.filename) != HERE]


def describe(exc):
    if isinstance(exc, AssertionError):
        msg = str(exc)
        if msg:
            return msg
        # A bare `assert x == y`: show the line that failed.
        tb = frames(exc)
        if tb:
            f = tb[-1]
            return "assertion failed: %s (%s:%d)" % (f.line, os.path.basename(f.filename), f.lineno)
        return "assertion failed"
    lines = ["Unexpected exception: %s: %s" % (type(exc).__name__, exc)]
    for f in frames(exc)[-6:]:
        lines.append("  %s:%d in %s: %s" % (os.path.basename(f.filename), f.lineno, f.name, f.line or ""))
    return "\n".join(lines)


def main():
    pattern = re.compile(sys.argv[1] if len(sys.argv) > 1 else r"^test")
    try:
        import tests
    except BaseException as e:  # noqa: B902 - e.g. an exception raised at import time in solution.py
        sys.stderr.write("Your code failed while loading:\n")
        for f in frames(e):
            sys.stderr.write('  File "%s", line %d, in %s\n    %s\n' % (os.path.basename(f.filename), f.lineno, f.name, f.line or ""))
        sys.stderr.write("".join(traceback.format_exception_only(type(e), e)))
        sys.stderr.flush()
        os._exit(1)

    selected = [
        (name, fn)
        for name, fn in vars(tests).items()
        if inspect.isfunction(fn) and fn.__module__ == tests.__name__ and pattern.search(name)
    ]
    for name, fn in selected:  # definition order
        box = {}

        def target(fn=fn, box=box):
            try:
                fn()
            except BaseException as e:  # noqa: B902 - report everything, including SystemExit
                box["exc"] = e

        t = threading.Thread(target=target, name=name, daemon=True)
        t.start()
        t.join(TIMEOUT)
        if t.is_alive():
            report(name, False, "Timed out after %ds (infinite loop or deadlock?)" % TIMEOUT)
            continue
        exc = box.get("exc")
        report(name, exc is None, "" if exc is None else describe(exc))
    sys.stdout.flush()
    os._exit(0)  # don't wait for threads stuck in a timed-out test


if __name__ == "__main__":
    main()
