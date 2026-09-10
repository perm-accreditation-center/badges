# Badge Generator Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a dependency-free Go application that converts SPO commission DOCX files into exact 90×55 mm accreditation badges in a printable DOCX.

**Architecture:** The application reads OOXML directly from each input DOCX, converts only recognized SPO rows into a neutral `Record`, normalizes and aggregates records by person, then plans physical cards before writing a new fixed-layout OOXML document. The input profile, aggregation, layout, DOCX writer, reports, and terminal orchestration are isolated so a future VO profile can share the latter five components without changing their contracts.

**Tech Stack:** Go 1.22+; standard library only (`archive/zip`, `encoding/xml`, `encoding/csv`, `testing`); OOXML/DOCX; GitHub Actions optional only after local functionality works.

---

## Planned file structure

```text
go.mod                                      module `badges`
cmd/badge-generator/main.go                 process startup and exit code
internal/model/model.go                     immutable domain structs and diagnostics
internal/normalize/text.go                  XML text and name canonicalization
internal/aggregate/aggregate.go             role mapping and aggregation by person
internal/input/discover.go                  safe input discovery and size limits
internal/ooxml/read.go                      ZIP and XML reading, visible text extraction
internal/spo/parser.go                      SPO table/header/row recognition
internal/layout/layout.go                   card segmentation and 2×5 pagination
internal/docx/write.go                      complete output DOCX package writer
internal/report/report.go                   CSV and run-log creation
internal/app/run.go                         six-stage orchestration and terminal output
internal/testdocx/fixture.go                synthetic DOCX fixture builder for tests
tests/manual/PRINTING.md                    operator print verification procedure
README.md                                   installation, folders, normal run, errors
```

All automated test files live beside the implementation file they exercise (`*_test.go`). Synthetic fixtures contain invented names only; the supplied real commission document is never added to Git.

## Task 1: Establish the Go module and domain contracts

**Files:**
- Create: `go.mod`
- Create: `internal/model/model.go`
- Create: `internal/model/model_test.go`
- Create: `cmd/badge-generator/main.go`

- [x] **Bootstrap: Create the Go module configuration.**

Create only `go.mod` before the first test:

```text
module badges

go 1.22
```

This is build configuration, not production behavior; it lets the deliberately failing test resolve the intended package path.

- [x] **Step 1: Write the failing model-contract test.**

```go
func TestDiagnosticErrorIncludesStageAndSource(t *testing.T) {
    d := model.Diagnostic{Severity: model.Error, Stage: "parse", Source: "input/a.docx", Message: "ФИО не найдено"}
    got := d.String()
    if !strings.Contains(got, "parse") || !strings.Contains(got, "input/a.docx") {
        t.Fatalf("diagnostic must be actionable, got %q", got)
    }
}
```

- [x] **Step 2: Run the test to establish the red state.**

Run: `go test ./internal/model -run TestDiagnosticErrorIncludesStageAndSource -v`  
Expected: FAIL because package `badges/internal/model` is absent.

- [x] **Step 3: Add the smallest stable domain model.**

Create `go.mod` with `module badges` and `go 1.22`. Create `internal/model/model.go` with the following exported API; do not put parsing logic in this package.

```go
package model

type Severity string
const ( Info Severity = "INFO"; Warning Severity = "WARNING"; Error Severity = "ERROR" )

type SourceRef struct { File string; Table, Row int }
type Record struct { FullName, RawRole, Specialty string; Source SourceRef }
type Person struct { FullName string; Roles, Specialties []string; Sources []SourceRef }
type Diagnostic struct { Severity Severity; Stage, Source, Code, Message string }
func (d Diagnostic) String() string { return string(d.Severity)+" ["+d.Stage+"] "+d.Source+": "+d.Message }
```

- [x] **Step 4: Re-run the package test.**

Run: `go test ./internal/model -v`  
Expected: PASS with one test.

- [x] **Step 5: Create a compiling command entry point and commit.**

Create `cmd/badge-generator/main.go`:

```go
package main

import "fmt"

func main() { fmt.Println("Генератор бейджей аккредитации") }
```

Run: `go test ./...`  
Expected: PASS.  
Commit: `git add go.mod cmd/badge-generator internal/model && git commit -m "chore: scaffold badge generator domain"`.

## Task 2: Normalize text, map roles, and aggregate by FIO

