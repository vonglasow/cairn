package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/build"
	"cairn/test"
)

type versionWorld struct {
	t                       *testing.T
	out                     test.Result
	archive, unpacked, repo string
	version                 string
	commit                  string
	built                   string
	committed               string
	restores                []func()
}

func newVersionWorld(t *testing.T) *versionWorld {
	w := &versionWorld{t: t}
	previous := stamps
	w.restores = append(w.restores, func() { stamps = previous })
	stamps = w.reading
	return w
}

func (w *versionWorld) reading() build.Info {
	return build.Stamped(w.version, w.commit, w.built, w.committed)
}

func (w *versionWorld) stamp(target *string, value string) { *target = value }

func (w *versionWorld) endScenario() {
	for i := len(w.restores) - 1; i >= 0; i-- {
		w.restores[i]()
	}
	w.restores = nil
}

func (w *versionWorld) stampedThere(at int, want string) error {
	said, err := exec.Command("make", "-s", "-C", w.unpacked, "stamps").Output()
	if err != nil {
		return fmt.Errorf("make stamps: %w", err)
	}
	if got := strings.Fields(string(said)); len(got) != 3 || got[at] != want {
		return fmt.Errorf("make stamps answered %q, want %q in place %d", said, want, at)
	}
	return nil
}

func (w *versionWorld) binaries(version string, systems ...string) error {
	w.version, w.archive = version, filepath.Join(w.t.TempDir(), "dist")
	if err := os.MkdirAll(w.archive, 0o700); err != nil {
		return err
	}
	for _, system := range systems {
		name := "cairn-" + version + "-" + system + ".tar.gz"
		if err := os.WriteFile(filepath.Join(w.archive, name), []byte(system), 0o600); err != nil {
			return err
		}
		if err := command(root(), "scripts/sums.sh", w.archive, name); err != nil {
			return err
		}
	}
	return nil
}

func root() string { return filepath.Join("..", "..") }

func command(dir, name string, args ...string) error {
	run := exec.Command(name, args...)
	run.Dir = dir
	if out, err := run.CombinedOutput(); err != nil {
		return fmt.Errorf("%s %v: %w\n%s", name, args, err, out)
	}
	return nil
}

