package internal_test

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/h3jfc/todo/internal"
)

// ---- fake FS ----------------------------------------------------------

type memFile struct {
	data  []byte
	isDir bool
}

type fakeFS struct {
	files map[string]*memFile
}

func newFakeFS() *fakeFS { return &fakeFS{files: map[string]*memFile{}} }

func (f *fakeFS) MkdirAll(path string, _ fs.FileMode) error {
	f.files[path] = &memFile{isDir: true}
	return nil
}

func (f *fakeFS) ReadDir(name string) ([]fs.DirEntry, error) {
	var entries []fakeEntry
	for k, v := range f.files {
		dir := filepath.Dir(k)
		if dir == name && k != name {
			entries = append(entries, fakeEntry{name: filepath.Base(k), isDir: v.isDir})
		}
	}
	result := make([]fs.DirEntry, len(entries))
	for i, e := range entries {
		result[i] = e
	}
	return result, nil
}

func (f *fakeFS) ReadFile(name string) ([]byte, error) {
	if mf, ok := f.files[name]; ok {
		cp := make([]byte, len(mf.data))
		copy(cp, mf.data)
		return cp, nil
	}
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

func (f *fakeFS) WriteFile(name string, data []byte, _ fs.FileMode) error {
	cp := make([]byte, len(data))
	copy(cp, data)
	f.files[name] = &memFile{data: cp}
	return nil
}

func (f *fakeFS) Remove(name string) error {
	if _, ok := f.files[name]; !ok {
		return &fs.PathError{Op: "remove", Path: name, Err: fs.ErrNotExist}
	}
	delete(f.files, name)
	return nil
}

func (f *fakeFS) Rename(old, new string) error {
	mf, ok := f.files[old]
	if !ok {
		return &fs.PathError{Op: "rename", Path: old, Err: fs.ErrNotExist}
	}
	f.files[new] = mf
	delete(f.files, old)
	return nil
}

func (f *fakeFS) Stat(name string) (fs.FileInfo, error) {
	if mf, ok := f.files[name]; ok {
		return fakeInfo{name: filepath.Base(name), isDir: mf.isDir}, nil
	}
	return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
}

// fakeEntry implements fs.DirEntry
type fakeEntry struct {
	name  string
	isDir bool
}

func (e fakeEntry) Name() string               { return e.name }
func (e fakeEntry) IsDir() bool                { return e.isDir }
func (e fakeEntry) Type() fs.FileMode          { return 0 }
func (e fakeEntry) Info() (fs.FileInfo, error) { return fakeInfo{name: e.name, isDir: e.isDir}, nil }

// fakeInfo implements fs.FileInfo
type fakeInfo struct {
	name  string
	isDir bool
}

func (i fakeInfo) Name() string       { return i.name }
func (i fakeInfo) Size() int64        { return 0 }
func (i fakeInfo) Mode() fs.FileMode  { return 0o644 }
func (i fakeInfo) ModTime() time.Time { return time.Time{} }
func (i fakeInfo) IsDir() bool        { return i.isDir }
func (i fakeInfo) Sys() interface{}   { return nil }

// ---- helpers ----------------------------------------------------------

func newStore(t *testing.T) (*internal.Store, *fakeFS) {
	t.Helper()
	fake := newFakeFS()
	s := &internal.Store{
		Dir: "/todos",
		FS:  fake,
		Now: func() time.Time { return time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC) },
	}
	return s, fake
}

// ---- tests ------------------------------------------------------------

func TestCreate_NewFile(t *testing.T) {
	s, fake := newStore(t)
	path, created, err := s.Create("buy-milk", "get oat milk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Error("expected created=true for new file")
	}
	data, ok := fake.files[path]
	if !ok {
		t.Fatalf("file not found at %s", path)
	}
	content := string(data.data)
	if !strings.Contains(content, "# buy-milk") {
		t.Errorf("expected title in content, got: %s", content)
	}
	if !strings.Contains(content, "get oat milk") {
		t.Errorf("expected description in content, got: %s", content)
	}
	if !strings.Contains(content, "## Sub-Tasks") {
		t.Errorf("expected Sub-Tasks section in content, got: %s", content)
	}
}

