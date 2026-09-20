package aggregate

import (
	"regexp"
	"strings"

	"badges/internal/model"
	"badges/internal/normalize"
)

var rolePattern = regexp.MustCompile(`(?i)(заместитель\s+председателя|председатель|секретарь|член)\s*АПК`)

// MapRole maps source statuses to the wording used by the approved badge
// reference. The second value reports whether a known mapping was used.
func MapRole(raw string) (string, bool) {
	match := rolePattern.FindString(raw)
	if match == "" {
		return normalize.VisibleText(raw), false
	}

	key := strings.ToUpper(normalize.CollapseSpace(match))
	switch key {
	case "ПРЕДСЕДАТЕЛЬ АПК":
		return "Председатель подкомиссии", true
	case "ЗАМЕСТИТЕЛЬ ПРЕДСЕДАТЕЛЯ АПК":
		return "Заместитель председателя подкомиссии", true
	case "СЕКРЕТАРЬ АПК":
		return "Секретарь подкомиссии", true
	case "ЧЛЕН АПК":
		return "Член подкомиссии", true
	default:
		return normalize.VisibleText(raw), false
	}
}

// Records groups source records by canonical FIO while retaining first-seen
// display spelling and ordering for the printed result.
func Records(records []model.Record) ([]model.Person, []model.Diagnostic) {
	people := make([]model.Person, 0)
	indices := make(map[string]int)
	diagnostics := make([]model.Diagnostic, 0)

	for _, record := range records {
		name := normalize.VisibleText(record.FullName)
		key := normalize.CanonicalName(name)
		if key == "" {
			diagnostics = append(diagnostics, diagnostic(record, "MISSING_NAME", "ФИО не указано"))
			continue
		}

		index, found := indices[key]
		if !found {
			index = len(people)
			indices[key] = index
			people = append(people, model.Person{FullName: name})
		}

		person := &people[index]
		role, knownRole := MapRole(record.RawRole)
		if role == "" {
			role = "Роль не указана"
			diagnostics = append(diagnostics, diagnostic(record, "MISSING_ROLE", "Роль не указана"))
		} else if !knownRole {
			diagnostics = append(diagnostics, diagnostic(record, "UNKNOWN_ROLE", "Неизвестная роль сохранена без изменения"))
		}
		addUnique(&person.Roles, role)

		specialty := normalize.VisibleText(record.Specialty)
		if specialty == "" {
			specialty = "Специальность не определена"
			diagnostics = append(diagnostics, diagnostic(record, "MISSING_SPECIALTY", "Специальность не определена"))
		}
		addUnique(&person.Specialties, specialty)
		addAssignment(&person.Assignments, role, specialty, record.Source)
		person.Sources = append(person.Sources, record.Source)
	}

	return people, diagnostics
}

func addAssignment(assignments *[]model.Assignment, role, specialty string, source model.SourceRef) {
	for i := range *assignments {
		if (*assignments)[i].Role == role && (*assignments)[i].Specialty == specialty {
			(*assignments)[i].Sources = append((*assignments)[i].Sources, source)
			return
		}
	}
	*assignments = append(*assignments, model.Assignment{Role: role, Specialty: specialty, Sources: []model.SourceRef{source}})
}

func addUnique(values *[]string, value string) {
	for _, existing := range *values {
		if existing == value {
			return
		}
	}
	*values = append(*values, value)
}

func diagnostic(record model.Record, code, message string) model.Diagnostic {
	return model.Diagnostic{
		Severity: model.Warning,
		Stage:    "aggregate",
		Source:   record.Source.File,
		Code:     code,
		Message:  message,
	}
}
