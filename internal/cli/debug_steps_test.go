package cli

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"cairn/internal/engine"

	"github.com/cucumber/godog"
)

func (w *world) registerDebug(sc *godog.ScenarioContext) {
	var oldest string
	folder := func() string { return engine.Places{Store: w.StorePath}.Debug() }
	files := func() ([]string, error) {
		entries, err := os.ReadDir(folder())
		if err != nil {
			return nil, err
		}
		var out []string
		for _, e := range entries {
			out = append(out, e.Name())
		}
		slices.Sort(out)
		return out, nil
	}

	newest := func() (string, error) {
		held, err := files()
		if err != nil || len(held) == 0 {
			return "", fmt.Errorf("the debug folder holds %v: %w", held, err)
		}
		read, err := os.ReadFile(filepath.Join(folder(), held[len(held)-1]))
		return string(read), err
	}
	var archive string
	var before []string

	sc.Step(`^the debug folder cannot be made$`, func() error {
		if err := os.MkdirAll(filepath.Dir(folder()), 0o700); err != nil {
			return err
		}
		return os.WriteFile(folder(), []byte("a file where the folder would go"), 0o600)
	})
	sc.Step(`^it says on standard error that it cannot write down$`, func() error {
		if !strings.Contains(w.Out.Stderr, "debug: cannot write down") {
			return fmt.Errorf("standard error says %q", w.Out.Stderr)
		}
		return nil
	})
	sc.Step(`^the debug file says "([^"]*)"$`, func(said string) error {
		read, err := newest()
		if err != nil {
			return err
		}
		if !strings.Contains(read, said) {
			return fmt.Errorf("the debug file does not say %q:\n%s", said, read)
		}
		return nil
	})
	sc.Step(`^the debug file never says "([^"]*)"$`, func(said string) error {
		held, err := files()
		if err != nil {
			return err
		}
		for _, name := range held {
			read, err := os.ReadFile(filepath.Join(folder(), name))
			if err != nil {
				return err
			}
			if strings.Contains(string(read), said) {
				return fmt.Errorf("%s says %q:\n%s", name, said, read)
			}
		}
		return nil
	})
	sc.Step(`^I work in a folder of my own$`, func() error {
		w.T.Chdir(w.T.TempDir())
		return nil
	})
	sc.Step("^I ran `([^`]*)`$", func(line string) error {
		w.Run(line)
		if w.Out.Code != 0 {
			return fmt.Errorf("%s exited %d: %s", line, w.Out.Code, w.Out.Stderr)
		}
		held, err := files()
		before = held
		return err
	})
	sc.Step(`^it says where the archive is$`, func() error {
		for _, word := range strings.Fields(w.Out.Stdout) {
			if word = strings.TrimSuffix(word, ","); strings.HasSuffix(word, ".zip") {
				archive = word
			}
		}
		if _, err := os.Stat(archive); err != nil {
			return fmt.Errorf("standard output says %q: %w", w.Out.Stdout, err)
		}
		return nil
	})
	sc.Step(`^the archive holds every debug file and a header$`, func() error {
		packed, err := zip.OpenReader(archive)
		if err != nil {
			return err
		}
		defer func() { _ = packed.Close() }()
		var names []string
		for _, f := range packed.File {
			names = append(names, f.Name)
		}
		if !slices.Contains(names, "header.txt") {
			return fmt.Errorf("the archive holds %v, no header", names)
		}
		for _, name := range before {
			if !slices.Contains(names, "debug/"+name) {
				return fmt.Errorf("the archive holds %v, not %s", names, name)
			}
		}
		return nil
	})
	sc.Step(`^debug mode is on$`, func() error {
		w.Setenv("CAIRN_DEBUG", "1")
		return nil
	})
	sc.Step(`^the debug folder already holds (\d+) files$`, func(n int) error {
		if err := os.MkdirAll(folder(), 0o700); err != nil {
			return err
		}
		for i := range n {
			name := fmt.Sprintf("20000101-%06d-task-list-old.log", i)
			if err := os.WriteFile(filepath.Join(folder(), name), nil, 0o600); err != nil {
				return err
			}
		}
		held, err := files()
		if err != nil {
			return err
		}
		oldest = held[0]
		return nil
	})
	sc.Step(`^there is no debug folder beside the store$`, func() error {
		if _, err := os.Stat(folder()); !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s is there: %w", folder(), err)
		}
		return nil
	})
	sc.Step(`^the debug folder beside the store holds (\d+) files?$`, func(n int) error {
		held, err := files()
		if err != nil {
			return err
		}
		if len(held) != n {
			return fmt.Errorf("the debug folder holds %d files: %v", len(held), held)
		}
		return nil
	})
	sc.Step(`^it says on standard error where that file is$`, func() error {
		held, err := files()
		if err != nil || len(held) != 1 {
			return fmt.Errorf("the debug folder holds %v, %w", held, err)
		}
		if at := filepath.Join(folder(), held[0]); !strings.Contains(w.Out.Stderr, at) {
			return fmt.Errorf("standard error says %q, not where %s is", w.Out.Stderr, at)
		}
		return nil
	})
	sc.Step(`^the oldest of them is gone$`, func() error {
		held, err := files()
		if err != nil {
			return err
		}
		if slices.Contains(held, oldest) {
			return fmt.Errorf("%s is still there", oldest)
		}
		return nil
	})
}