**Files:**
- Create: `internal/normalize/text.go`
- Create: `internal/normalize/text_test.go`
- Create: `internal/aggregate/aggregate.go`
- Create: `internal/aggregate/aggregate_test.go`

- [x] **Step 1: Write failing normalization and aggregation tests.**

```go
func TestCanonicalNameTreatsYoAndSpacingAsDuplicate(t *testing.T) {
    if normalize.CanonicalName("  Алёна\u00a0Иванова ") != normalize.CanonicalName("АЛЕНА ИВАНОВА") {
        t.Fatal("Ё/Е, case, and whitespace must normalize")
    }
}

func TestAggregateMakesOnePersonWithMappedRolesAndSpecialties(t *testing.T) {
    in := []model.Record{
        {FullName: "Иванова Анна Петровна", RawRole: "Член АПК «Гистология» ПГМУ", Specialty: "Гистология"},
        {FullName: "иванова анна петровна", RawRole: "Председатель АПК «Рентгенология" , Specialty: "Рентгенология"},
    }
    people, diags := aggregate.Records(in)
    if len(people) != 1 || len(people[0].Roles) != 2 || len(people[0].Specialties) != 2 || len(diags) != 0 { t.Fatal("records were not aggregated") }
    if people[0].Roles[0] != "Член подкомиссии" || people[0].Roles[1] != "Председатель подкомиссии" { t.Fatal("roles not mapped") }
}
```

- [x] **Step 2: Run both tests to establish the red state.**

Run: `go test ./internal/normalize ./internal/aggregate -v`  
Expected: FAIL because packages and functions are absent.

- [x] **Step 3: Implement deterministic text and role functions.**

Implement `normalize.VisibleText`, `normalize.CollapseSpace`, and `normalize.CanonicalName`. `CanonicalName` must trim, collapse Unicode whitespace, uppercase with `strings.ToUpper`, and replace `Ё` by `Е` only in the key. Implement `aggregate.MapRole(raw string) (string, bool)` with these exact mappings:

```go
"ПРЕДСЕДАТЕЛЬ АПК"              -> "Председатель подкомиссии"
"ЗАМЕСТИТЕЛЬ ПРЕДСЕДАТЕЛЯ АПК"  -> "Заместитель председателя подкомиссии"
"СЕКРЕТАРЬ АПК"                 -> "Секретарь подкомиссии"
"ЧЛЕН АПК"                      -> "Член подкомиссии"
```

Strip the first quoted specialty fragment and trailing organization tokens only when determining the role. Preserve the original first-seen spelling of FIO and preserve first-seen order of unique roles and specialties. Return a warning diagnostic with code `UNKNOWN_ROLE` for an unmapped role; never invent a mapping.

- [x] **Step 4: Add edge tests and run them.**

Add cases for `секретарь АПК`, duplicate role from a second file, blank role, two people with similar FIO, and an unknown status. Run: `go test ./internal/normalize ./internal/aggregate -v`  
Expected: PASS, including stable first-seen order.

- [x] **Step 5: Run the full suite and commit.**

Run: `go test ./...`  
Expected: PASS.  
Commit: `git add internal/normalize internal/aggregate && git commit -m "feat: normalize and aggregate commission members"`.

## Task 3: Safely read DOCX OOXML and discover input files

**Files:**
- Create: `internal/input/discover.go`
- Create: `internal/input/discover_test.go`
- Create: `internal/ooxml/read.go`
- Create: `internal/ooxml/read_test.go`
- Create: `internal/testdocx/fixture.go`

- [ ] **Step 1: Write failing discovery and OOXML tests.**

```go
func TestDiscoverSkipsWordLockAndNonDocx(t *testing.T) {
    dir := t.TempDir()
    for _, n := range []string{"schedule.docx", "~$schedule.docx", "note.pdf"} { os.WriteFile(filepath.Join(dir, n), []byte("x"), 0600) }
    got, diags := input.Discover(dir)
    if len(got) != 1 || filepath.Base(got[0]) != "schedule.docx" || len(diags) != 2 { t.Fatal("unexpected discovery result") }
}

func TestReadDocumentIgnoresDeletedText(t *testing.T) {
    path := testdocx.Write(t, `<w:document><w:body><w:p><w:del><w:delText>Удалённая Фамилия</w:delText></w:del><w:r><w:t>Актуальная Фамилия</w:t></w:r></w:p></w:body></w:document>`)
    doc, err := ooxml.ReadDocument(path)
    if err != nil || strings.Contains(doc.Text(), "Удалённая") || !strings.Contains(doc.Text(), "Актуальная") { t.Fatal("tracked deletion handling failed") }
}
```

