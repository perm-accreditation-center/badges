package model_test

import (
	"strings"
	"testing"

	"badges/internal/model"
)

func TestDiagnosticErrorIncludesStageAndSource(t *testing.T) {
	diagnostic := model.Diagnostic{
		Severity: model.Error,
		Stage:    "parse",
		Source:   "input/a.docx",
		Message:  "ФИО не найдено",
	}

	got := diagnostic.String()
	if !strings.Contains(got, "parse") || !strings.Contains(got, "input/a.docx") {
		t.Fatalf("diagnostic must be actionable, got %q", got)
	}
}
