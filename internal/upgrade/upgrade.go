// Package upgrade takes the newest release of cairn from GitHub and puts its binary in the place of the cairn that was run.
package upgrade

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Latest is where the newest release of cairn is described.
var Latest = "https://api.github.com/repos/vonglasow/cairn/releases/latest"

var client = &http.Client{Timeout: 5 * time.Minute}

// What an upgrade can be turned down for.
var (
	ErrOutOfReach   = errors.New("the releases could not be read")
	ErrNoBinaryHere = errors.New("the release has no binary for this system")
	ErrNotItsSum    = errors.New("the archive does not match its sum in SHA256SUMS; nothing was replaced")
	ErrNotWritable  = errors.New("the place of this cairn cannot be written; nothing was replaced")
)

// Release is one release of cairn: its tag, and where each file it hands over is fetched from.
type Release struct {
	Tag   string
	files map[string]string
}

// Newest reads which release of cairn is the newest, and what it hands over.
func Newest() (Release, error) {
	said, err := fetched(Latest)
	if err != nil {
		return Release{}, err
	}
	var read struct {
		Tag    string `json:"tag_name"`
		Assets []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(said, &read); err != nil || read.Tag == "" {
		return Release{}, fmt.Errorf("%w: %s says no release", ErrOutOfReach, Latest)
	}
	release := Release{Tag: read.Tag, files: map[string]string{}}
	for _, asset := range read.Assets {
		release.files[asset.Name] = asset.URL
	}
	return release, nil
}

// Install fetches the archive made for platform, a GOOS/GOARCH, checks it against the release's SHA256SUMS, and puts the cairn it holds in binary's place.
func (r Release) Install(binary, platform string) error {
	goos, arch, _ := strings.Cut(platform, "/")
	name := "cairn-" + r.Tag + "-" + goos + "-" + arch + ".tar.gz"
	if goos == "windows" {
		name = strings.TrimSuffix(name, ".tar.gz") + ".zip"
	}
	at, found := r.files[name]
	if !found {
		return fmt.Errorf("%s, %s: %w", r.Tag, platform, ErrNoBinaryHere)
	}
	sumsAt, found := r.files["SHA256SUMS"]
	if !found {
		return fmt.Errorf("%s gives no SHA256SUMS: %w", r.Tag, ErrNotItsSum)
	}
	archive, err := fetched(at)
	if err != nil {
		return err
	}
	sums, err := fetched(sumsAt)
	if err != nil {
		return err
	}
	if !summed(sums, name, archive) {
		return fmt.Errorf("%s: %w", name, ErrNotItsSum)
	}
	held, err := unpacked(archive, goos)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return replaced(binary, held)
}

func fetched(url string) ([]byte, error) {
	answer, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOutOfReach, err)
	}
	defer func() { _ = answer.Body.Close() }()
	if answer.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s answered %s", ErrOutOfReach, url, answer.Status)
	}
	return io.ReadAll(answer.Body)
}

func summed(sums []byte, name string, archive []byte) bool {
	sum := sha256.Sum256(archive)
	for _, line := range strings.Split(string(sums), "\n") {
		if fields := strings.Fields(line); len(fields) == 2 && fields[1] == name {
			return fields[0] == hex.EncodeToString(sum[:])
		}
	}
	return false
}

func unpacked(archive []byte, goos string) ([]byte, error) {
	if goos == "windows" {
		return unzipped(archive)
	}
	zipped, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	files := tar.NewReader(zipped)
	for {
		header, err := files.Next()
		if err != nil {
			return nil, fmt.Errorf("no cairn in it: %w", ErrNoBinaryHere)
		}
		if header.Typeflag == tar.TypeReg && path.Base(header.Name) == "cairn" {
			return io.ReadAll(files)
		}
	}
}

func unzipped(archive []byte) ([]byte, error) {
	files, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, err
	}
	for _, f := range files.File {
		if path.Base(f.Name) == "cairn.exe" {
			opened, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer func() { _ = opened.Close() }()
			return io.ReadAll(opened)
		}
	}
	return nil, fmt.Errorf("no cairn.exe in it: %w", ErrNoBinaryHere)
}

func replaced(binary string, held []byte) error {
	written, err := os.CreateTemp(filepath.Dir(binary), ".cairn-upgrade-*")
	if err != nil {
		return fmt.Errorf("%s: %w: %w", binary, ErrNotWritable, err)
	}
	defer func() { _ = os.Remove(written.Name()) }()
	_, err = written.Write(held)
	if closing := written.Close(); err == nil {
		err = closing
	}
	if err != nil {
		return err
	}
	if err := os.Chmod(written.Name(), 0o755); err != nil {
		return err
	}
	if err := put(written.Name(), binary); err != nil {
		return fmt.Errorf("%s: %w: %w", binary, ErrNotWritable, err)
	}
	return nil
}
