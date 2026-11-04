# Folder search

Cmd+K searches a shared directory index. Typing does not traverse the disk.
The first use starts background enumeration; later uses reuse memory and a
versioned local cache (`os.UserCacheDir()/Twintrail/folders-v1.gob`). Cache files
are private, atomically replaced, and contain directory paths only.

Enumeration is breadth-first, starting with the active directory and its parent,
recent directories, the home directory, then filesystem roots (all drive letters
on Windows). Hidden directories and dependency folders are included, except `.git` and its
entire subtree. The same exclusion filters old cached paths and explicit scan
roots inside `.git`. Symlink
folders are searchable but not followed, preventing cycles; their physical
locations are indexed through filesystem roots. Linux kernel pseudo-filesystems
and macOS's duplicate `/System/Volumes/Data` tree aren't traversed. Access-denied
locations cannot be enumerated. Initial results are incomplete until indexing
finishes, and the footer says so.

A query is normalized once and compared with pre-normalized index records.
Ranking prefers exact directory names, name prefixes, contained names, fuzzy
names, then paths. A 100-entry max heap selects the best results from **the entire
index**: 100 limits presentation, never enumeration. Queries run off the UI thread,
are debounced by 75 ms, and are canceled when superseded. While indexing, the
current query checks index revisions every 250 ms; unchanged results aren't
replaced. The fixed-height dialog is unaffected by result count.

A new process validates its cached index with a fresh scan. Within a process,
opening the picker starts another scan if the previous one finished more than
five minutes ago. Cmd+R inside the picker forces a refresh. Deleted cache entries
are removed after the scan, and opening any result checks that it still exists.
This is eventual consistency, not an OS filesystem watcher. It avoids permanent
per-directory watchers and does not require Spotlight (which can be disabled).

Validation: `go test ./...`, `go test -race ./...`, `go tool mygo vet .`.
Benchmark: `go test -run '^$' -bench BenchmarkFolderSearch100K -benchmem`.
