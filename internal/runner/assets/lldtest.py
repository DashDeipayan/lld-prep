"""Tiny assertion helpers for tests.py files (the Python twin of Java's T)."""


def check(cond, msg):
    if not cond:
        raise AssertionError(msg)


def eq(expected, actual, msg):
    if expected != actual:
        raise AssertionError("%s\nexpected: %r\n  actual: %r" % (msg, expected, actual))


def near(expected, actual, msg, tol=1e-6):
    if abs(expected - actual) > tol:
        raise AssertionError("%s\nexpected: %r\n  actual: %r" % (msg, expected, actual))


def raises(exc_type, fn, msg):
    """Calls fn() and returns the exception it raised, which must be an exc_type."""
    try:
        fn()
    except exc_type as e:
        return e
    except BaseException as e:  # noqa: B902
        raise AssertionError("%s\nexpected %s but got %s: %s" % (msg, exc_type.__name__, type(e).__name__, e))
    raise AssertionError("%s\nexpected %s to be raised" % (msg, exc_type.__name__))
