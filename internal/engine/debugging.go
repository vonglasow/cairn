package engine

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"cairn/internal/build"
	"cairn/internal/debug"
)

var noting struct {
	mu      sync.Mutex
	log     *debug.Log
	version int64
	known   bool
}

// NoteTo has every operation asked from now on written down in l, once each, and the version a board reads each time it is another; nil stops it.
func NoteTo(l *debug.Log) {
	noting.mu.Lock()
	defer noting.mu.Unlock()
	noting.log, noting.known = l, false
}

func notingTo() *debug.Log {
	noting.mu.Lock()
	defer noting.mu.Unlock()
	return noting.log
}

func noteDone(op Operation, in Input, from time.Time, answer Answer, err error) {
	l := notingTo()
	if l == nil {
		return
	}
	said := op.Line() + givenIn(l, op, in) + " → " + cameTo(l, answer, err) + " " + time.Since(from).Round(time.Millisecond).String()
	l.Say("engine", said)
}

func noteVersion(version int64, err error) {
	noting.mu.Lock()
	l := noting.log
	changed := l != nil && err == nil && (!noting.known || version != noting.version)
	if changed {
		noting.version, noting.known = version, true
	}
	noting.mu.Unlock()
	if changed {
		l.Say("engine", "version "+strconv.FormatInt(version, 10))
	}
}

func givenIn(l *debug.Log, op Operation, in Input) string {
	kinds := map[string]Argument{}
	for _, a := range op.Takes {
		kinds[a.Name] = a
	}
	var said []string
	for name, value := range in.Given {
		said = append(said, name+"="+shown(l, kinds[name], value))
	}
	for name, values := range in.Listed {
		var each []string
		for _, value := range values {
			each = append(each, shown(l, kinds[name], value))
		}
		said = append(said, name+"=["+strings.Join(each, " ")+"]")
	}
	if in.Confirmed {
		said = append(said, "yes")
	}
	slices.Sort(said)
	if len(said) == 0 {
		return ""
	}
	return " " + strings.Join(said, " ")
}

func shown(l *debug.Log, a Argument, value string) string {
	switch {
	case a.Kind == Number || a.Kind == Day || a.Kind == Duration || a.Kind == Bool:
		return value
	case slices.Contains(a.OneOf, value):
		return value
	case value == "true" || value == "false":
		return value
	}
	if _, err := strconv.Atoi(value); err == nil {
		return value
	}
	return l.Text(value)
}

func cameTo(l *debug.Log, answer Answer, err error) string {
	if err != nil {
		refusal := RefusalOf(err)
		if refusal.Kind == KindFailed {
			return "failed " + chainOf(err) + " " + l.Text(err.Error())
		}
		return refusal.Kind + " " + refusal.Reason
	}
	return strings.TrimSpace("ok " + summary(answer))
}

func chainOf(err error) string {
	var types []string
	for err != nil {
		types = append(types, fmt.Sprintf("%T", err))
		err = errors.Unwrap(err)
	}
	return strings.Join(types, "→")
}

func summary(a Answer) string {
	var said []string
	if a.Task != nil {
		c := a.Task.Carried()
		said = append(said, fmt.Sprintf("task #%d %s", c.Number, c.State))
	}
	v := reflect.ValueOf(a)
	for i := range v.NumField() {
		field, name := v.Field(i), v.Type().Field(i).Name
		if name == "Task" || field.IsZero() {
			continue
		}
		named := string(unicode.ToLower(rune(name[0]))) + name[1:]
		if field.Kind() == reflect.Pointer && field.Elem().Kind() == reflect.Slice {
			field = field.Elem()
		}
		switch field.Kind() {
		case reflect.Slice, reflect.Map:
			if field.Len() > 0 {
				said = append(said, fmt.Sprintf("%s %d", named, field.Len()))
			}
		case reflect.Pointer:
			if field.Elem().Kind() == reflect.Int {
				said = append(said, fmt.Sprintf("%s %d", named, field.Elem().Int()))
				continue
			}
			said = append(said, named)
		default:
			said = append(said, named)
		}
	}
	return strings.Join(said, ", ")
}

// DebugHeader is what a debug file begins with, a fact a line: this build, the machine it runs on, the configuration as config show says it with the home written ~, the store in numbers — never what it holds — and the command, written the way the caller masked it.
func DebugHeader(at Places, command string) [][2]string {
	b := build.Read()
	facts := [][2]string{
		{"recorded", time.Now().Format("2006-01-02 15:04:05 -0700")},
		{"cairn", b.Version},
		{"commit", b.Commit},
		{"go", b.Go},
	}
	facts = append(facts, debug.Machine(runtime.GOOS, runtime.GOARCH, os.Getenv, debug.Ran)...)
	for _, line := range configLines(at.Home) {
		facts = append(facts, [2]string{"config", line})
	}
	facts = append(facts, storeInNumbers(at)...)
	return append(facts, [2]string{"command", command})
}

