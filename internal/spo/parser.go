package spo

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"badges/internal/model"
	"badges/internal/normalize"
	"badges/internal/ooxml"
)

var role = regexp.MustCompile(`(?i)(заместитель\s+председателя|председатель|секретарь|член)\s*АПК`)
var specialty = regexp.MustCompile(`(?i)подкомисси.{0,50}специальнос\s*ти\s*[«"]([^»"]+)`)
var roleWithSpecialty = regexp.MustCompile(`(?i)^(?:заместитель\s+председателя|председатель|секретарь|член)\s*АПК(.*)$`)
var educationMark = regexp.MustCompile(`(?i)ПГМУ|ПБМК|ПМБК`)

func Parse(doc *ooxml.Document, source string) ([]model.Record, []model.Diagnostic) {
	if doc == nil || doc.Root == nil {
		return nil, []model.Diagnostic{{Severity: model.Error, Stage: "parse", Source: source, Code: "EMPTY_DOCX", Message: "Документ пуст"}}
	}
	children := doc.Root.DescendantsNamed("body")
	if len(children) == 0 {
		return nil, []model.Diagnostic{{Severity: model.Error, Stage: "parse", Source: source, Code: "NO_BODY", Message: "Не найдено содержимое документа"}}
	}
	body := children[0]
	records := []model.Record{}
	diags := []model.Diagnostic{}
	current := ""
	tableNo := 0
	for _, node := range body.Children {
		if node.Name.Local == "p" {
			if m := specialty.FindStringSubmatch(normalize.CollapseSpace(node.VisibleText())); len(m) > 1 {
				current = normalize.CollapseSpace(m[1])
			}
			continue
		}
		if node.Name.Local != "tbl" {
			continue
		}
		tableNo++
		rows := node.ChildrenNamed("tr")
		if len(rows) == 0 {
			continue
		}
		header := normalize.CanonicalName(rows[0].VisibleText())
		if !strings.Contains(header, "ФИО ЭКСПЕРТА") || !strings.Contains(header, "СТАТУС В СОСТАВЕ ПОДКОМИССИИ") {
			continue
		}
		for rowNo, row := range rows[1:] {
			cells := row.ChildrenNamed("tc")
			texts := make([]string, len(cells))
			status := -1
			for i, c := range cells {
				texts[i] = normalize.CollapseSpace(c.VisibleText())
				if role.MatchString(texts[i]) {
					status = i
				}
			}
			if status < 0 {
				if rowName(texts) != "" {
					diags = append(diags, model.Diagnostic{Severity: model.Warning, Stage: "parse", Source: source, Code: "ROLE_UNRECOGNIZED", Message: "Строка с ФИО пропущена: не распознана роль (таблица " + strconv.Itoa(tableNo) + ", строка " + strconv.Itoa(rowNo+2) + ")"})
				}
				continue
			}
			if status < 1 {
				continue
			}
			name := ""
			for i := status - 1; i >= 0; i-- {
				if texts[i] != "" && !strings.EqualFold(texts[i], "Ответственное лицо") {
					name = texts[i]
					break
				}
			}
			if !validName(name) {
				diags = append(diags, model.Diagnostic{Severity: model.Warning, Stage: "parse", Source: source, Code: "FIO_MISSING", Message: "Не удалось определить ФИО"})
				continue
			}
			s := specialtyInRole(texts[status])
			if s == "" {
				s = current
			}
			if s == "" {
				diags = append(diags, model.Diagnostic{Severity: model.Warning, Stage: "parse", Source: source, Code: "SPECIALTY_MISSING", Message: "Не определена специальность"})
			}
			records = append(records, model.Record{FullName: name, RawRole: texts[status], Specialty: s, Source: model.SourceRef{File: source, Table: tableNo, Row: rowNo + 2}})
		}
	}
	return records, diags
}

func rowName(texts []string) string {
	for _, text := range texts {
		if validName(text) {
			return text
		}
	}
	return ""
}

// specialtyInRole extracts the specialty printed beside the role. Source
// files often keep several specialties in one schedule table, so this value
// is more specific than the table heading. It also accepts broken quote pairs
// found in supplied Word documents.
func specialtyInRole(value string) string {
	match := roleWithSpecialty.FindStringSubmatch(normalize.CollapseSpace(value))
	if len(match) < 2 {
		return ""
	}
	tail := strings.TrimSpace(match[1])
	if tail == "" {
		return ""
	}

	if open := strings.Index(tail, "«"); open >= 0 {
		tail = tail[open+len("«"):]
		if close := strings.Index(tail, "»"); close >= 0 {
			tail = tail[:close]
		}
	} else if open := strings.Index(tail, `"`); open >= 0 {
		tail = tail[open+1:]
		if close := strings.Index(tail, `"`); close >= 0 {
			tail = tail[:close]
		}
	} else if close := strings.Index(tail, "»"); close >= 0 {
		tail = tail[:close]
	}
	tail = educationMark.ReplaceAllString(tail, "")
	return normalize.CollapseSpace(strings.Trim(tail, " «»\"'"))
}
func validName(s string) bool {
	p := strings.Fields(s)
	if len(p) < 2 || len(p) > 4 {
		return false
	}
	for _, w := range p {
		r := []rune(w)
		if len(r) == 0 || !unicode.IsUpper(r[0]) {
			return false
		}
	}
	return true
}
