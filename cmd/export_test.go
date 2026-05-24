package cmd

import "github.com/h3jfc/todo/internal"

// SetNewStore replaces the store constructor. Call in tests only.
func SetNewStore(fn func() *internal.Store) {
	newStore = fn
}

// ResetNewStore restores the default store constructor.
func ResetNewStore() {
	newStore = func() *internal.Store {
		// re-uses the default closure defined in root.go via the package-level var
		// We reset to a no-op that will panic if actually called after cleanup,
		// but tests set it again in each sub-test via SetNewStore.
		panic("newStore not set – call SetNewStore in your test")
	}
}

// SetEditorFunc replaces the editor launcher. Call in tests only.
func SetEditorFunc(fn func(path string) error) {
	editorFunc = fn
}

// ResetEditorFunc restores the real editor launcher.
func ResetEditorFunc() {
	editorFunc = openEditor
}
