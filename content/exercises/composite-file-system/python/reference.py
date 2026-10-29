from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Dict, Iterator, List


class Node(ABC):
    def __init__(self, name: str) -> None:
        self.name = name

    @abstractmethod
    def size(self) -> int: ...


class FileNode(Node):
    def __init__(self, name: str, size: int) -> None:
        super().__init__(name)
        self._size = size

    def size(self) -> int:
        return self._size


class DirNode(Node):
    def __init__(self, name: str) -> None:
        super().__init__(name)
        self.children: Dict[str, Node] = {}

    def size(self) -> int:
        return sum(c.size() for c in self.children.values())  # the Composite recursion

    def sorted_children(self) -> List[Node]:
        return [self.children[k] for k in sorted(self.children)]


def _split(path: str) -> List[str]:
    if not path or not path.startswith("/"):
        raise ValueError("absolute path required: %r" % path)
    return [p for p in path.split("/") if p]


class FileSystem:
    def __init__(self) -> None:
        self._root = DirNode("")

    def mkdir(self, path: str) -> None:
        cur = self._root
        for part in _split(path):
            nxt = cur.children.get(part)
            if nxt is None:
                nxt = cur.children[part] = DirNode(part)
            elif not isinstance(nxt, DirNode):
                raise NotADirectoryError("file exists at " + part)
            cur = nxt

    def create_file(self, path: str, size: int) -> None:
        parts = _split(path)
        if not parts:
            raise ValueError("invalid path")
        parent = self._dir(parts[:-1])
        name = parts[-1]
        if name in parent.children:
            raise FileExistsError(path)
        parent.children[name] = FileNode(name, size)

    def size(self, path: str) -> int:
        return self._resolve(_split(path)).size()

    def ls(self, path: str) -> List[str]:
        node = self._resolve(_split(path))
        if isinstance(node, FileNode):
            return [node.name]
        return [c.name + "/" if isinstance(c, DirNode) else c.name for c in node.sorted_children()]

    def find(self, path: str, ext: str) -> List[str]:
        parts = _split(path)
        start = self._resolve(parts)
        prefix = "/" + "/".join(parts) if parts else ""
        suffix = "." + ext
        return sorted(p for p in self._walk(start, prefix) if p.endswith(suffix))

    def delete(self, path: str) -> None:
        parts = _split(path)
        if not parts:
            raise ValueError("cannot delete root")
        parent = self._dir(parts[:-1])
        if parent.children.pop(parts[-1], None) is None:
            raise FileNotFoundError(path)

    # ---- helpers ----
    def _walk(self, node: Node, full_path: str) -> Iterator[str]:
        """Yields the full path of every file under node."""
        if isinstance(node, FileNode):
            yield full_path
            return
        for c in node.sorted_children():
            yield from self._walk(c, full_path + "/" + c.name)

    def _resolve(self, parts: List[str]) -> Node:
        cur: Node = self._root
        for p in parts:
            if not isinstance(cur, DirNode) or p not in cur.children:
                raise FileNotFoundError(p)
            cur = cur.children[p]
        return cur

    def _dir(self, parts: List[str]) -> DirNode:
        node = self._resolve(parts)
        if not isinstance(node, DirNode):
            raise NotADirectoryError("/" + "/".join(parts))
        return node
