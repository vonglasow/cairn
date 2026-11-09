package theme

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

//go:embed schemes/*.yaml
var carried embed.FS

// Carried is where a shipped scheme is said to come from: cairn itself.
const Carried = "shipped"

// DefaultSlug is the scheme the window wears when nothing names one, or when the one named cannot be worn.
const DefaultSlug = "default"

// ErrUnknown is what a scheme's name is said to be when neither cairn nor the person's folder holds it.
var ErrUnknown = errors.New("is not a scheme cairn sees")

var aSlug = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

const extension = ".yaml"

// Shipped is the slug of every scheme cairn carries, sorted.
func Shipped() []string {
	files, _ := fs.Glob(carried, "schemes/*"+extension)
	slugs := make([]string, 0, len(files))
	for _, file := range files {
		slugs = append(slugs, strings.TrimSuffix(filepath.Base(file), extension))
	}
	return slugs
}

// Find is the scheme named slug: the person's own, a file of that name in folder, before the one cairn carries.
func Find(folder, slug string) (Scheme, error) {
	if !aSlug.MatchString(slug) {
		return Scheme{}, fmt.Errorf("%q %w", slug, ErrUnknown)
	}
	if folder != "" {
		at := filepath.Join(folder, slug+extension)
		data, err := os.ReadFile(at)
		if err == nil {
			return Parse(slug, at, data)
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return Scheme{}, fmt.Errorf("%s (%s): %w", slug, at, err)
		}
	}
	data, err := carried.ReadFile("schemes/" + slug + extension)
	if err != nil {
		return Scheme{}, fmt.Errorf("%q %w", slug, ErrUnknown)
	}
	return Parse(slug, Carried, data)
}

// Default is the colours the window wears when nothing names a scheme: the scheme called default.
func Default() Colours {
	s, err := Find("", DefaultSlug)
	if err != nil {
		panic(err)
	}
	return s.Colours()
}

// Found is a scheme as it was looked for: its name, where it comes from, and what it holds, or why it cannot be worn.
type Found struct {
	Slug, From string
	Scheme     Scheme
	Err        error
}

// Seen is every scheme there is to wear, sorted by name: the ones cairn carries and the person's own in folder, the person's taking the place of one cairn carries under the same name.
func Seen(folder string) []Found {
	slugs := Shipped()
	var own []string
	if folder != "" {
		own, _ = filepath.Glob(filepath.Join(folder, "*"+extension))
	}
	for _, file := range own {
		if slug := strings.TrimSuffix(filepath.Base(file), extension); aSlug.MatchString(slug) && !slices.Contains(slugs, slug) {
			slugs = append(slugs, slug)
		}
	}
	slices.Sort(slugs)
	seen := make([]Found, 0, len(slugs))
	for _, slug := range slugs {
		from := Carried
		if at := filepath.Join(folder, slug+extension); slices.Contains(own, at) {
			from = at
		}
		s, err := Find(folder, slug)
		seen = append(seen, Found{Slug: slug, From: from, Scheme: s, Err: err})
	}
	return seen
}
