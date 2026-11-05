package test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// Release is a release of cairn as GitHub hands it over: its tag, the system its archive was made for, the binary that archive holds — none when nil — whether SHA256SUMS gives another sum for it, and a count of what it was asked, when there is one to keep.
type Release struct {
	Tag      string
	Platform string
	Binary   []byte
	WrongSum bool
	Asked    *atomic.Int32
}

// Serve starts a GitHub serving r as the newest release of cairn; the server's URL is where that release is described.
func (r Release) Serve(t *testing.T) *httptest.Server {
	t.Helper()
	goos, arch, _ := strings.Cut(r.Platform, "/")
	name := "cairn-" + r.Tag + "-" + goos + "-" + arch + ".tar.gz"
	archive := r.tarred(t)
	if goos == "windows" {
		name = strings.TrimSuffix(name, ".tar.gz") + ".zip"
		archive = r.zipped(t)
	}
	summed := archive
	if r.WrongSum {
		summed = append(bytes.Clone(archive), '!')
	}
	sum := sha256.Sum256(summed)
	files := map[string][]byte{
		name:           archive,
		"SHA256SUMS":   []byte(hex.EncodeToString(sum[:]) + "  " + name + "\n"),
		"CHANGELOG.md": []byte("# Changelog\n"),
	}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if r.Asked != nil {
			r.Asked.Add(1)
		}
		if said, found := files[strings.TrimPrefix(req.URL.Path, "/")]; found {
			_, _ = w.Write(said)
			return
		}
		type asset struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		}
		var assets []asset
		for named := range files {
			assets = append(assets, asset{named, server.URL + "/" + named})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": r.Tag, "assets": assets})
	}))
	return server
}

func (r Release) tarred(t *testing.T) []byte {
	t.Helper()
	var out bytes.Buffer
	zipped := gzip.NewWriter(&out)
	archive := tar.NewWriter(zipped)
	for _, f := range []struct {
		name string
		says []byte
	}{{"./README.md", []byte("# cairn\n")}, {"./cairn", r.Binary}} {
		if f.says == nil {
			continue
		}
		if err := archive.WriteHeader(&tar.Header{Name: f.name, Mode: 0o755, Size: int64(len(f.says)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := archive.Write(f.says); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zipped.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func (r Release) zipped(t *testing.T) []byte {
	t.Helper()
	var out bytes.Buffer
	archive := zip.NewWriter(&out)
	for name, says := range map[string][]byte{"README.md": []byte("# cairn\n"), "cairn.exe": r.Binary} {
		if says == nil {
			continue
		}
		f, err := archive.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(says); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