- [ ] **Step 2: Run the new tests to confirm they fail.**

Run: `go test ./internal/input ./internal/ooxml -v`  
Expected: FAIL because the packages do not exist.

- [ ] **Step 3: Implement bounded discovery and reading.**

Define `func input.Discover(dir string) ([]string, []model.Diagnostic)` and `func ooxml.ReadDocument(path string) (*ooxml.Document, error)`. `Document` exposes `Text() string`, `BodyChildren() []*Node`, and each `Node` exposes its Word local name, children, attributes, and visible text. In `internal/testdocx`, define `func Write(t *testing.T, documentXML string) string` and `func ReadZipFile(t *testing.T, path, member string) string`. `input.Discover` scans only direct children, ignores folders, `~$*.docx`, and non-DOCX files, returning informational diagnostics for ignored items. `ooxml.ReadDocument` opens the ZIP without extracting it, requires `word/document.xml`, rejects encrypted/unreadable archives, limits file size to 50 MiB and uncompressed XML to 25 MiB, and uses an `encoding/xml.Decoder` that rejects `xml.Directive` and any external entity. Build a small node tree retaining element name, attributes, children, and text; omit text nested inside `w:del`, retain ordinary text and `w:ins`.

- [ ] **Step 4: Add the hardening cases and run them.**

Add tests for missing `word/document.xml`, a ZIP containing a 25 MiB+ XML part, a 50 MiB+ input file, malformed XML, and an XML directive. Run: `go test ./internal/input ./internal/ooxml -v`  
Expected: PASS; each failure returns an actionable error, not a panic.

- [ ] **Step 5: Run all tests and commit.**

Run: `go test ./...`  
Expected: PASS.  
Commit: `git add internal/input internal/ooxml internal/testdocx && git commit -m "feat: safely discover and read DOCX input"`.

## Task 4: Parse the SPO profile, including shifted responsible-person rows

**Files:**
- Create: `internal/spo/parser.go`
- Create: `internal/spo/parser_test.go`
- Modify: `internal/testdocx/fixture.go`

- [ ] **Step 1: Write failing table-driven parser tests.**

Create a synthetic document with two headings and tables. The first table uses normal cells `[number, time, fio, status, workplace]`; the second contains `["Ответственное лицо", fio, status, workplace]`. Assert that both produce records, the specialities come from their nearest preceding headings, and `SourceRef.Table`/`Row` are populated.

```go
func TestParseIncludesShiftedResponsiblePerson(t *testing.T) {
    doc := testdocx.SPO(t, testdocx.Table{Specialty: "Гистология", Rows: [][]string{
        {"1.", "", "Иванова Анна Петровна", "Член АПК «Гистология» ПГМУ", ""},
        {"Ответственное лицо", "Петрова Ольга Сергеевна", "Председатель АПК «Гистология»", ""},
    }})
    records, diags := spo.Parse(doc, "input/source.docx")
    if len(records) != 2 || records[1].FullName != "Петрова Ольга Сергеевна" || records[1].Specialty != "Гистология" || len(diags) != 0 { t.Fatal("shifted row was not parsed") }
}
```

- [ ] **Step 2: Run the parser test to confirm the red state.**

Run: `go test ./internal/spo -run TestParseIncludesShiftedResponsiblePerson -v`  
Expected: FAIL because `spo.Parse` is absent.

- [ ] **Step 3: Implement only the explicit SPO recognition rules.**

Implement `func Parse(doc *ooxml.Document, source string) ([]model.Record, []model.Diagnostic)`. In `internal/testdocx`, define `type Table struct { Specialty string; Rows [][]string }`, `func SPO(t *testing.T, tables ...Table) *ooxml.Document`, and `func WriteSPOFile(t *testing.T, path string)`. Select outer tables whose header text includes normalized `ФИО эксперта` and `Статус в составе подкомиссии`. For each table, walk preceding body paragraphs backwards and match `подкомисси.*специальнос\s*ти\s*[«"]([^»"]+)`; use the matched value as primary specialty. For each data row, find a cell containing a known `… АПК` status, select the nearest non-empty cell on its left excluding `Ответственное лицо`, and validate it as two to four title-cased Cyrillic name words. Only if that path fails, scan other cells for a validated FIO and emit diagnostic code `FIO_FALLBACK`.

