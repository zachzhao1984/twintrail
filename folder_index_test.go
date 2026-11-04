package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestFolderIndexRanksExactBeyondFirstHundred(t *testing.T) {
	idx := newFolderIndex("")
	paths := []string{}
	for i := 0; i < 300; i++ {
		paths = append(paths, fmt.Sprintf("/home/.cache/bucketeer/dependency/%d", i))
	}
	target := "/home/Workspace/bucketeer"
	paths = append(paths, target, "/home/Workspace/Bucketeer-tools")
	idx.add(paths)
	got, _, _ := idx.search(context.Background(), "bucketeer")
	if len(got) != 100 || got[0] != target {
		t.Fatalf("exact match lost beyond first 100: %v", got)
	}
	got, _, _ = idx.search(context.Background(), "bcktr")
	if len(got) == 0 || got[0] != target {
		t.Fatalf("fuzzy name not ranked first: %v", got)
	}
	got, _, _ = idx.search(context.Background(), "workspace/buck")
	if !slices.Contains(got, target) {
		t.Fatalf("path query missed target: %v", got)
	}
}

func TestFolderScanCoverageAndCycles(t *testing.T) {
	root := t.TempDir()
	for _, p := range []string{".cache/deep/bucketeer", "Workspace/bucketeer", "node_modules/hidden-package", "空 格/文件夹"} {
		if err := os.MkdirAll(filepath.Join(root, p), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(root, filepath.Join(root, "Workspace", "cycle")); err != nil {
		t.Fatal(err)
	}
	paths := []string{}
	scanFolders(context.Background(), []string{root, filepath.Join(root, "Workspace")}, func(batch []string) { paths = append(paths, batch...) })
	for _, p := range []string{"Workspace/bucketeer", ".cache/deep/bucketeer", "node_modules/hidden-package", "空 格/文件夹", "Workspace/cycle"} {
		if !slices.Contains(paths, filepath.Join(root, p)) {
			t.Errorf("missing %s", p)
		}
	}
	seen := map[string]bool{}
	for _, p := range paths {
		if seen[p] {
			t.Errorf("duplicate %s", p)
		}
		seen[p] = true
	}
	if len(paths) > 20 {
		t.Fatal("followed symlink cycle")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	scanFolders(ctx, []string{root}, func([]string) { t.Error("canceled scan published") })
}

func TestFolderIndexCacheRefreshRemovesDeletedFolders(t *testing.T) {
	root := t.TempDir()
	cache := filepath.Join(t.TempDir(), "folders.gob")
	idx := newFolderIndex(cache)
	deleted := filepath.Join(root, "deleted")
	idx.writeCache([]string{deleted})
	idx.readCache()
	got, _, _ := idx.search(context.Background(), "deleted")
	if !slices.Contains(got, deleted) {
		t.Fatal("cache not loaded")
	}
	target := filepath.Join(root, "bucketeer")
	os.Mkdir(target, 0700)
	idx.refresh([]string{root}, true)
	defer idx.stop()
	deadline := time.Now().Add(5 * time.Second)
	for {
		_, _, building := idx.search(context.Background(), "bucketeer")
		if !building {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("scan did not finish")
		}
		time.Sleep(time.Millisecond)
	}
	got, _, _ = idx.search(context.Background(), "deleted")
	if slices.Contains(got, deleted) {
		t.Fatal("stale cached path retained")
	}
	got, _, _ = idx.search(context.Background(), "bucketeer")
	if len(got) != 1 || got[0] != target {
		t.Fatal(got)
	}
}

func BenchmarkFolderSearch100K(b *testing.B) {
	idx := newFolderIndex("")
	paths := make([]string, 100000)
	for i := range paths {
		paths[i] = fmt.Sprintf("/Users/test/Workspace/project-%d/node_modules/package-%d", i/100, i)
	}
	paths = append(paths, "/Users/test/Workspace/bucketeer")
	idx.add(paths)
	for _, query := range []string{"bucketeer", "package", "pckg", "workspace/buck"} {
		b.Run(query, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				idx.search(context.Background(), query)
			}
		})
	}
}

func TestFolderIndexExcludesGitTreesAndOldCache(t *testing.T) {
	root := t.TempDir()
	excluded := []string{filepath.Join(root, "repo", ".git"), filepath.Join(root, "repo", ".git", "objects", "pack")}
	kept := []string{filepath.Join(root, "repo", "src"), filepath.Join(root, "repo", ".github"), filepath.Join(root, "repo", ".git-tools")}
	for _, path := range append(slices.Clone(excluded), kept...) {
		if err := os.MkdirAll(path, 0700); err != nil {
			t.Fatal(err)
		}
	}
	var found []string
	// Even an explicitly requested root within .git must be skipped.
	scanFolders(context.Background(), append([]string{root}, excluded...), func(paths []string) { found = append(found, paths...) })
	for _, path := range excluded {
		if slices.Contains(found, path) {
			t.Errorf("scanned excluded directory %s", path)
		}
	}
	for _, path := range kept {
		if !slices.Contains(found, path) {
			t.Errorf("excluded valid directory %s", path)
		}
	}
	idx := newFolderIndex(filepath.Join(t.TempDir(), "old-cache.gob"))
	idx.writeCache(append(slices.Clone(excluded), kept...))
	idx.readCache()
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	for _, path := range excluded {
		if idx.seen[path] {
			t.Errorf("loaded excluded cached directory %s", path)
		}
	}
	for _, path := range kept {
		if !idx.seen[path] {
			t.Errorf("lost valid cached directory %s", path)
		}
	}
}
