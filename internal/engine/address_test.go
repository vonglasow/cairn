package engine

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestStoreAddress_prefersWhatTheEnvironmentSays(t *testing.T) {
	env := func(string) (string, bool) { return "/somewhere/of/my/own.db", true }
	home := func() (string, error) { return "/home/nobody", nil }

	got, err := StoreAddress(env, home)
	if err != nil {
		t.Fatalf("StoreAddress: %v", err)
	}
	if got != "/somewhere/of/my/own.db" {
		t.Errorf("StoreAddress = %q, want what the environment said", got)
	}
}

func TestStoreAddress_fallsBackToTheDefaultPlace(t *testing.T) {
	env := func(string) (string, bool) { return "", false }
	home := func() (string, error) { return "/home/nobody", nil }

	got, err := StoreAddress(env, home)
	if err != nil {
		t.Fatalf("StoreAddress: %v", err)
	}
	if want := filepath.Join("/home/nobody", ".cairn", "cairn.db"); got != want {
		t.Errorf("StoreAddress = %q, want %q", got, want)
	}
}

func TestStoreAddress_ignoresAnEnvironmentVariableSetToNothing(t *testing.T) {
	env := func(string) (string, bool) { return "", true }
	home := func() (string, error) { return "/home/nobody", nil }

	got, err := StoreAddress(env, home)
	if err != nil {
		t.Fatalf("StoreAddress: %v", err)
	}
	if want := filepath.Join("/home/nobody", ".cairn", "cairn.db"); got != want {
		t.Errorf("StoreAddress = %q, want the default place: an empty value says nothing", got)
	}
}

func TestStoreAddress_reportsAHomeItCannotFind(t *testing.T) {
	lost := errors.New("no home here")
	env := func(string) (string, bool) { return "", false }
	home := func() (string, error) { return "", lost }

	if _, err := StoreAddress(env, home); !errors.Is(err, lost) {
		t.Errorf("err = %v, want the reason the home could not be found", err)
	}
}
