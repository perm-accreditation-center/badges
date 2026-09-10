package input_test

import (
	"os"
	"path/filepath"
	"testing"

	"badges/internal/input"
)

func TestDiscoverSkipsWordLockAndNonDocx(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"schedule.docx", "~$schedule.docx", "note.pdf"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0600); err != nil { t.Fatal(err) }
	}
	files, diagnostics := input.Discover(dir)
	if len(files) != 1 || filepath.Base(files[0]) != "schedule.docx" || len(diagnostics) != 2 {
		t.Fatalf("unexpected discovery result: %#v %#v", files, diagnostics)
	}
}
