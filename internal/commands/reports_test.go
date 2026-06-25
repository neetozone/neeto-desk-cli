package commands

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTicketStatusReport(t *testing.T) {
	data := json.RawMessage(`{"ticket_statuses":[
		{"name":"closed","value":{"present":8,"previous":null,"change_percentage":null}},
		{"name":"open","value":{"present":3,"previous":2,"change_percentage":50}}
	]}`)

	items, headers, rows, ok := ticketStatusReport(data)
	if !ok {
		t.Fatal("ticketStatusReport() ok = false, want true")
	}

	wantHeaders := []string{"NAME", "PRESENT", "PREVIOUS", "CHANGE %"}
	if !reflect.DeepEqual(headers, wantHeaders) {
		t.Errorf("headers = %v, want %v", headers, wantHeaders)
	}

	wantRows := [][]interface{}{
		{"closed", float64(8), nil, nil},
		{"open", float64(3), float64(2), float64(50)},
	}
	if !reflect.DeepEqual(rows, wantRows) {
		t.Errorf("rows = %#v, want %#v", rows, wantRows)
	}

	var arr []map[string]interface{}
	if err := json.Unmarshal(items, &arr); err != nil {
		t.Fatalf("items is not a JSON array: %v", err)
	}
	if len(arr) != 2 {
		t.Errorf("items length = %d, want 2 (raw payload preserved for --json)", len(arr))
	}
}

func TestTicketStatusReport_MissingKey(t *testing.T) {
	data := json.RawMessage(`{"something_else":[]}`)
	if _, _, _, ok := ticketStatusReport(data); ok {
		t.Error("ticketStatusReport() ok = true, want false when ticket_statuses absent")
	}
}

func TestTicketStatusReport_MissingValueObject(t *testing.T) {
	data := json.RawMessage(`{"ticket_statuses":[{"name":"closed"}]}`)

	_, headers, rows, ok := ticketStatusReport(data)
	if !ok {
		t.Fatal("ticketStatusReport() ok = false, want true")
	}
	if !reflect.DeepEqual(headers, []string{"NAME", "PRESENT", "PREVIOUS", "CHANGE %"}) {
		t.Errorf("headers = %v", headers)
	}
	wantRows := [][]interface{}{{"closed", nil, nil, nil}}
	if !reflect.DeepEqual(rows, wantRows) {
		t.Errorf("rows = %#v, want %#v", rows, wantRows)
	}
}

func TestTicketStatusReport_Empty(t *testing.T) {
	if _, _, _, ok := ticketStatusReport(json.RawMessage(`{"ticket_statuses":[]}`)); ok {
		t.Error("ticketStatusReport() ok = true, want false for empty list (fall back to 'No records found')")
	}
}

func TestTicketTimeSeriesReport(t *testing.T) {
	data := json.RawMessage(`{
		"dates":["2026-06-19","2026-06-20"],
		"data":[
			{"name":"new","values":[0,7]},
			{"name":"closed","values":[0,6]}
		]
	}`)

	headers, rows, ok := ticketTimeSeriesReport(data)
	if !ok {
		t.Fatal("ticketTimeSeriesReport() ok = false, want true")
	}

	wantHeaders := []string{"DATE", "NEW", "CLOSED"}
	if !reflect.DeepEqual(headers, wantHeaders) {
		t.Errorf("headers = %v, want %v", headers, wantHeaders)
	}

	wantRows := [][]interface{}{
		{"2026-06-19", float64(0), float64(0)},
		{"2026-06-20", float64(7), float64(6)},
	}
	if !reflect.DeepEqual(rows, wantRows) {
		t.Errorf("rows = %#v, want %#v", rows, wantRows)
	}
}

func TestTicketTimeSeriesReport_ShortSeries(t *testing.T) {
	data := json.RawMessage(`{"dates":["d1","d2","d3"],"data":[{"name":"new","values":[1,2]}]}`)

	headers, rows, ok := ticketTimeSeriesReport(data)
	if !ok {
		t.Fatal("ticketTimeSeriesReport() ok = false, want true")
	}
	if !reflect.DeepEqual(headers, []string{"DATE", "NEW"}) {
		t.Errorf("headers = %v", headers)
	}
	wantRows := [][]interface{}{
		{"d1", float64(1)},
		{"d2", float64(2)},
		{"d3", nil},
	}
	if !reflect.DeepEqual(rows, wantRows) {
		t.Errorf("rows = %#v, want %#v", rows, wantRows)
	}
}

func TestTicketTimeSeriesReport_MissingKeys(t *testing.T) {
	if _, _, ok := ticketTimeSeriesReport(json.RawMessage(`{"foo":1}`)); ok {
		t.Error("ticketTimeSeriesReport() ok = true, want false when dates/data absent")
	}
}

func TestTicketTimeSeriesReport_EmptyDates(t *testing.T) {
	data := json.RawMessage(`{"dates":[],"data":[{"name":"new","values":[]}]}`)
	if _, _, ok := ticketTimeSeriesReport(data); ok {
		t.Error("ticketTimeSeriesReport() ok = true, want false for empty dates")
	}
}
