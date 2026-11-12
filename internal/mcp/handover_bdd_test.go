package mcp_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cucumber/godog"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/test"
)

const quietFor = 300 * time.Millisecond

func (w *world) registerHandOver(sc *godog.ScenarioContext) {
	sc.Step(`^a cairn server started from a cairn on disk$`, w.aCairnServerStartedFromACairnOnDisk)
	sc.Step(`^the page "([^"]*)" turned back on$`, func(page string) error {
		return w.WriteConfig(fmt.Sprintf("[pages]\n%s = true\n", page))
	})
	sc.Step(`^the page "([^"]*)" turned off$`, func(page string) error {
		return w.WriteConfig(fmt.Sprintf("[pages]\n%s = false\n", page))
	})
	sc.Step(`^the configuration written again, the page "([^"]*)" still off and the weekends hidden$`, func(page string) error {
		return w.WriteConfig(fmt.Sprintf("[pages]\n%s = false\n\n[time]\nweekends = \"hidden\"\n", page))
	})
	sc.Step(`^a newer cairn installed in its place$`, func() error {
		built, err := os.ReadFile(test.Built(w.T))
		if err != nil {
			return err
		}
		return w.installed(built)
	})
	sc.Step(`^something that cannot be run installed in its place$`, func() error {
		return w.installed([]byte("not a program\n"))
	})
	sc.Step(`^the client is told the tools changed$`, func() error {
		select {
		case <-w.told:
			return nil
		case <-time.After(5 * time.Second):
			return fmt.Errorf("the client was never told the tools changed")
		}
	})
	sc.Step(`^the client is not told the tools changed$`, func() error {
		select {
		case <-w.told:
			return fmt.Errorf("the client was told the tools changed, and they had not")
		case <-time.After(quietFor):
			return nil
		}
	})
}

func (w *world) aCairnServerStartedFromACairnOnDisk() error {
	built, err := os.ReadFile(test.Built(w.T))
	if err != nil {
		return err
	}
	w.onDisk = filepath.Join(w.T.TempDir(), "cairn")
	if err := os.WriteFile(w.onDisk, built, 0o700); err != nil {
		return err
	}
	client := w.client()
	cmd := exec.Command(w.onDisk, "mcp", "serve", "--access", "full")
	cmd.Env = w.Environ()
	cmd.Stderr = io.Discard
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	session, err := client.Connect(ctx, &sdk.CommandTransport{Command: cmd}, nil)
	if err != nil {
		return err
	}
	w.session = session
	w.OnEnd(func() { _ = session.Close() })
	return nil
}

func (w *world) client() *sdk.Client {
	w.told = make(chan struct{}, 1)
	return sdk.NewClient(&sdk.Implementation{Name: "acceptance", Version: "0"}, &sdk.ClientOptions{
		ToolListChangedHandler: func(context.Context, *sdk.ToolListChangedRequest) {
			select {
			case w.told <- struct{}{}:
			default:
			}
		},
	})
}

func (w *world) installed(content []byte) error {
	if w.onDisk == "" {
		return fmt.Errorf("no cairn on disk: start a server from one first")
	}
	beside := w.onDisk + ".new"
	if err := os.WriteFile(beside, content, 0o700); err != nil {
		return err
	}
	return os.Rename(beside, w.onDisk)
}
