import threading
import time


class IdGenerator:
    _instance = None
    _instance_lock = threading.Lock()

    def __init__(self) -> None:
        time.sleep(0.02)  # simulate expensive initialisation (loading config); keep this
        self._counter = 0
        self._lock = threading.Lock()

    @classmethod
    def get_instance(cls) -> "IdGenerator":
        # Double-checked locking: the fast path skips the lock once the instance exists.
        # (The GIL makes the unlocked read safe in CPython; the lock makes creation happen once.)
        if cls._instance is None:
            with cls._instance_lock:
                if cls._instance is None:
                    cls._instance = cls()
        return cls._instance

    def next_id(self) -> int:
        # `+=` is several bytecodes (load, add, store): a thread switch in between loses an update.
        with self._lock:
            self._counter += 1
            return self._counter

    def next_id_with_prefix(self, prefix: str) -> str:
        return "%s-%06d" % (prefix, self.next_id())
