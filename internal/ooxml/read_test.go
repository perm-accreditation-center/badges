package ooxml_test

import (
	"strings"
	"testing"

	"badges/internal/ooxml"
	"badges/internal/testdocx"
)

func TestReadDocumentIgnoresDeletedText(t *testing.T) {
	path := testdocx.Write(t, `<w:document><w:body><w:p><w:del><w:delText>Удалённая Фамилия</w:delText></w:del><w:r><w:t>Актуальная Фамилия</w:t></w:r></w:p></w:body></w:document>`)
	document, err := ooxml.ReadDocument(path)
	if err != nil { t.Fatal(err) }
	text := document.Text()
	if strings.Contains(text, "Удалённая") || !strings.Contains(text, "Актуальная") { t.Fatalf("wrong visible text: %q", text) }
}
