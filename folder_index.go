package main

import (
	"container/heap"
	"context"
	"encoding/gob"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type indexedFolder struct{ path, lower, name string }

func folderRecord(path string) indexedFolder {
	return indexedFolder{path, strings.ToLower(filepath.ToSlash(path)), strings.ToLower(filepath.Base(path))}
}

// One index belongs to the window, not to a query. Canceling typing never
// discards enumeration progress. Only directories are stored, never file data.
type folderIndex struct {
	mu       sync.RWMutex
	folders  []indexedFolder
	seen     map[string]bool
	revision uint64
	building bool
	finished time.Time
	cancel   context.CancelFunc
	cache    string
}

func newFolderIndex(cache string) *folderIndex {
	return &folderIndex{seen: map[string]bool{}, cache: cache}
}
func (idx *folderIndex) stop() {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	if idx.cancel != nil {
		idx.cancel()
	}
}

// Check path components so .github and names such as .git-tools remain searchable.
func excludedFolder(path string) bool {
	path = filepath.ToSlash(filepath.Clean(path))
	return path == ".git" || strings.HasPrefix(path, ".git/") || strings.HasSuffix(path, "/.git") || strings.Contains(path, "/.git/")
}

func (idx *folderIndex) add(paths []string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	for _, path := range paths {
		if excludedFolder(path) {
			continue
		}
		if !idx.seen[path] {
			idx.seen[path] = true
			idx.folders = append(idx.folders, folderRecord(path))
			idx.revision++
		}
	}
}
func (idx *folderIndex) refresh(roots []string, force bool) {
	idx.mu.Lock()
	if idx.building || (!force && !idx.finished.IsZero() && time.Since(idx.finished) < 5*time.Minute) {
		idx.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	idx.cancel = cancel
	idx.building = true
	idx.revision++
	idx.mu.Unlock()
	go func() {
		// Cache is only a warm start; a fresh traversal reconciles deletions.
		idx.mu.RLock()
		empty := len(idx.folders) == 0
		idx.mu.RUnlock()
		if empty {
			idx.readCache()
		}
		fresh := []string{}
		scanFolders(ctx, roots, func(paths []string) { fresh = append(fresh, paths...); idx.add(paths) })
		if ctx.Err() != nil {
			return
		}
		records := make([]indexedFolder, 0, len(fresh))
		seen := make(map[string]bool, len(fresh))
		for _, path := range fresh {
			if !seen[path] {
				seen[path] = true
				records = append(records, folderRecord(path))
			}
		}
		idx.writeCache(fresh)
		idx.mu.Lock()
		idx.folders = records
		idx.seen = seen
		idx.building = false
		idx.finished = time.Now()
		idx.revision++
		idx.mu.Unlock()
	}()
}

// Breadth-first traversal reaches ordinary folders before deep dependency
// caches. Roots may overlap; each physical path is enumerated once. Symlink
// directories are searchable but aren't followed, avoiding cycles.
func scanFolders(ctx context.Context, roots []string, publish func([]string)) {
	type directory struct {
		path     string
		knownDir bool
	}
	queue := make([]directory, 0, len(roots))
	for _, root := range roots {
		if excludedFolder(root) {
			continue
		}
		queue = append(queue, directory{path: root})
	}
	seen := map[string]bool{}
	batch := make([]string, 0, 256)
	last := time.Now()
	flush := func() {
		if len(batch) > 0 {
			publish(batch)
			batch = make([]string, 0, 256)
			last = time.Now()
		}
	}
	for head := 0; head < len(queue); head++ {
		if ctx.Err() != nil {
			return
		}
		item := queue[head]
		path := filepath.Clean(item.path)
		queue[head] = directory{}
		if seen[path] {
			continue
		}
		seen[path] = true
		symlink := false
		if !item.knownDir {
			info, err := os.Lstat(path)
			if err != nil {
				continue
			}
			symlink = info.Mode()&os.ModeSymlink != 0
			if symlink {
				info, err = os.Stat(path)
			}
			if err != nil || !info.IsDir() {
				continue
			}
		}
		batch = append(batch, path)
		if len(batch) >= 256 || time.Since(last) > 100*time.Millisecond {
			flush()
		}
		// These are virtual kernel trees or macOS's duplicate Data-volume mount.
		if (runtime.GOOS == "linux" && (path == "/proc" || path == "/sys" || path == "/dev")) || (runtime.GOOS == "darwin" && (path == "/dev" || path == "/System/Volumes/Data")) {
			continue
		}
		if symlink {
			continue
		}
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		for {
			if ctx.Err() != nil {
				f.Close()
				return
			}
			entries, readErr := f.ReadDir(256)
			for _, e := range entries {
				if e.Name() == ".git" {
					continue
				}
				if e.IsDir() || e.Type()&os.ModeSymlink != 0 {
					queue = append(queue, directory{path: filepath.Join(path, e.Name()), knownDir: e.IsDir()})
				}
			}
			if readErr != nil {
				break
			}
		}
		f.Close()
	}
	flush()
}

func folderSearchRoots(priority []string) []string {
	roots := append([]string(nil), priority...)
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, home)
	}
	if runtime.GOOS == "windows" {
		for drive := 'A'; drive <= 'Z'; drive++ {
			roots = append(roots, string(drive)+":\\")
		}
	} else {
		roots = append(roots, "/")
	}
	return roots
}
func folderCachePath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "Twintrail", "folders-v1.gob")
}
func (idx *folderIndex) readCache() {
	if idx.cache == "" {
		return
	}
	f, err := os.Open(idx.cache)
	if err != nil {
		return
	}
	defer f.Close()
	var paths []string
	if gob.NewDecoder(f).Decode(&paths) == nil {
		idx.add(paths)
	}
}
func (idx *folderIndex) writeCache(paths []string) {
	if idx.cache == "" {
		return
	}
	dir := filepath.Dir(idx.cache)
	if os.MkdirAll(dir, 0700) != nil {
		return
	}
	f, err := os.CreateTemp(dir, "folders-*.tmp")
	if err != nil {
		return
	}
	defer os.Remove(f.Name())
	err = gob.NewEncoder(f).Encode(paths)
	closeErr := f.Close()
	if err == nil && closeErr == nil {
		_ = os.Rename(f.Name(), idx.cache)
	}
}