func configLines(h Home) []string {
	answer, err := doAt(ConfigShow, Places{Home: h}, Input{}, Acting{})
	if err != nil || answer.Config == nil {
		return []string{"cannot be read"}
	}
	said := configText(*answer.Config)
	if h.Dir != "" {
		said = strings.ReplaceAll(said, h.Dir, "~")
	}
	var lines []string
	for _, line := range strings.Split(said, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, strings.Join(strings.Fields(line), " "))
		}
	}
	return lines
}

func storeInNumbers(at Places) [][2]string {
	if _, err := os.Stat(at.Store); errors.Is(err, fs.ErrNotExist) {
		return [][2]string{{"store", "none yet"}}
	}
	held, err := Open(at.Store)
	if err != nil {
		return [][2]string{{"store", "cannot be read: " + RefusalOf(err).Reason}}
	}
	defer func() { _ = held.Close() }()
	listed, err := held.run(TaskList, Input{Given: map[string]string{"with-archived": "true"}})
	if err != nil {
		return [][2]string{{"store", "cannot be read: " + RefusalOf(err).Reason}}
	}
	goals, _ := held.run(ObjectiveList, Input{})
	columns := map[string]int{}
	labels := map[string]bool{}
	archived := 0
	for _, t := range listed.Tasks {
		c := t.Carried()
		columns[string(c.State)]++
		if c.Archived {
			archived++
		}
		for _, label := range c.Labels {
			labels[label] = true
		}
	}
	var counted []string
	for _, state := range States() {
		if columns[state] > 0 {
			counted = append(counted, fmt.Sprintf("%d %s", columns[state], state))
		}
	}
	if len(counted) == 0 {
		counted = []string{"no task"}
	}
	facts := [][2]string{
		{"store", strings.Join(counted, ", ")},
		{"archived", strconv.Itoa(archived)},
		{"objectives", strconv.Itoa(len(goals.Objectives))},
		{"labels", strconv.Itoa(len(labels))},
	}
	if changes, err := doAt(MigrationList, Places{Store: at.Store}, Input{}, Acting{}); err == nil {
		applied := 0
		for _, m := range changes.Changes {
			if m.Carried().Applied {
				applied++
			}
		}
		facts = append(facts, [2]string{"schema", fmt.Sprintf("%d of %d changes applied", applied, len(changes.Changes))})
	}
	return facts
}

// DebugReported says where debug report wrote its archive, and how many debug files it packed.
type DebugReported struct {
	Written string `json:"written"`
	Files   int    `json:"files"`
}

// DebugReport packs every file debug mode kept, and a header made at that moment, into one archive to send.
var DebugReport = Operation{
	Thing:   "debug",
	Verb:    "report",
	Short:   "Pack what debug mode wrote down into one archive to send",
	Does:    "Pack every file debug mode kept beside the store, and a header saying what this machine and this store are now, into one zip archive, written in the folder named or the current one, to send to whoever looks into a bug. Nothing a person wrote is in it: debug mode masks every text.",
	Example: "cairn debug report",
	Level:   Write,
	Needs:   NoStore,
	Again:   DoesItAgain,
	Takes: []Argument{
		{Name: "to", Kind: Text, Means: "the folder to write the archive in; the current one when not said"},
	},
	Answers: DebugReported{},
	Run: func(on On, in Input) (Answer, error) {
		to := in.Given["to"]
		if to == "" {
			here, err := os.Getwd()
			if err != nil {
				return Answer{}, err
			}
			to = here
		}
		at := Places{Store: on.Path, Notes: on.Notes, Home: on.Home}
		written := filepath.Join(to, "cairn-debug-"+time.Now().Format("20060102-150405")+".zip")
		packed, err := packed(written, at)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Written: &written, Packed: &packed}, nil
	},
}

func packed(written string, at Places) (count int, err error) {
	file, err := os.OpenFile(written, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return 0, err
	}
	defer func() {
		if closed := file.Close(); err == nil {
			err = closed
		}
	}()
	archive := zip.NewWriter(file)
	header, err := archive.Create("header.txt")
	if err != nil {
		return 0, err
	}
	for _, f := range DebugHeader(at, "debug report") {
		if _, err := fmt.Fprintf(header, "# %-16s %s\n", f[0], f[1]); err != nil {
			return 0, err
		}
	}
	entries, err := os.ReadDir(at.Debug())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return 0, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".log") {
			continue
		}
		if err := packOne(archive, filepath.Join(at.Debug(), e.Name())); err != nil {
			return 0, err
		}
		count++
	}
	return count, archive.Close()
}

func packOne(archive *zip.Writer, path string) error {
	read, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = read.Close() }()
	into, err := archive.Create("debug/" + filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(into, read)
	return err
}
