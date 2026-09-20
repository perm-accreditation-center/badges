package spo_test

import (
	"testing"

	"badges/internal/ooxml"
	"badges/internal/spo"
	"badges/internal/testdocx"
)

func TestParsePrefersSpecialtyWrittenInStatusAndRecognizesJoinedAPK(t *testing.T) {
	path := testdocx.Write(t, `<w:document><w:body>
		<w:p><w:r><w:t>График работы подкомиссии по специальности «Физиотерапия»</w:t></w:r></w:p>
		<w:tbl>
			<w:tr><w:tc><w:p><w:r><w:t>ФИО эксперта</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>Статус в составе подкомиссии</w:t></w:r></w:p></w:tc></w:tr>
			<w:tr><w:tc><w:p><w:r><w:t>Зарипова Марина Александровна</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>Член АПК«Медицинский массаж»</w:t></w:r></w:p></w:tc></w:tr>
			<w:tr><w:tc><w:p><w:r><w:t>Яковлева Татьяна Александровна</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>Член АПК Медицинский массаж ОВЗ» ПГМУ</w:t></w:r></w:p></w:tc></w:tr>
			<w:tr><w:tc><w:p><w:r><w:t>Смирнова Ольга Сергеевна</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>СекретарьАПК «Техподдержки»</w:t></w:r></w:p></w:tc></w:tr>
		</w:tbl>
	</w:body></w:document>`)
	document, err := ooxml.ReadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	records, diagnostics := spo.Parse(document, path)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if len(records) != 3 {
		t.Fatalf("expected every commission member to be parsed, got %#v", records)
	}
	want := []string{"Медицинский массаж", "Медицинский массаж ОВЗ", "Техподдержки"}
	for i, specialty := range want {
		if records[i].Specialty != specialty {
			t.Fatalf("record %d specialty: got %q, want %q", i, records[i].Specialty, specialty)
		}
	}
}

func TestParseReportsNamedMemberWhoseRoleCannotBeRecognized(t *testing.T) {
	path := testdocx.Write(t, `<w:document><w:body>
		<w:p><w:r><w:t>График работы подкомиссии по специальности «Физиотерапия»</w:t></w:r></w:p>
		<w:tbl>
			<w:tr><w:tc><w:p><w:r><w:t>ФИО эксперта</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>Статус в составе подкомиссии</w:t></w:r></w:p></w:tc></w:tr>
			<w:tr><w:tc><w:p><w:r><w:t>Орлова Наталья Сергеевна</w:t></w:r></w:p></w:tc><w:tc><w:p><w:r><w:t>Куратор подкомиссии</w:t></w:r></w:p></w:tc></w:tr>
		</w:tbl>
	</w:body></w:document>`)
	document, err := ooxml.ReadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	records, diagnostics := spo.Parse(document, path)
	if len(records) != 0 {
		t.Fatalf("unrecognized role must not become a silently wrong badge: %#v", records)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != "ROLE_UNRECOGNIZED" {
		t.Fatalf("expected ROLE_UNRECOGNIZED warning, got %#v", diagnostics)
	}
}