func TestCreate_ExistingFileUnchanged(t *testing.T) {
	s, fake := newStore(t)
	path := "/todos/buy-milk.md"
	original := []byte("existing content")
	fake.files[path] = &memFile{data: original}

	_, created, err := s.Create("buy-milk", "new description")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created {
		t.Error("expected created=false for existing file")
	}
	if string(fake.files[path].data) != string(original) {
		t.Error("existing file should not be modified")
	}
}

func TestCreate_StripsMdSuffix(t *testing.T) {
	s, _ := newStore(t)
	path, _, err := s.Create("buy-milk.md", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/todos/buy-milk.md" {
		t.Errorf("expected /todos/buy-milk.md, got %s", path)
	}
}

func TestList_ReturnsOnlyMdFiles(t *testing.T) {
	s, fake := newStore(t)
	fake.files["/todos"] = &memFile{isDir: true}
	fake.files["/todos/alpha.md"] = &memFile{data: []byte("a")}
	fake.files["/todos/beta.md"] = &memFile{data: []byte("b")}
	fake.files["/todos/README.txt"] = &memFile{data: []byte("c")}
	fake.files["/todos/Finished"] = &memFile{isDir: true}

	names, err := s.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 todos, got %d: %v", len(names), names)
	}
	got := map[string]bool{}
	for _, n := range names {
		got[n] = true
	}
	if !got["alpha"] || !got["beta"] {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestList_EmptyDir(t *testing.T) {
	s, fake := newStore(t)
	fake.files["/todos"] = &memFile{isDir: true}
	names, err := s.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected 0 todos, got %d", len(names))
	}
}

func TestAddSubTask_AppendsToExistingFile(t *testing.T) {
	s, fake := newStore(t)
	path := "/todos/buy-milk.md"
	fake.files[path] = &memFile{data: []byte("# buy-milk\n\n## Sub-Tasks\n")}

	if err := s.AddSubTask("buy-milk", "check expiry dates"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content := string(fake.files[path].data)
	if !strings.Contains(content, "- [ ] check expiry dates") {
		t.Errorf("sub-task not appended, got: %s", content)
	}
}

func TestAddSubTask_CreatesFileIfMissing(t *testing.T) {
	s, fake := newStore(t)
	if err := s.AddSubTask("new-todo", "first task"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	path := "/todos/new-todo.md"
	if _, ok := fake.files[path]; !ok {
		t.Fatal("expected file to be created")
	}
	content := string(fake.files[path].data)
	if !strings.Contains(content, "- [ ] first task") {
		t.Errorf("task missing from new file: %s", content)
	}
}

func TestFinish_MovesToFinishedDir(t *testing.T) {
	s, fake := newStore(t)
	fake.files["/todos/buy-milk.md"] = &memFile{data: []byte("content")}
	fake.files["/todos/Finished"] = &memFile{isDir: true}

	if err := s.Finish("buy-milk"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "/todos/Finished/2024.06.15buy-milk.md"
	if _, ok := fake.files[expected]; !ok {
		t.Errorf("expected finished file at %s", expected)
	}
	if _, ok := fake.files["/todos/buy-milk.md"]; ok {
		t.Error("original file should have been moved")
	}
}

func TestFinish_ErrorIfNotFound(t *testing.T) {
	s, _ := newStore(t)
	if err := s.Finish("nonexistent"); err == nil {
		t.Error("expected error for missing todo")
	}
}

func TestRemove_DeletesFile(t *testing.T) {
	s, fake := newStore(t)
	fake.files["/todos/buy-milk.md"] = &memFile{data: []byte("content")}

	if err := s.Remove("buy-milk"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := fake.files["/todos/buy-milk.md"]; ok {
		t.Error("file should have been deleted")
	}
}

func TestRemove_ErrorIfNotFound(t *testing.T) {
	s, _ := newStore(t)
	if err := s.Remove("nonexistent"); err == nil {
		t.Error("expected error for missing todo")
	}
}

func TestFilePath_StripsSuffix(t *testing.T) {
	s, _ := newStore(t)
	p := s.FilePath("my-task.md")
	if p != "/todos/my-task.md" {
		t.Errorf("unexpected path: %s", p)
	}
	p2 := s.FilePath("my-task")
	if p2 != "/todos/my-task.md" {
		t.Errorf("unexpected path: %s", p2)
	}
}