func (w *versionWorld) line(key string) string {
	for _, l := range strings.Split(w.out.Stdout, "\n") {
		if rest, ok := strings.CutPrefix(l, key); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

func initVersion(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newVersionWorld(t)

		sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
			w.endScenario()
			return ctx, err
		})

		sc.Step(`^a binary built with no version stamps$`, func() error {
			w.stamp(&w.version, "")
			w.stamp(&w.commit, "")
			w.stamp(&w.built, "")
			w.stamp(&w.committed, "")
			return nil
		})

		sc.Step(`^a binary stamped with version "([^"]*)"$`, func(v string) error {
			w.stamp(&w.version, v)
			return nil
		})

		sc.Step(`^stamped with commit "([^"]*)"$`, func(v string) error {
			w.stamp(&w.commit, v)
			return nil
		})

		sc.Step(`^stamped with build time "([^"]*)"$`, func(v string) error {
			w.stamp(&w.built, v)
			return nil
		})

		sc.Step(`^stamped with commit date "([^"]*)"$`, func(v string) error {
			w.stamp(&w.committed, v)
			return nil
		})

		sc.Step(`^the source archive of this repository made as version "([^"]*)"$`, func(version string) error {
			w.archive = filepath.Join(w.t.TempDir(), "dist")
			return command(root(), "scripts/dist.sh", version, w.archive)
		})

		formula := func(name string) (string, error) {
			said, err := os.ReadFile(filepath.Join(w.archive, "Formula", name+".rb"))
			return string(said), err
		}
		sc.Step(`^the binary archives of version "([^"]*)" for macOS and Linux, on arm and on Intel$`, func(version string) error {
			return w.binaries(version, "darwin-arm64", "darwin-amd64", "linux-arm64", "linux-amd64")
		})
		sc.Step(`^the binary archives of version "([^"]*)" for macOS only$`, func(version string) error {
			return w.binaries(version, "darwin-arm64", "darwin-amd64")
		})
		sc.Step(`^its formula is made, fetching from "([^"]*)"$`, func(from string) error {
			made := exec.Command("scripts/formula.sh", w.version, w.archive, from)
			made.Dir = root()
			out, err := made.CombinedOutput()
			w.out = test.Result{Stdout: string(out)}
			if err != nil {
				w.out.Code = 1
			}
			return nil
		})
		sc.Step(`^the formula "([^"]*)" is for version "([^"]*)"$`, func(name, version string) error {
			said, err := formula(name)
			if err != nil {
				return fmt.Errorf("%w\n%s", err, w.out.Stdout)
			}
			class := "class Cairn < Formula"
			if minor, ok := strings.CutPrefix(name, "cairn@"); ok {
				class = "class CairnAT" + strings.ReplaceAll(minor, ".", "") + " < Formula"
			}
			for _, must := range []string{class, `version "` + version + `"`} {
				if !strings.Contains(said, must) {
					return fmt.Errorf("the formula %s says nothing of %q:\n%s", name, must, said)
				}
			}
			return nil
		})
		sc.Step(`^it fetches "([^"]*)" on (macOS|Linux) on (arm|Intel), with its checksum$`, func(archive, system, cpu string) error {
			said, err := formula("cairn")
			if err != nil {
				return err
			}
			sums, err := os.ReadFile(filepath.Join(w.archive, "SHA256SUMS"))
			if err != nil {
				return err
			}
			sum := ""
			for _, l := range strings.Split(string(sums), "\n") {
				if f := strings.Fields(l); len(f) == 2 && f[1] == archive {
					sum = f[0]
				}
			}
			on := map[string]string{"macOS": "on_macos do", "Linux": "on_linux do"}[system]
			_, block, found := strings.Cut(said, on)
			if !found {
				return fmt.Errorf("the formula has no %q:\n%s", on, said)
			}
			_, block, _ = strings.Cut(block, map[string]string{"arm": "on_arm do", "Intel": "on_intel do"}[cpu])
			block, _, _ = strings.Cut(block, "end")
			if !strings.Contains(block, "/v9.9.9/"+archive+`"`) || sum == "" || !strings.Contains(block, `sha256 "`+sum+`"`) {
				return fmt.Errorf("on %s on %s the formula says:%s", system, cpu, block)
			}
			return nil
		})
		sc.Step(`^it installs the binary, and builds nothing$`, func() error {
			said, err := formula("cairn")
			if err != nil {
				return err
			}
			if !strings.Contains(said, `bin.install "cairn"`) || strings.Contains(said, "depends_on") || strings.Contains(said, `system "make"`) {
				return fmt.Errorf("the formula builds, or installs nothing:\n%s", said)
			}
			return nil
		})
		sc.Step(`^no formula was made$`, func() error {
			if w.out.Code == 0 {
				return fmt.Errorf("the formula was made with a system's binary missing:\n%s", w.out.Stdout)
			}
			if _, err := os.Stat(filepath.Join(w.archive, "Formula")); err == nil {
				return fmt.Errorf("a formula was left behind")
			}
			return nil
		})

		packed := func(version string) string {
			ext := ".tar.gz"
			if runtime.GOOS == "windows" {
				ext = ".zip"
			}
			return "cairn-" + version + "-" + runtime.GOOS + "-" + runtime.GOARCH + ext
		}
		sc.Step(`^a binary archive of this repository made as version "([^"]*)" for this system$`, func(version string) error {
			w.archive = filepath.Join(w.t.TempDir(), "dist")
			return command(root(), "scripts/package.sh", version, w.archive)
		})
		sc.Step(`^the archive is named for version "([^"]*)" and this system$`, func(version string) error {
			_, err := os.Stat(filepath.Join(w.archive, packed(version)))
			return err
		})
		sc.Step(`^the binary in it says it is version "([^"]*)"$`, func(version string) error {
			w.unpacked = w.t.TempDir()
			if err := command(w.unpacked, "tar", "-xzf", filepath.Join(w.archive, packed(version))); err != nil {
				return err
			}
			said, err := exec.Command(filepath.Join(w.unpacked, "cairn"), "--version").Output()
			if err != nil {
				return err
			}
			if !strings.Contains(string(said), version) {
				return fmt.Errorf("the binary says %q", said)
			}
			return nil
		})
		sc.Step(`^the checksums name it$`, func() error {
			sums, err := os.ReadFile(filepath.Join(w.archive, "SHA256SUMS"))
			if err != nil {
				return err
			}
			if !strings.Contains(string(sums), "  cairn-") || !strings.Contains(string(sums), runtime.GOOS+"-"+runtime.GOARCH) {
				return fmt.Errorf("the checksums say:\n%s", sums)
			}
			return nil
		})

		sc.Step(`^I unpack it somewhere git has never been$`, func() error {
			found, err := filepath.Glob(filepath.Join(w.archive, "cairn-*-src.tar.gz"))
			if err != nil {
				return err
			}
			if len(found) != 1 {
				return fmt.Errorf("the archive made is %v, want one source archive", found)
			}
			w.unpacked = w.t.TempDir()
			if err := command(w.unpacked, "tar", "-xzf", found[0]); err != nil {
				return err
			}
			inside, err := filepath.Glob(filepath.Join(w.unpacked, "cairn-*"))
			if err != nil {
				return err
			}
			if len(inside) != 1 {
				return fmt.Errorf("the archive unpacks to %v, want one directory", inside)
			}
			w.unpacked = inside[0]
			if _, err := os.Stat(filepath.Join(w.unpacked, ".git")); err == nil {
				return fmt.Errorf("the archive carries a git history")
			}
			return nil
		})

		sc.Step(`^a build there would be stamped with version "([^"]*)"$`, func(want string) error {
			return w.stampedThere(0, want)
		})

		sc.Step(`^a build there would be stamped with the commit the archive was made from$`, func() error {
			head, err := exec.Command("git", "-C", root(), "rev-parse", "HEAD").Output()
			if err != nil {
				return err
			}
			return w.stampedThere(1, strings.TrimSpace(string(head)))
		})

		sc.Step(`^a build there would be stamped with the date of that commit$`, func() error {
			ask := exec.Command("git", "-C", root(), "show", "-s", "--format=%cd",
				"--date=format-local:%Y-%m-%dT%H:%M:%SZ", "HEAD")
			ask.Env = append(os.Environ(), "TZ=UTC")
			when, err := ask.Output()
			if err != nil {
				return err
			}
			return w.stampedThere(2, strings.TrimSpace(string(when)))
		})

		sc.Step(`^the archive says how to build and install it$`, func() error {
			said, err := os.ReadFile(filepath.Join(w.unpacked, "INSTALL.md"))
			if err != nil {
				return err
			}
			for _, must := range []string{"make build", "bin/cairn --version"} {
				if !strings.Contains(string(said), must) {
					return fmt.Errorf("INSTALL.md says nothing of %q", must)
				}
			}
			return nil
		})

		sc.Step(`^the archive's install guide is written in ASCII alone$`, func() error {
			said, err := os.ReadFile(filepath.Join(w.unpacked, "INSTALL.md"))
			if err != nil {
				return err
			}
			for i, line := range strings.Split(string(said), "\n") {
				for _, r := range line {
					if r > 127 {
						return fmt.Errorf("INSTALL.md line %d carries %q: %s", i+1, r, line)
					}
				}
			}
			return nil
		})

		sc.Step(`^a repository on a branch called "([^"]*)"$`, func(branch string) error {
			w.repo = w.t.TempDir()
			if err := os.MkdirAll(filepath.Join(w.repo, "scripts"), 0o755); err != nil {
				return err
			}
			for _, name := range []string{"release.sh", "dist.sh", "changelog.sh", "sums.sh"} {
				script, err := os.ReadFile(filepath.Join(root(), "scripts", name))
				if err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(w.repo, "scripts", name), script, 0o755); err != nil {
					return err
				}
			}
			for _, args := range [][]string{
				{"init", "-q", "-b", branch},
				{"-c", "user.name=t", "-c", "user.email=t@t", "add", "."},
				{"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "feat: begin"},
			} {
				if err := command(w.repo, "git", args...); err != nil {
					return err
				}
			}
			return nil
		})

		sc.Step(`^a version "([^"]*)" made there$`, func(version string) error {
			return command(w.repo, "git", "-c", "user.name=t", "-c", "user.email=t@t", "tag", "-a", version, "-m", version)
		})

		sc.Step(`^the commits (".*")$`, func(said string) error {
			for _, message := range quoted.FindAllStringSubmatch(said, -1) {
				if err := command(w.repo, "git", "-c", "user.name=t", "-c", "user.email=t@t",
					"commit", "-q", "--allow-empty", "-m", message[1]); err != nil {
					return err
				}
			}
			return nil
		})

		sc.Step(`^(?:I make )?its source archive(?: made)? as version "([^"]*)"$`, func(version string) error {
			if w.archive == "" {
				w.archive = filepath.Join(w.t.TempDir(), "dist")
			}
			return command(w.repo, "bash", filepath.Join("scripts", "dist.sh"), version, w.archive)
		})

		sc.Step(`^the archive's changelog says, under "([^"]*)", "([^"]*)" among the (features|fixes|breaking changes)$`, func(version, what, kind string) error {
			section, err := w.changelogSection(version, map[string]string{
				"features": "### Features", "fixes": "### Fixes", "breaking changes": "### Breaking changes",
			}[kind])
			if err != nil {
				return err
			}
			if !strings.Contains(section, what) {
				return fmt.Errorf("the %s of %s say nothing of %q:\n%s", kind, version, what, section)
			}
			return nil
		})

		sc.Step(`^the archive's changelog says nothing of "([^"]*)"$`, func(what string) error {
			said, err := w.changelog()
			if err != nil {
				return err
			}
			if strings.Contains(said, what) {
				return fmt.Errorf("the changelog speaks of %q:\n%s", what, said)
			}
			return nil
		})

		sc.Step(`^the checksums name "([^"]*)" and "([^"]*)"$`, func(first, second string) error {
			sums, err := os.ReadFile(filepath.Join(w.archive, "SHA256SUMS"))
			if err != nil {
				return err
			}
			for _, name := range []string{first, second} {
				if !strings.Contains(string(sums), "  "+name+"\n") {
					return fmt.Errorf("the checksums say nothing of %s:\n%s", name, sums)
				}
			}
			return nil
		})

		sc.Step(`^a file changed and not committed$`, func() error {
			return os.WriteFile(filepath.Join(w.repo, "notes.txt"), []byte("not committed\n"), 0o644)
		})

		sc.Step(`^I make a release there$`, func() error {
			run := exec.Command("bash", filepath.Join(w.repo, "scripts", "release.sh"))
			run.Dir = w.repo
			out, err := run.CombinedOutput()
			w.out = test.Result{Stderr: string(out), Code: exitOK}
			if err != nil {
				w.out.Code = 1
			}
			return nil
		})

		sc.Step(`^it refuses, saying "([^"]*)"$`, func(want string) error {
			if w.out.Code == exitOK || !strings.Contains(w.out.Stderr, want) {
				return fmt.Errorf("it answered %d, saying %q", w.out.Code, w.out.Stderr)
			}
			return nil
		})

		sc.Step(`^no tag was made$`, func() error {
			tags, err := exec.Command("git", "-C", w.repo, "tag").Output()
			if err != nil {
				return err
			}
			if len(strings.TrimSpace(string(tags))) != 0 {
				return fmt.Errorf("tags were made: %s", tags)
			}
			return nil
		})

		sc.Step("^I run `cairn --version`$", func() error {
			w.out = run(w.t, "--version")
			if w.out.Code != exitOK {
				return fmt.Errorf("exit %d (stderr: %s)", w.out.Code, w.out.Stderr)
			}
			return nil
		})

		sc.Step(`^it reports the version as "([^"]*)"$`, func(want string) error {
			first := strings.SplitN(w.out.Stdout, "\n", 2)[0]
			if got := strings.TrimSpace(first); got != "cairn "+want {
				return fmt.Errorf("first line = %q, want %q", got, "cairn "+want)
			}
			return nil
		})

		sc.Step(`^it reports the commit as "([^"]*)"$`, func(want string) error {
			return wantLine(w, "commit", want)
		})

		sc.Step(`^it reports the build time as "([^"]*)"$`, func(want string) error {
			return wantLine(w, "built", want)
		})

		sc.Step(`^it reports the commit date as "([^"]*)"$`, func(want string) error {
			return wantLine(w, "commit date", want)
		})

		sc.Step(`^it reports a real Go toolchain$`, func() error {
			if got := w.line("go"); got != runtime.Version() {
				return fmt.Errorf("go = %q, want the running toolchain %q", got, runtime.Version())
			}
			return nil
		})

		sc.Step(`^it reports a real platform$`, func() error {
			want := runtime.GOOS + "/" + runtime.GOARCH
			if got := w.line("platform"); got != want {
				return fmt.Errorf("platform = %q, want %q", got, want)
			}
			return nil
		})
	}
}

