package cmd_test

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/h3jfc/todo/cmd"
	"github.com/h3jfc/todo/internal"
)

// ---- fake FS (same as in internal/todo tests, duplicated to avoid coupling) --

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
		if filepath.Dir(k) == name && k != name {
			entries = append(entries, fakeEntry{name: filepath.Base(k), isDir: v.isDir})
		}
	}
	out := make([]fs.DirEntry, len(entries))
	for i, e := range entries {
		out[i] = e
	}
	return out, nil
}

func (f *fakeFS) ReadFile(name string) ([]byte, error) {
	if m, ok := f.files[name]; ok {
		cp := make([]byte, len(m.data))
		copy(cp, m.data)
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

func (f *fakeFS) Rename(o, n string) error {
	m, ok := f.files[o]
	if !ok {
		return &fs.PathError{Op: "rename", Path: o, Err: fs.ErrNotExist}
	}
	f.files[n] = m
	delete(f.files, o)
	return nil
}

func (f *fakeFS) Stat(name string) (fs.FileInfo, error) {
	if m, ok := f.files[name]; ok {
		return fakeInfo{name: filepath.Base(name), isDir: m.isDir}, nil
	}
	return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
}

type fakeEntry struct {
	name  string
	isDir bool
}

func (e fakeEntry) Name() string               { return e.name }
func (e fakeEntry) IsDir() bool                { return e.isDir }
func (e fakeEntry) Type() fs.FileMode          { return 0 }
func (e fakeEntry) Info() (fs.FileInfo, error) { return fakeInfo{name: e.name, isDir: e.isDir}, nil }

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

// ---- test harness -----------------------------------------------------

const testDir = "/todos"

// setupCmd replaces the package-level newStore hook and returns a fake FS
// so tests can inspect the resulting files.
func setupCmd(t *testing.T) *fakeFS {
	t.Helper()
	fake := newFakeFS()
	fake.files[testDir] = &memFile{isDir: true}

	cmd.SetNewStore(func() *internal.Store {
		return &internal.Store{
			Dir: testDir,
			FS:  fake,
			Now: func() time.Time { return time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC) },
		}
	})
	t.Cleanup(func() { cmd.ResetNewStore() })
	return fake
}

// execute runs the CLI with the given args and returns stdout + error.
func execute(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := cmd.New()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

// ---- tests ------------------------------------------------------------

func TestRootNoArgs_ListsEmpty(t *testing.T) {
	setupCmd(t)
	out, err := execute(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No todos") {
		t.Errorf("expected empty-list message, got: %q", out)
	}
}

func TestRootWithTitle_CreatesTodo(t *testing.T) {
	fake := setupCmd(t)
	_, err := execute(t, "buy-milk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := fake.files[testDir+"/buy-milk.md"]; !ok {
		t.Error("expected todo file to be created")
	}
}

func TestRootWithTitleAndDesc_Propagates(t *testing.T) {
	fake := setupCmd(t)
	_, err := execute(t, "buy-milk", "oat milk please")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content := string(fake.files[testDir+"/buy-milk.md"].data)
	if !strings.Contains(content, "oat milk please") {
		t.Errorf("description missing from file: %s", content)
	}
}

func TestInitCmd(t *testing.T) {
	fake := setupCmd(t)
	_, err := execute(t, "init", "read-book", "finish the chapter")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content := string(fake.files[testDir+"/read-book.md"].data)
	if !strings.Contains(content, "finish the chapter") {
		t.Errorf("description missing: %s", content)
	}
}

func TestSubCmd(t *testing.T) {
	fake := setupCmd(t)
	fake.files[testDir+"/buy-milk.md"] = &memFile{data: []byte("# buy-milk\n\n## Sub-Tasks\n")}
	_, err := execute(t, "sub", "buy-milk", "check dates")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content := string(fake.files[testDir+"/buy-milk.md"].data)
	if !strings.Contains(content, "- [ ] check dates") {
		t.Errorf("sub-task missing: %s", content)
	}
}

func TestSubCmd_MissingArgs(t *testing.T) {
	setupCmd(t)
	_, err := execute(t, "sub", "buy-milk")
	if err == nil {
		t.Error("expected error for missing task arg")
	}
}

func TestEditCmd_CreatesFile(t *testing.T) {
	fake := setupCmd(t)
	opened := ""
	cmd.SetEditorFunc(func(path string) error {
		opened = path
		return nil
	})
	t.Cleanup(cmd.ResetEditorFunc)

	_, err := execute(t, "edit", "buy-milk", "oat milk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := fake.files[testDir+"/buy-milk.md"]; !ok {
		t.Error("expected file to be created before opening editor")
	}
	if opened != testDir+"/buy-milk.md" {
		t.Errorf("editor opened wrong path: %s", opened)
	}
}

func TestFinishCmd(t *testing.T) {
	fake := setupCmd(t)
	fake.files[testDir+"/buy-milk.md"] = &memFile{data: []byte("content")}
	fake.files[testDir+"/Finished"] = &memFile{isDir: true}

	_, err := execute(t, "finish", "buy-milk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := testDir + "/Finished/2024.06.15buy-milk.md"
	if _, ok := fake.files[expected]; !ok {
		t.Errorf("expected archived file at %s", expected)
	}
}

func TestFinishCmd_NotFound(t *testing.T) {
	setupCmd(t)
	_, err := execute(t, "finish", "nonexistent")
	if err == nil {
		t.Error("expected error for missing todo")
	}
}

func TestRmCmd(t *testing.T) {
	fake := setupCmd(t)
	fake.files[testDir+"/buy-milk.md"] = &memFile{data: []byte("content")}

	_, err := execute(t, "rm", "buy-milk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := fake.files[testDir+"/buy-milk.md"]; ok {
		t.Error("file should have been deleted")
	}
}

func TestRmCmd_NotFound(t *testing.T) {
	setupCmd(t)
	_, err := execute(t, "rm", "nonexistent")
	if err == nil {
		t.Error("expected error for missing todo")
	}
}

func TestListShowsTodos(t *testing.T) {
	fake := setupCmd(t)
	fake.files[testDir+"/alpha.md"] = &memFile{data: []byte("a")}
	fake.files[testDir+"/beta.md"] = &memFile{data: []byte("b")}

	out, err := execute(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Errorf("expected todos in output, got: %q", out)
	}
}
