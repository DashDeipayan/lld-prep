from __future__ import annotations

from typing import List


class FileSystem:
    # TODO: Node / FileNode / DirNode (Composite)

    def mkdir(self, path: str) -> None:
        raise NotImplementedError("TODO")

    def create_file(self, path: str, size: int) -> None:
        raise NotImplementedError("TODO")

    def size(self, path: str) -> int:
        raise NotImplementedError("TODO")

    def ls(self, path: str) -> List[str]:
        raise NotImplementedError("TODO")

    def find(self, path: str, ext: str) -> List[str]:
        raise NotImplementedError("TODO")

    def delete(self, path: str) -> None:
        raise NotImplementedError("TODO")
