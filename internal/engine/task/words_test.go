package task

import (
	"reflect"
	"strings"
	"testing"
)

func TestArchivingIsSaidArchiving(t *testing.T) {
	if said := ErrAlreadyArchived.Error(); !strings.Contains(said, "archived") {
		t.Errorf("a task archived twice says %q", said)
	}
	field, _ := reflect.TypeOf(Facts{}).FieldByName("Archived")
	if said := field.Tag.Get("jsonschema"); !strings.Contains(said, "archived") {
		t.Errorf("an agent is told archived means %q", said)
	}
}