func wantLine(w *versionWorld, key, want string) error {
	if got := w.line(key); got != want {
		return fmt.Errorf("%s = %q, want %q", key, got, want)
	}
	return nil
}

func TestVersionWorld_putsTheBuildBackWhenTheScenarioEnds(t *testing.T) {
	found := stamps
	t.Cleanup(func() { stamps = found })

	w := newVersionWorld(t)
	w.stamp(&w.version, "what the scenario set")
	if got := stamps().Version; got != "what the scenario set" {
		t.Fatalf("version = %q during the scenario", got)
	}

	w.endScenario()
	if got := stamps().Version; got != found().Version {
		t.Errorf("version = %q, want %q put back for the next scenario", got, found().Version)
	}
}

func TestAcceptance_Version_DegradesCleanly(t *testing.T) {
	runAcceptance(t, "version.feature", initVersion(t), "@version-degraded")
}

func TestAcceptance_Version_ReportsItsStamps(t *testing.T) {
	runAcceptance(t, "version.feature", initVersion(t), "@version-stamped")
}

func TestAcceptance_Version_ComesWithASourceArchive(t *testing.T) {
	runAcceptance(t, "version.feature", initVersion(t), "@version-from-a-source-archive,@version-install-reads-anywhere,@version-brew-formula,@version-brew-formula-for-its-minor,@version-brew-formula-needs-every-binary")
}

