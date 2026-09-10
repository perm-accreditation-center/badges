package model

import "fmt"

// Severity describes how a diagnostic affects a run.
type Severity string

const (
	Info    Severity = "INFO"
	Warning Severity = "WARNING"
	Error   Severity = "ERROR"
)

// SourceRef identifies the row that supplied a record.
type SourceRef struct {
	File  string
	Table int
	Row   int
}

// Record is a member entry read from one source document.
type Record struct {
	FullName  string
	RawRole   string
	Specialty string
	Source    SourceRef
}

// Person is the aggregated representation printed on one or more badges.
type Person struct {
	FullName    string
	Roles       []string
	Specialties []string
	Sources     []SourceRef
}

// Diagnostic is a human-readable warning or error associated with a stage.
type Diagnostic struct {
	Severity Severity
	Stage    string
	Source   string
	Code     string
	Message  string
}

func (d Diagnostic) String() string {
	location := d.Source
	if location == "" {
		location = "без источника"
	}
	return fmt.Sprintf("%s [%s] %s: %s", d.Severity, d.Stage, location, d.Message)
}
