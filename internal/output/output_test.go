package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy error: %v", err)
	}
	return buf.String()
}

func TestUseJSON_ForceJSON(t *testing.T) {
	ForceJSON = true
	QuietMode = false
	defer func() { ForceJSON = false }()

	if !UseJSON() {
		t.Error("UseJSON() = false, want true when ForceJSON is set")
	}
}

func TestUseJSON_QuietMode(t *testing.T) {
	ForceJSON = false
	QuietMode = true
	defer func() { QuietMode = false }()

	if !UseJSON() {
		t.Error("UseJSON() = false, want true when QuietMode is set")
	}
}

func TestPrintMessage_JSON(t *testing.T) {
	ForceJSON = true
	QuietMode = false
	defer func() { ForceJSON = false }()

	out := captureStdout(t, func() {
		PrintMessage("hello world")
	})

	var parsed map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["message"] != "hello world" {
		t.Errorf("message = %q, want %q", parsed["message"], "hello world")
	}
}

func TestPrintMessage_Plain(t *testing.T) {
	ForceJSON = false
	QuietMode = false

	out := captureStdout(t, func() {
		PrintMessage("hello world")
	})

	trimmed := strings.TrimSpace(out)
	if !strings.Contains(trimmed, "hello world") {
		t.Errorf("output = %q, want it to contain %q", trimmed, "hello world")
	}
}

func TestPrint_QuietMode(t *testing.T) {
	ForceJSON = false
	QuietMode = true
	defer func() { QuietMode = false }()

	data := json.RawMessage(`[{"id":1}]`)
	out := captureStdout(t, func() {
		Print(data, nil)
	})

	trimmed := strings.TrimSpace(out)
	if trimmed != `[{"id":1}]` {
		t.Errorf("quiet output = %q, want raw data", trimmed)
	}
}

func TestPrint_JSONEnvelope(t *testing.T) {
	ForceJSON = true
	QuietMode = false
	defer func() { ForceJSON = false }()

	data := json.RawMessage(`{"name":"test"}`)
	breadcrumbs := []Breadcrumb{{Label: "details", Command: "app show 1"}}

	out := captureStdout(t, func() {
		Print(data, breadcrumbs)
	})

	var envelope Envelope
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &envelope); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	var parsed map[string]string
	if err := json.Unmarshal(envelope.Data, &parsed); err != nil {
		t.Fatalf("envelope data is not valid JSON: %v", err)
	}
	if parsed["name"] != "test" {
		t.Errorf("data.name = %q, want %q", parsed["name"], "test")
	}
	if len(envelope.Breadcrumbs) != 1 || envelope.Breadcrumbs[0].Label != "details" {
		t.Errorf("breadcrumbs = %v, want [{details app show 1}]", envelope.Breadcrumbs)
	}
}

func TestPrintWithPagination_QuietMode(t *testing.T) {
	ForceJSON = false
	QuietMode = true
	defer func() { QuietMode = false }()

	data := json.RawMessage(`[{"id":1}]`)
	pagination := json.RawMessage(`{"current_page_number":1,"total_pages":3,"total_records":25}`)

	out := captureStdout(t, func() {
		PrintWithPagination(data, pagination, nil)
	})

	trimmed := strings.TrimSpace(out)
	if trimmed != `[{"id":1}]` {
		t.Errorf("quiet output = %q, want raw data without pagination", trimmed)
	}
}

func TestBuildGrid(t *testing.T) {
	rows := [][]interface{}{
		{"closed", float64(8), nil},
		{"open", float64(3), float64(2)},
	}

	got := buildGrid(rows)
	want := [][]string{
		{"closed", "8", "-"},
		{"open", "3", "2"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("buildGrid() = %#v, want %#v", got, want)
	}
}

func TestRenderGrid(t *testing.T) {
	out := captureStdout(t, func() {
		renderGrid([]string{"NAME", "PRESENT"}, [][]string{{"closed", "8"}})
	})

	for _, want := range []string{"NAME", "PRESENT", "closed", "8", "─"} {
		if !strings.Contains(out, want) {
			t.Errorf("renderGrid() output %q missing %q", out, want)
		}
	}
}

func TestPrintTable_JSONPreservesRaw(t *testing.T) {
	ForceJSON = true
	QuietMode = false
	defer func() { ForceJSON = false }()

	raw := json.RawMessage(`[{"name":"closed","value":{"present":8,"previous":null}}]`)

	out := captureStdout(t, func() {
		PrintTable(raw, []string{"NAME", "PRESENT"}, [][]interface{}{{"closed", float64(8)}}, nil)
	})

	var env Envelope
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &env); err != nil {
		t.Fatalf("output is not valid JSON envelope: %v", err)
	}

	var got, want interface{}
	if err := json.Unmarshal(env.Data, &got); err != nil {
		t.Fatalf("envelope data invalid: %v", err)
	}
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("raw invalid: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("PrintTable() JSON data = %v, want raw payload %v", got, want)
	}
}

func TestPrintWithPagination_JSONEnvelope(t *testing.T) {
	ForceJSON = true
	QuietMode = false
	defer func() { ForceJSON = false }()

	data := json.RawMessage(`[{"id":1}]`)
	pagination := json.RawMessage(`{"current_page_number":1,"total_pages":3}`)

	out := captureStdout(t, func() {
		PrintWithPagination(data, pagination, nil)
	})

	var envelope Envelope
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &envelope); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if envelope.Pagination == nil {
		t.Error("pagination should be present in envelope")
	}
}