func TestAcceptance_Version_ComesAsABinaryForEachSystem(t *testing.T) {
	runAcceptance(t, "version.feature", initVersion(t), "@version-binary-archive")
}

func TestAcceptance_Version_IsMadeOnlyFromMainAndClean(t *testing.T) {
	runAcceptance(t, "version.feature", initVersion(t), "@version-is-made-on-main,@version-is-made-from-a-clean-tree")
}

func (w *versionWorld) changelog() (string, error) {
	found, err := filepath.Glob(filepath.Join(w.archive, "cairn-*-src.tar.gz"))
	if err != nil {
		return "", err
	}
	if len(found) == 0 {
		return "", fmt.Errorf("no source archive was made")
	}
	slices.Sort(found)
	last := found[len(found)-1]
	inside := strings.TrimSuffix(filepath.Base(last), "-src.tar.gz") + "/CHANGELOG.md"
	said, err := exec.Command("tar", "-xzOf", last, inside).Output()
	if err != nil {
		return "", fmt.Errorf("the archive holds no changelog: %w", err)
	}
	return string(said), nil
}

func (w *versionWorld) changelogSection(version, heading string) (string, error) {
	said, err := w.changelog()
	if err != nil {
		return "", err
	}
	_, notes, found := strings.Cut(said, "## "+version)
	if !found {
		return "", fmt.Errorf("the changelog has no %s:\n%s", version, said)
	}
	notes, _, _ = strings.Cut(notes, "\n## ")
	_, part, found := strings.Cut(notes, heading)
	if !found {
		return "", fmt.Errorf("%s has no %q:\n%s", version, heading, notes)
	}
	part, _, _ = strings.Cut(part, "\n### ")
	return part, nil
}

func TestAcceptance_Version_SaysWhatChangedAndKeepsEverySum(t *testing.T) {
	runAcceptance(t, "version.feature", initVersion(t), "@version-archive-says-what-changed,@version-sums-kept")
}
