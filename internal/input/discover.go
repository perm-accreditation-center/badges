package input

import (
	"os"
	"path/filepath"
	"strings"

	"badges/internal/model"
)

// Discover returns DOCX files placed directly in dir and explains skipped files.
func Discover(dir string) ([]string, []model.Diagnostic) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, []model.Diagnostic{{Severity: model.Error, Stage: "discovery", Source: dir, Code: "INPUT_READ", Message: err.Error()}}
	}

	files := make([]string, 0)
	diagnostics := make([]model.Diagnostic, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "~$") || !strings.EqualFold(filepath.Ext(name), ".docx") {
			diagnostics = append(diagnostics, model.Diagnostic{Severity: model.Info, Stage: "discovery", Source: name, Code: "IGNORED_FILE", Message: "Файл не является входным DOCX"})
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	return files, diagnostics
}
