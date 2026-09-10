package testdocx

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func Write(t *testing.T, documentXML string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture.docx")
	if err := WriteFile(path, documentXML); err != nil {
		t.Fatal(err)
	}
	return path
}

func WriteFile(path, documentXML string) error {
	file, err := os.Create(path)
	if err != nil { return err }
	defer file.Close()
	writer := zip.NewWriter(file)
	entry, err := writer.Create("word/document.xml")
	if err != nil { return err }
	if _, err = entry.Write([]byte(documentXML)); err != nil { return err }
	return writer.Close()
}
