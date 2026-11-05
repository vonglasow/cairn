package upgrade_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/upgrade"
	"cairn/test"
)

func newest(t *testing.T, r test.Release) upgrade.Release {
	t.Helper()
	server := r.Serve(t)
	t.Cleanup(server.Close)
	was := upgrade.Latest
	upgrade.Latest = server.URL + "/latest"
	t.Cleanup(func() { upgrade.Latest = was })
	release, err := upgrade.Newest()
	if err != nil {
		t.Fatal(err)
	}
	return release
}

func placed(t *testing.T) string {
	t.Helper()
	place := filepath.Join(t.TempDir(), "cairn")
	if err := os.WriteFile(place, []byte("before"), 0o755); err != nil {
		t.Fatal(err)
	}
	return place
}

func TestNewest_readsTheTag(t *testing.T) {
	if release := newest(t, test.Release{Tag: "v0.11.0", Platform: "linux/amd64"}); release.Tag != "v0.11.0" {
		t.Errorf("the newest release is %q", release.Tag)
	}
}

func TestInstall_takesTheBinaryOutOfAZipOnWindows(t *testing.T) {
	release := newest(t, test.Release{Tag: "v0.11.0", Platform: "windows/amd64", Binary: []byte("cairn.exe v0.11.0")})
	place := placed(t)
	if err := release.Install(place, "windows/amd64"); err != nil {
		t.Fatal(err)
	}
	if said, _ := os.ReadFile(place); string(said) != "cairn.exe v0.11.0" {
		t.Errorf("in its place: %q", said)
	}
}

func TestInstall_refusesAnArchiveForAnotherSystem(t *testing.T) {
	release := newest(t, test.Release{Tag: "v0.11.0", Platform: "linux/amd64", Binary: []byte("x")})
	if err := release.Install(placed(t), "linux/arm64"); !errors.Is(err, upgrade.ErrNoBinaryHere) {
		t.Errorf("err = %v, want ErrNoBinaryHere", err)
	}
}

func TestNewest_refusesWhatIsNotARelease(t *testing.T) {
	for name, answer := range map[string]http.HandlerFunc{
		"an error":      func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "rate limited", http.StatusForbidden) },
		"not JSON":      func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("<html>")) },
		"no tag at all": func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"assets":[]}`)) },
	} {
		server := httptest.NewServer(answer)
		was := upgrade.Latest
		upgrade.Latest = server.URL
		if _, err := upgrade.Newest(); !errors.Is(err, upgrade.ErrOutOfReach) {
			t.Errorf("%s: err = %v, want ErrOutOfReach", name, err)
		}
		upgrade.Latest = was
		server.Close()
	}
}

func TestInstall_refusesAnArchiveHoldingNoCairn(t *testing.T) {
	for _, platform := range []string{"linux/amd64", "windows/amd64"} {
		release := newest(t, test.Release{Tag: "v0.11.0", Platform: platform})
		place := placed(t)
		if err := release.Install(place, platform); !errors.Is(err, upgrade.ErrNoBinaryHere) {
			t.Errorf("%s: err = %v, want ErrNoBinaryHere", platform, err)
		}
		if said, _ := os.ReadFile(place); string(said) != "before" {
			t.Errorf("%s: in its place: %q", platform, said)
		}
	}
}

func releasing(t *testing.T, files map[string][]byte) upgrade.Release {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/latest" {
			assets := `{"tag_name":"v0.11.0","assets":[`
			for name := range files {
				assets += `{"name":"` + name + `","browser_download_url":"` + server.URL + "/" + name + `"},`
			}
			_, _ = w.Write([]byte(strings.TrimSuffix(assets, ",") + "]}"))
			return
		}
		said := files[strings.TrimPrefix(req.URL.Path, "/")]
		if said == nil {
			http.NotFound(w, req)
			return
		}
		_, _ = w.Write(said)
	}))
	t.Cleanup(server.Close)
	was := upgrade.Latest
	upgrade.Latest = server.URL + "/latest"
	t.Cleanup(func() { upgrade.Latest = was })
	release, err := upgrade.Newest()
	if err != nil {
		t.Fatal(err)
	}
	return release
}

func sumOf(name string, archive []byte) []byte {
	sum := sha256.Sum256(archive)
	return []byte(hex.EncodeToString(sum[:]) + "  " + name + "\n")
}

func TestInstall_refusesWhatCannotBeTrusted(t *testing.T) {
	const tarred, zipped = "cairn-v0.11.0-linux-amd64.tar.gz", "cairn-v0.11.0-windows-amd64.zip"
	broken := []byte("not an archive")
	for _, c := range []struct {
		name     string
		files    map[string][]byte
		platform string
		want     error
	}{
		{"no SHA256SUMS", map[string][]byte{tarred: broken}, "linux/amd64", upgrade.ErrNotItsSum},
		{"no sum for the archive", map[string][]byte{tarred: broken, "SHA256SUMS": sumOf("another", broken)}, "linux/amd64", upgrade.ErrNotItsSum},
		{"the archive gone", map[string][]byte{tarred: nil, "SHA256SUMS": sumOf(tarred, broken)}, "linux/amd64", upgrade.ErrOutOfReach},
		{"SHA256SUMS gone", map[string][]byte{tarred: broken, "SHA256SUMS": nil}, "linux/amd64", upgrade.ErrOutOfReach},
	} {
		place := placed(t)
		if err := releasing(t, c.files).Install(place, c.platform); !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
		if said, _ := os.ReadFile(place); string(said) != "before" {
			t.Errorf("%s: in its place: %q", c.name, said)
		}
	}
	for platform, name := range map[string]string{"linux/amd64": tarred, "windows/amd64": zipped} {
		place := placed(t)
		if err := releasing(t, map[string][]byte{name: broken, "SHA256SUMS": sumOf(name, broken)}).Install(place, platform); err == nil {
			t.Errorf("%s: an archive that is none was unpacked", platform)
		}
		if said, _ := os.ReadFile(place); string(said) != "before" {
			t.Errorf("%s: in its place: %q", platform, said)
		}
	}
}
