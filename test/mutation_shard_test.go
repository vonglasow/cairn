package test

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func mutationScript(t *testing.T, args ...string) string {
	t.Helper()
	run := exec.Command("bash", append([]string{filepath.Join("scripts", "mutation.sh")}, args...)...)
	run.Dir = ".."
	out, err := run.Output()
	if err != nil {
		t.Fatalf("mutation.sh %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func TestMutationList_givesTheEngineAJobPerShard(t *testing.T) {
	var listed []string
	if err := json.Unmarshal([]byte(mutationScript(t, "--list")), &listed); err != nil {
		t.Fatalf("--list is not a JSON list: %v", err)
	}
	if slices.Contains(listed, "./internal/engine") {
		t.Errorf("the engine is listed whole, a job longer than its runner allows: %v", listed)
	}
	for _, shard := range []string{"./internal/engine@1/4", "./internal/engine@4/4"} {
		if !slices.Contains(listed, shard) {
			t.Errorf("%s is not listed: %v", shard, listed)
		}
	}
	if !slices.Contains(listed, "./internal/domain") {
		t.Errorf("a package not split is no longer listed whole: %v", listed)
	}
}

func TestMutationShards_judgeEveryFileOfTheEngineOnce(t *testing.T) {
	sources, err := filepath.Glob(filepath.Join("..", "internal", "engine", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, s := range sources {
		if !strings.HasSuffix(s, "_test.go") {
			want = append(want, filepath.Base(s))
		}
	}
	seen := map[string]int{}
	for _, shard := range []string{"1/4", "2/4", "3/4", "4/4"} {
		files := strings.Fields(mutationScript(t, "--shard", "./internal/engine@"+shard))
		if len(files) == 0 {
			t.Errorf("shard %s judges no file", shard)
		}
		for _, f := range files {
			seen[f]++
		}
	}
	for _, f := range want {
		if seen[f] != 1 {
			t.Errorf("%s is judged by %d shards, want 1", f, seen[f])
		}
	}
	if len(seen) != len(want) {
		t.Errorf("the shards judge %d files, the engine has %d", len(seen), len(want))
	}
}
