package task

import (
	"reflect"
	"strings"
	"testing"
)

// Archiving is said archiving, to a person and to an agent alike: the
// command is archive, and what is said of it uses no other word.
func TestArchivingIsSaidArchiving(t *testing.T) {
	if said := ErrPartGoesWithItsWork.Error(); !strings.Contains(said, "archived") {
		t.Errorf("a part refused archiving says %q", said)
	}
	field, _ := reflect.TypeOf(Facts{}).FieldByName("Archived")
	if said := field.Tag.Get("jsonschema"); !strings.Contains(said, "archived") {
		t.Errorf("an agent is told archived means %q", said)
	}
}