Use a quoted specialty from the status only if no table heading is found; report `SPECIALTY_MISSING` if both fail. Rows with no role or no valid FIO produce warnings and no record. Do not add a general “guess the table” mode.

- [ ] **Step 4: Add every parser regression fixture and run it.**

Add tests for reversed header columns, split `специальнос`/`ти` runs, lower-case `секретарь`, broken closing quote in the status, a workplace containing a FIO-like phrase, no matching table, and tracked deletion inside a name. Run: `go test ./internal/spo -v`  
Expected: PASS with every rejected row represented by a diagnostic.

- [ ] **Step 5: Add a non-committed manual regression command and commit.**

Document in `README.md` that maintainers can run `go run ./cmd/badge-generator --no-pause` against a copied real source in `input`; never commit that source. Run: `go test ./...`  
Expected: PASS.  
Commit: `git add internal/spo internal/testdocx README.md && git commit -m "feat: parse SPO commission tables"`.

## Task 5: Plan exact cards, pages, and continuation cards

**Files:**
- Create: `internal/layout/layout.go`
- Create: `internal/layout/layout_test.go`

- [ ] **Step 1: Write failing continuation and pagination tests.**

```go
func TestPlanStartsNewContinuationWithoutSplittingItem(t *testing.T) {
    p := model.Person{FullName: "Иванова Анна Петровна", Specialties: []string{"Гистология", "Рентгенология", "Сестринское дело"}, Roles: []string{"Член подкомиссии", "Председатель подкомиссии"}}
    cards, diags := layout.Plan([]model.Person{p}, layout.Measure{MaxLines: 5})
    if len(diags) != 0 || len(cards) < 2 || cards[0].Part != 1 || cards[0].Parts != len(cards) || cards[1].FullName != p.FullName { t.Fatal("continuation metadata invalid") }
    for _, c := range cards { if strings.Contains(strings.Join(c.Lines, "\n"), "Рентгеноло\nгия") { t.Fatal("item was split") } }
}

func TestPaginateUsesTenCardsPerPage(t *testing.T) {
    cards := make([]layout.Card, 11)
    pages := layout.Paginate(cards)
    if len(pages) != 2 || len(pages[0].Cards) != 10 || len(pages[1].Cards) != 1 { t.Fatal("2x5 pagination broken") }
}
```

- [ ] **Step 2: Run tests to confirm they fail.**

Run: `go test ./internal/layout -v`  
Expected: FAIL because the package is absent.

- [ ] **Step 3: Implement deterministic segmentation.**

Define the complete API `type Card struct { FullName string; Specialties, Roles, Lines []string; Part, Parts int }`, `type Page struct { Cards []Card }`, `type Measure struct { MaxLines int }`, `func Plan([]model.Person, Measure) ([]Card, []model.Diagnostic)`, and `func Paginate([]Card) []Page`. `Lines` is the exact ordered text sequence rendered by the writer. `Measure` is a deterministic, conservative line estimator: it wraps only at spaces, counts heading/FIO/continuation lines, and never breaks a word. Start each card with the fixed heading and FIO, then append whole specialty and role strings in their established order. If a card exceeds available lines at 8 pt, close it and start the next with the same heading and FIO. If a single item alone cannot fit, return diagnostic code `ITEM_TOO_LONG` and omit only that person's physical cards; do not split the word or silently abbreviate it. `Paginate` must fill left-to-right then top-to-bottom in groups of 10.

- [ ] **Step 4: Add boundary cases and run them.**

Add 1, 9, 10, 11, and 101 cards; one person with three continuation cards; an empty list; and one item that cannot fit. Run: `go test ./internal/layout -v`  
Expected: PASS.

- [ ] **Step 5: Run all tests and commit.**

Run: `go test ./...`  
Expected: PASS.  
Commit: `git add internal/layout && git commit -m "feat: plan badge cards and continuations"`.

## Task 6: Write a fixed-geometry output DOCX and structurally verify it

**Files:**
- Create: `internal/docx/write.go`
- Create: `internal/docx/write_test.go`

- [ ] **Step 1: Write failing OOXML-geometry tests.**

