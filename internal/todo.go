// Package todo provides the core file-system operations for the todo CLI.
// All mutations live here so they can be unit-tested without touching a real
// file system – callers inject an FS abstraction via the Store interface.
package internal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	finishedDir = "Finished"
	ext         = ".md"
)

// FS is the subset of file-system operations the store needs.
// Swapping this out in tests avoids touching the real disk.
type FS interface {
	MkdirAll(path string, perm fs.FileMode) error
	ReadDir(name string) ([]fs.DirEntry, error)
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	Remove(name string) error
	Rename(oldpath, newpath string) error
	Stat(name string) (fs.FileInfo, error)
}

// RealFS delegates to the real os package.
type RealFS struct{}

func (RealFS) MkdirAll(path string, perm fs.FileMode) error      { return os.MkdirAll(path, perm) }
func (RealFS) ReadDir(name string) ([]fs.DirEntry, error)        { return os.ReadDir(name) }
func (RealFS) ReadFile(name string) ([]byte, error)              { return os.ReadFile(name) }
func (RealFS) WriteFile(n string, d []byte, p fs.FileMode) error { return os.WriteFile(n, d, p) }
func (RealFS) Remove(name string) error                          { return os.Remove(name) }
func (RealFS) Rename(o, n string) error                          { return os.Rename(o, n) }
func (RealFS) Stat(name string) (fs.FileInfo, error)             { return os.Stat(name) }

// Store handles all todo persistence.
type Store struct {
	Dir string
	FS  FS
	Now func() time.Time // injectable for tests
}

// NewStore returns a Store rooted at dir using the real file system.
func NewStore(dir string) *Store {
	return &Store{Dir: dir, FS: RealFS{}, Now: time.Now}
}

// EnsureDir creates the todo directory if it doesn't exist.
func (s *Store) EnsureDir() error {
	return s.FS.MkdirAll(s.Dir, 0o755)
}

// filePath returns the full path for a todo file, stripping any .md suffix
// from the supplied title so callers can be sloppy.
func (s *Store) filePath(title string) string {
	base := strings.TrimSuffix(title, ext)
	return filepath.Join(s.Dir, base+ext)
}

// template renders the markdown template for a new todo file.
func template(title, description string) []byte {
	return []byte(fmt.Sprintf("# %s\n\n## Description\n%s\n\n## Sub-Tasks\n", title, description))
}

// exists returns true when the given path is present on the FS.
func (s *Store) exists(path string) bool {
	_, err := s.FS.Stat(path)
	return err == nil
}

// List returns the names (without extension) of all todos in the directory.
// The Finished sub-directory is excluded.
func (s *Store) List() ([]string, error) {
	entries, err := s.FS.ReadDir(s.Dir)
	if err != nil {
		return nil, fmt.Errorf("reading todo dir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ext) {
			names = append(names, strings.TrimSuffix(e.Name(), ext))
		}
	}
	return names, nil
}

// Description returns the description body of the named todo, or "" if absent.
func (s *Store) Description(title string) (string, error) {
	data, err := s.FS.ReadFile(s.filePath(title))
	if err != nil {
		return "", err
	}
	const marker = "## Description\n"
	i := strings.Index(string(data), marker)
	if i < 0 {
		return "", nil
	}
	rest := string(data)[i+len(marker):]
	if j := strings.Index(rest, "\n##"); j >= 0 {
		rest = rest[:j]
	}
	return strings.TrimSpace(rest), nil
}

// Create writes a new todo file; if one already exists it is left unchanged.
// Returns the file path and whether the file was newly created.
func (s *Store) Create(title, description string) (path string, created bool, err error) {
	path = s.filePath(title)
	if s.exists(path) {
		return path, false, nil
	}
	if err = s.FS.WriteFile(path, template(title, description), 0o644); err != nil {
		return "", false, fmt.Errorf("writing todo file: %w", err)
	}
	return path, true, nil
}

// AddSubTask appends a GFM checkbox item to the named todo, creating the file
// first (with an empty description) if it does not yet exist.
func (s *Store) AddSubTask(title, task string) error {
	path := s.filePath(title)
	base := strings.TrimSuffix(title, ext)
	if !s.exists(path) {
		if err := s.FS.WriteFile(path, template(base, ""), 0o644); err != nil {
			return fmt.Errorf("creating todo file for sub-task: %w", err)
		}
	}
	line := fmt.Sprintf("- [ ] %s\n", task)
	f, err := s.FS.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading todo file: %w", err)
	}
	return s.FS.WriteFile(path, append(f, []byte(line)...), 0o644)
}

// FilePath exposes the resolved file path for a title (used by `edit`).
func (s *Store) FilePath(title string) string {
	return s.filePath(title)
}

// Finish moves the todo to ~/ToDo/Finished/YYYY.MM.DD<title>.md.
func (s *Store) Finish(title string) error {
	src := s.filePath(title)
	if !s.exists(src) {
		return fmt.Errorf("todo %q not found", title)
	}
	base := strings.TrimSuffix(filepath.Base(src), ext)
	finDir := filepath.Join(s.Dir, finishedDir)
	if err := s.FS.MkdirAll(finDir, 0o755); err != nil {
		return fmt.Errorf("creating Finished dir: %w", err)
	}
	date := s.Now().Format("2006.01.02")
	dst := filepath.Join(finDir, date+base+ext)
	if err := s.FS.Rename(src, dst); err != nil {
		return fmt.Errorf("moving todo to Finished: %w", err)
	}
	return nil
}

// Remove deletes a todo file without archiving it.
func (s *Store) Remove(title string) error {
	path := s.filePath(title)
	if !s.exists(path) {
		return fmt.Errorf("todo %q not found", title)
	}
	if err := s.FS.Remove(path); err != nil {
		return fmt.Errorf("deleting todo file: %w", err)
	}
	return nil
}
