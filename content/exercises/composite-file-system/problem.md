Design an in-memory file system. Variants of it are asked at Amazon and Google, and it's the canonical **Composite** problem: a directory contains files *or other directories*, and `size()` works the same on both.

## Paths

Absolute only, `/`-separated, for example `/a/b/c.txt`. The root is `/`. You can assume names never contain `/` and that paths have no trailing slash (other than the root).

## Operations

| Operation | Behaviour | Errors |
|---|---|---|
| `mkdir(path)` | create the directory **and any missing parents** (`mkdir -p`). Fine if it already exists as a directory | a file exists at that path or on the way → *exists / not a dir* |
| `createFile(path, size)` | create a file. The parent directory must exist | parent missing → *not found*. Parent is a file → *not a dir*. Path exists → *exists* |
| `size(path)` | file: its size. Directory: the sum of everything below it | missing → *not found* |
| `ls(path)` | directory: child names **sorted**, with directories suffixed `/`. File: `[name]` | missing → *not found* |
| `find(path, ext)` | full paths of all **files** under `path` ending in `.ext`, **sorted** | missing → *not found* |
| `delete(path)` | remove the file or the whole subtree | missing → *not found*. Deleting `/` → *invalid* |

<div class="lang-note" data-lang="java">

Java errors: *not found* → `NoSuchElementException`. *exists* and *not a dir* → `IllegalStateException`. *invalid* → `IllegalArgumentException`.

</div>

<div class="lang-note" data-lang="go">

Go: `NewFileSystem() *FileSystem` with `Mkdir(p) error`, `CreateFile(p string, size int64) error`, `Size(p) (int64, error)`, `Ls(p) ([]string, error)`, `Find(p, ext string) ([]string, error)`, `Delete(p) error`. Errors: `ErrNotFound`, `ErrExists`, `ErrNotDir`, `ErrInvalidPath`.

</div>

<div class="lang-note" data-lang="python">

Python: `FileSystem()` with `mkdir(p)`, `create_file(p, size)`, `size(p)`, `ls(p)`, `find(p, ext)` and `delete(p)`. Use the built-in exceptions: *not found* → `FileNotFoundError`, *exists* → `FileExistsError`, *not a dir* → `NotADirectoryError`, *invalid* → `ValueError`.

</div>

## Follow-ups

- *Add `mv`.* (Detach a node from its parent map and attach it elsewhere. Guard against moving a directory into itself.)
- *Add permissions or owners.* (Fields on `Node`, checked in `resolve`, or a Proxy around the file system.)
- *Make it thread-safe.* (One `ReadWriteLock` is simplest. Per-directory locks need careful lock ordering.)
