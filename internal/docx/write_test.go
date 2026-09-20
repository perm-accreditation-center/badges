package docx

import (
	"archive/zip"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"badges/internal/model"
)

func TestWritePrintsOnlyActualRoleSpecialtyAssignments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "badges.docx")
	people := []model.Person{{
		FullName: "Калинина Марина Валерьевна",
		Assignments: []model.Assignment{
			{Role: "Член подкомиссии", Specialty: "Медицинский массаж"},
			{Role: "Председатель подкомиссии", Specialty: "Реабилитационное сестринское дело"},
		},
	}}
	if err := Write(path, people); err != nil {
		t.Fatal(err)
	}

	xml := documentXML(t, path)
	if got := strings.Count(xml, "Аккредитационная подкомиссия по специальности"); got != 2 {
		t.Fatalf("got %d badges; expected exactly two source assignments", got)
	}
}

func documentXML(t *testing.T, path string) string {
	t.Helper()
	archive, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	for _, file := range archive.File {
		if file.Name != "word/document.xml" {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()
		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		return string(content)
	}
	t.Fatal("word/document.xml not found")
	return ""
}