```go
func TestWriteSetsExactA4BadgeGrid(t *testing.T) {
    path := filepath.Join(t.TempDir(), "badges.docx")
    if err := docx.Write(path, []layout.Page{{Cards: make([]layout.Card, 10)}}); err != nil { t.Fatal(err) }
    xml := testdocx.ReadZipFile(t, path, "word/document.xml")
    for _, want := range []string{`w:w="11906" w:h="16838"`, `w:left="850"`, `w:top="624"`, `w:gridCol w:w="5102"`, `w:trHeight w:val="3118"`, `w:hRule="exact"`, `w:tblLayout w:type="fixed"`} {
        if !strings.Contains(xml, want) { t.Fatalf("missing %s", want) }
    }
}
```

- [ ] **Step 2: Run the test to establish the red state.**

Run: `go test ./internal/docx -run TestWriteSetsExactA4BadgeGrid -v`  
Expected: FAIL because the writer is absent.

- [ ] **Step 3: Implement the minimal complete DOCX package.**

`docx.Write(path string, pages []layout.Page) error` must create a ZIP containing `[Content_Types].xml`, `_rels/.rels`, `word/document.xml`, `word/_rels/document.xml.rels`, `word/styles.xml`, and `word/settings.xml`. In `document.xml` generate one 2-column table per A4 page, five exact-height rows per table, no cell spacing, no cell padding, no header/footer references, and thin solid borders. Use page size 11906×16838 DXA, margins left/right 850 DXA (15 mm), top/bottom 624 DXA (11 mm), cell widths 5102 DXA, and row height 3118 DXA. Escape all text with `encoding/xml` before placing it in `<w:t>`.

Use Times New Roman, centered paragraphs, 16 pt first, decreasing only to 8 pt according to the layout fit result. Render the title exactly as `Аккредитационная подкомиссия по специальности`, each specialty as `«…»`, then FIO and mapped roles. Add `Продолжение i из N` only for cards with `Parts > 1`.

- [ ] **Step 4: Add content and package tests.**

Test 1, 10, and 11 cards; exact page count; special-character XML escaping; no `header` or `footer` relation; repeated FIO and continuation label; and valid ZIP reopening. Run: `go test ./internal/docx -v`  
Expected: PASS.

- [ ] **Step 5: Run full tests and commit.**

Run: `go test ./...`  
Expected: PASS.  
Commit: `git add internal/docx && git commit -m "feat: generate exact-size badge DOCX"`.

## Task 7: Add reports, six-stage run flow, and employee-facing terminal UX

**Files:**
- Create: `internal/report/report.go`
- Create: `internal/report/report_test.go`
- Create: `internal/app/run.go`
- Create: `internal/app/run_test.go`
- Modify: `cmd/badge-generator/main.go`

- [ ] **Step 1: Write failing end-to-end test in a temporary directory.**

```go
func TestRunCreatesTimestampedBundleAndReportsPartialFailure(t *testing.T) {
    root := t.TempDir(); mustMkdir(t, filepath.Join(root, "input")); mustMkdir(t, filepath.Join(root, "output"))
    testdocx.WriteSPOFile(t, filepath.Join(root, "input", "valid.docx"))
    os.WriteFile(filepath.Join(root, "input", "broken.docx"), []byte("not a zip"), 0600)
    result := app.Run(app.Options{Root: root, Pause: false, Writer: io.Discard})
    if result.ExitCode != app.ExitPartial || !fileExists(filepath.Join(result.Bundle, "бейджи.docx")) || !fileExists(filepath.Join(result.Bundle, "ошибки.csv")) { t.Fatal("partial run contract broken") }
}
```

- [ ] **Step 2: Run the end-to-end test to confirm it fails.**

Run: `go test ./internal/app -run TestRunCreatesTimestampedBundleAndReportsPartialFailure -v`  
Expected: FAIL because the application runner is absent.

- [ ] **Step 3: Implement reports and orchestration.**

Create `report.WriteCSV(bundle string, people []model.Person, diags []model.Diagnostic) error` that writes UTF-8 BOM CSV with headers and source references. Define `type Options struct { Root string; Pause bool; Writer io.Writer }`, `type Result struct { ExitCode int; Bundle string; Diagnostics []model.Diagnostic }`, constants `ExitSuccess = 0`, `ExitFatal = 1`, `ExitPartial = 2`, and `func Run(Options) Result`. `app.Run` must execute exactly these named stages: folders, discovery, extraction, aggregation, DOCX creation, reports. It creates `output/YYYY-MM-DD_HHMMSS[_N]/`, never deletes inputs or previous bundles, writes `бейджи.docx`, `отчёт.csv`, `ошибки.csv` only if diagnostics exist, and `run.log` with no full list of FIO. Its exit codes are `0` success, `1` no valid records/fatal output error, and `2` partial success.

