package aggregate_test

import (
	"testing"

	"badges/internal/aggregate"
	"badges/internal/model"
)

func TestRecordsAggregateMappedRolesAndSpecialties(t *testing.T) {
	input := []model.Record{
		{FullName: "Иванова Анна Петровна", RawRole: "Член АПК «Гистология» ПГМУ", Specialty: "Гистология"},
		{FullName: "иванова анна петровна", RawRole: "Председатель АПК «Рентгенология", Specialty: "Рентгенология"},
		{FullName: "Иванова Анна Петровна", RawRole: "Член АПК «Гистология» ПГМУ", Specialty: "Гистология"},
	}

	people, diagnostics := aggregate.Records(input)
	if len(diagnostics) != 0 || len(people) != 1 {
		t.Fatalf("got %d people and %d diagnostics", len(people), len(diagnostics))
	}
	person := people[0]
	if len(person.Roles) != 2 || person.Roles[0] != "Член подкомиссии" || person.Roles[1] != "Председатель подкомиссии" {
		t.Fatalf("unexpected roles: %#v", person.Roles)
	}
	if len(person.Specialties) != 2 || person.Specialties[0] != "Гистология" || person.Specialties[1] != "Рентгенология" {
		t.Fatalf("unexpected specialties: %#v", person.Specialties)
	}
}

func TestRecordsKeepsUnknownRoleAsWarning(t *testing.T) {
	people, diagnostics := aggregate.Records([]model.Record{{FullName: "Петрова Ольга Сергеевна", RawRole: "Куратор АПК", Specialty: "Гистология"}})
	if len(people) != 1 || len(people[0].Roles) != 1 || people[0].Roles[0] != "Куратор АПК" {
		t.Fatalf("unknown role should be preserved: %#v", people)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != "UNKNOWN_ROLE" {
		t.Fatalf("expected UNKNOWN_ROLE warning, got %#v", diagnostics)
	}
}

func TestRecordsMapsLowercaseSecretaryAndKeepsSimilarNamesSeparate(t *testing.T) {
	people, diagnostics := aggregate.Records([]model.Record{
		{FullName: "Сергеева Анна Петровна", RawRole: "секретарь АПК «Гистология»", Specialty: "Гистология"},
		{FullName: "Сергеева Анна Павловна", RawRole: "", Specialty: "Гистология"},
	})
	if len(people) != 2 {
		t.Fatalf("similar names must stay separate, got %d people", len(people))
	}
	if people[0].Roles[0] != "Секретарь подкомиссии" || people[1].Roles[0] != "Роль не указана" {
		t.Fatalf("unexpected role mapping: %#v", people)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != "MISSING_ROLE" {
		t.Fatalf("expected MISSING_ROLE, got %#v", diagnostics)
	}
}