// Query normalization happens once; indexed paths are normalized once at insert.
func folderRank(f indexedFolder, query string) int {
	if f.name == query {
		return 0
	}
	if strings.HasPrefix(f.name, query) {
		return 1
	}
	if strings.Contains(f.name, query) {
		return 2
	}
	if subsequence(f.name, query) {
		return 3
	}
	if strings.Contains(f.lower, query) {
		return 4
	}
	if strings.Contains(query, "/") && subsequence(f.lower, query) {
		return 5
	}
	return -1
}
func subsequence(s, q string) bool {
	for _, r := range q {
		i := strings.IndexRune(s, r)
		if i < 0 {
			return false
		}
		s = s[i+len(string(r)):]
	}
	return true
}

type folderMatch struct {
	path string
	rank int
}

func betterFolder(a, b folderMatch) bool {
	if a.rank != b.rank {
		return a.rank < b.rank
	}
	if len(a.path) != len(b.path) {
		return len(a.path) < len(b.path)
	}
	return a.path < b.path
}

type folderMatches []folderMatch

func (h folderMatches) Len() int           { return len(h) }
func (h folderMatches) Less(i, j int) bool { return betterFolder(h[j], h[i]) }
func (h folderMatches) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *folderMatches) Push(x any)        { *h = append(*h, x.(folderMatch)) }
func (h *folderMatches) Pop() any          { old := *h; x := old[len(old)-1]; *h = old[:len(old)-1]; return x }
func (idx *folderIndex) search(ctx context.Context, query string) ([]string, uint64, bool) {
	query = strings.ToLower(filepath.ToSlash(strings.TrimSpace(query)))
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	best := folderMatches{}
	for i, f := range idx.folders {
		if i%1024 == 0 && ctx.Err() != nil {
			return nil, idx.revision, idx.building
		}
		rank := folderRank(f, query)
		if rank < 0 {
			continue
		}
		m := folderMatch{f.path, rank}
		if len(best) < 100 {
			heap.Push(&best, m)
		} else if betterFolder(m, best[0]) {
			best[0] = m
			heap.Fix(&best, 0)
		}
	}
	sort.Slice(best, func(i, j int) bool { return betterFolder(best[i], best[j]) })
	paths := make([]string, len(best))
	for i, m := range best {
		paths[i] = m.path
	}
	return paths, idx.revision, idx.building
}