Terminal lines must use this stable form: `[n/6] <stage>  <result>`. The final summary reports people, physical cards, pages, warnings, continuation count, bundle path, and the exact 100 % print reminder. `main.go` parses only `--no-pause`; it pauses only when requested by its launch mode, never during tests.

- [ ] **Step 4: Add failure and output-collision tests.**

Test empty `input`, no matching tables, an unwritable `output`, two bundles in the same second, ignored `~$` file, no diagnostics (no `ошибки.csv`), and the six required terminal stage labels. Run: `go test ./internal/report ./internal/app -v`  
Expected: PASS.

- [ ] **Step 5: Run full tests and commit.**

Run: `go test ./...`  
Expected: PASS.  
Commit: `git add internal/report internal/app cmd/badge-generator && git commit -m "feat: add terminal workflow and run reports"`.

## Task 8: Build, render, and perform release-quality verification

**Files:**
- Modify: `README.md`
- Create: `tests/manual/PRINTING.md`
- Create: `scripts/build.ps1`
- Create: `scripts/check_docx.go`

- [ ] **Step 1: Write the failing structural-check test command.**

Create `scripts/check_docx.go` as a standard-library program that opens a DOCX and exits non-zero unless each page table has two 5102-DXA grid columns, five 3118-DXA exact rows, A4 page geometry, and no header/footer references. Run it against a deliberately malformed synthetic DOCX.  
Expected: non-zero exit and a message naming the missing constraint.

- [ ] **Step 2: Implement the structural checker and build script.**

`scripts/build.ps1` must run `go test ./...`, build `bin/badge-generator.exe`, recreate only the explicit `bin` build artifact, and create `bin/input` and `bin/output`. It must not copy real source documents. `check_docx.go` must print `DOCX geometry check passed` only after all constraints succeed.

- [ ] **Step 3: Verify standard build artifacts.**

Run: `go test ./...`  
Expected: PASS.  
Run: `powershell -ExecutionPolicy Bypass -File scripts/build.ps1`  
Expected: `bin/badge-generator.exe` exists and all Go tests passed.  
Run: `go run ./scripts/check_docx.go bin/output/<generated-run>/бейджи.docx` after a synthetic end-to-end run.  
Expected: `DOCX geometry check passed`.

- [ ] **Step 4: Add operator documentation and execute visual QA.**

In `README.md`, document only: install location, `input`/`output`, normal run, error CSV, continuation cards, and “Фактический размер / 100 %”. In `tests/manual/PRINTING.md`, require rendering every page to PNG/PDF, opening the generated document in Word, printing a 10-card page at 100 %, and measuring top-left, middle, and bottom-right cards within ±0.5 mm. Record date, printer, driver, and three measurements in the release checklist; do not claim physical accuracy without this record.

- [ ] **Step 5: Verify final repository state and commit.**

Run: `go test ./...`  
Expected: PASS.  
Run: `git diff --check`  
Expected: no output and exit code 0.  
Commit: `git add README.md tests/manual scripts && git commit -m "docs: add build and print verification guidance"`.

## Plan self-review

| Specification requirement | Covered by |
|---|---|
| Input/output folders, no Office/network, safe source preservation | Tasks 3, 7, 8 |
| SPO headers, shifted `Ответственное лицо`, heading specialty, current Word text | Task 4 |
| FIO aggregation, role mapping to the reference, sources and warnings | Task 2 |
| Exact 90×55 mm, A4 2×5 grid, cut lines, no offsetting page furniture | Task 6 |
| Multiple physical cards without data loss for long people | Task 5 and Task 6 |
| Terminal progress, partial runs, CSV/log reports | Task 7 |
| Security limits and diagnostic errors | Task 3 and Task 7 |
| Automated, structural, visual, and physical-print acceptance | Tasks 1–8, especially Task 8 |
| Future VO extensibility | Task 1 contracts and Task 4 profile boundary |

The plan defines the same exported types and function names throughout, contains no production source changes, and keeps real personal data outside Git.
