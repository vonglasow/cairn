package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const gremlinsSaid = `{"files":[{"file_name":"wrap.go","mutations":[
 {"type":"ARITHMETIC_BASE","status":"KILLED","line":2,"column":9},
 {"type":"CONDITIONALS_BOUNDARY","status":"LIVED","line":3,"column":11},
 {"type":"ARITHMETIC_BASE","status":"NOT COVERED","line":4,"column":4},
 {"type":"INCREMENT_DECREMENT","status":"TIMED OUT","line":5,"column":5}]}]}`

func TestMutationReport_namesWhatToFixWithTheLineItSitsOn(t *testing.T) {
	root := t.TempDir()
	pkg := filepath.Join(root, "internal", "wrap")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	source := "package wrap\nvar a = 1 + 2\nif from < len(kept) {\nn := n * 2\nfor i := 0; i < n; i++ {\n"
	if err := os.WriteFile(filepath.Join(pkg, "wrap.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	run := exec.Command("bash", filepath.Join("..", "scripts", "mutation-report.sh"), "./internal/wrap", root)
	run.Stdin = strings.NewReader(gremlinsSaid)
	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("mutation-report.sh: %v\n%s", err, out)
	}
	said := string(out)
	for _, want := range []string{
		"## `./internal/wrap` — 2 to fix",
		"- LIVED CONDITIONALS_BOUNDARY at `internal/wrap/wrap.go:3:11`",
		"`if from < len(kept) {`",
		"- NOT COVERED ARITHMETIC_BASE at `internal/wrap/wrap.go:4:4`",
		"`n := n * 2`",
		"hang caught: INCREMENT_DECREMENT at `internal/wrap/wrap.go:5:5`",
	} {
		if !strings.Contains(said, want) {
			t.Errorf("the report says nothing of %q:\n%s", want, said)
		}
	}
	if strings.Contains(said, "KILLED") {
		t.Errorf("the report names a mutant that was killed:\n%s", said)
	}
}

func TestMutationReport_saysAPackageIsClean(t *testing.T) {
	run := exec.Command("bash", filepath.Join("..", "scripts", "mutation-report.sh"), "./internal/wrap", t.TempDir())
	run.Stdin = strings.NewReader(`{"files":[{"file_name":"wrap.go","mutations":[{"type":"ARITHMETIC_BASE","status":"KILLED","line":2,"column":9}]}]}`)
	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("mutation-report.sh: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); got != "- `./internal/wrap` — clean: 1 killed" {
		t.Errorf("a clean package is reported as %q", got)
	}
}
