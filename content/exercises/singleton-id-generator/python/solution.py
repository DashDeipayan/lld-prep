import time


# Current production code: works single-threaded, breaks under load.
class IdGenerator:
    _instance = None

    def __init__(self) -> None:
        time.sleep(0.02)  # simulate expensive initialisation (loading config); keep this
        self._counter = 0

    @classmethod
    def get_instance(cls) -> "IdGenerator":
        if cls._instance is None:
            cls._instance = cls()
        return cls._instance

    def next_id(self) -> int:
        value = self._counter
        value += 1
        self._counter = value
        return value

    def next_id_with_prefix(self, prefix: str) -> str:
        return "%s-%d" % (prefix, self.next_id())
