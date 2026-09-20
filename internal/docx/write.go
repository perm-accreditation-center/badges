package docx

import (
	"archive/zip"
	"badges/internal/model"
	"fmt"
	"html"
	"os"
	"strings"
)

func Write(path string, people []model.Person) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	z := zip.NewWriter(f)
	defer z.Close()
	parts := map[string]string{"[Content_Types].xml": `<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/></Types>`, "_rels/.rels": `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`, "word/_rels/document.xml.rels": `<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"/>`, "word/styles.xml": `<?xml version="1.0"?><w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"/>`, "word/document.xml": document(people)}
	for n, s := range parts {
		w, x := z.Create(n)
		if x != nil {
			return x
		}
		if _, x = w.Write([]byte(s)); x != nil {
			return x
		}
	}
	return nil
}
func document(p []model.Person) string {
	p = expand(p)
	var b strings.Builder
	b.WriteString(`<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	for start := 0; start < len(p) || start == 0; start += 10 {
		end := start + 10
		if end > len(p) {
			end = len(p)
		}
		b.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="10204" w:type="dxa"/><w:tblLayout w:type="fixed"/><w:tblCellMar><w:top w:w="100" w:type="dxa"/><w:left w:w="100" w:type="dxa"/><w:bottom w:w="100" w:type="dxa"/><w:right w:w="100" w:type="dxa"/></w:tblCellMar><w:tblBorders><w:top w:val="single" w:sz="4"/><w:left w:val="single" w:sz="4"/><w:bottom w:val="single" w:sz="4"/><w:right w:val="single" w:sz="4"/><w:insideH w:val="single" w:sz="4"/><w:insideV w:val="single" w:sz="4"/></w:tblBorders></w:tblPr><w:tblGrid><w:gridCol w:w="5102"/><w:gridCol w:w="5102"/></w:tblGrid>`)
		rows := 5
		if end-start < 10 {
			rows = (end - start + 1) / 2
		}
		for r := 0; r < rows; r++ {
			b.WriteString(`<w:tr><w:trPr><w:trHeight w:val="3118" w:hRule="exact"/></w:trPr>`)
			for c := 0; c < 2; c++ {
				i := start + r*2 + c
				b.WriteString(`<w:tc><w:tcPr><w:tcW w:w="5102" w:type="dxa"/>`)
				if i >= end {
					b.WriteString(`<w:tcBorders><w:top w:val="nil"/><w:left w:val="nil"/><w:bottom w:val="nil"/><w:right w:val="nil"/></w:tcBorders>`)
				}
				b.WriteString(`</w:tcPr>`)
				if i < end {
					b.WriteString(card(p[i]))
				} else {
					b.WriteString(`<w:p/>`)
				}
				b.WriteString(`</w:tc>`)
			}
			b.WriteString(`</w:tr>`)
		}
		b.WriteString(`</w:tbl>`)
		if end == len(p) {
			break
		}
		b.WriteString(`<w:p><w:pPr><w:spacing w:before="0" w:after="0" w:line="20" w:lineRule="exact"/></w:pPr><w:r><w:rPr><w:sz w:val="2"/></w:rPr><w:t></w:t></w:r></w:p>`)
	}
	b.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="500" w:right="850" w:bottom="500" w:left="850"/></w:sectPr></w:body></w:document>`)
	return b.String()
}
func card(p model.Person) string {
	lines := append([]string{"Аккредитационная подкомиссия по специальности"}, p.Specialties...)
	for i := 1; i < len(lines); i++ {
		lines[i] = `«` + lines[i] + `»`
	}
	lines = append(lines, "", p.FullName, "", strings.Join(p.Roles, ", "))
	var b strings.Builder
	for _, s := range lines {
		b.WriteString(fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="center"/><w:spacing w:before="0" w:after="0" w:line="320" w:lineRule="exact"/></w:pPr><w:r><w:rPr><w:rFonts w:ascii="Times New Roman" w:hAnsi="Times New Roman"/><w:sz w:val="32"/></w:rPr><w:t>%s</w:t></w:r></w:p>`, html.EscapeString(s)))
	}
	return b.String()
}

func expand(people []model.Person) []model.Person {
	cards := make([]model.Person, 0, len(people))
	for _, person := range people {
		if len(person.Assignments) > 0 {
			for _, assignment := range person.Assignments {
				cards = append(cards, model.Person{FullName: person.FullName, Specialties: []string{assignment.Specialty}, Roles: []string{assignment.Role}})
			}
			continue
		}
		specialties := person.Specialties
		roles := person.Roles
		if len(specialties) == 0 {
			specialties = []string{"Специальность не определена"}
		}
		if len(roles) == 0 {
			roles = []string{"Роль не указана"}
		}
		for _, specialty := range specialties {
			for _, role := range roles {
				cards = append(cards, model.Person{FullName: person.FullName, Specialties: []string{specialty}, Roles: []string{role}})
			}
		}
	}
	return cards
}
