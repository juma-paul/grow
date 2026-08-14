package events

import (
	"encoding/json"
	"testing"
)

func TestMarshalResizeBegin(t *testing.T) {
	data, err := Marshal(ResizeBegin{OldCap: 4, NewCap: 8})
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result["type"] != "resize_begin" {
		t.Errorf("type = %v, want resize_begin", result["type"])
	}

	if result["old_cap"] != float64(4) {
		t.Errorf("old_cap = %v, want 4", result["old_cap"])
	}

	if result["new_cap"] != float64(8) {
		t.Errorf("new_cap = %v, want 8", result["new_cap"])
	}
}


func TestMarshalAppendBegin(t *testing.T) {
	data, err := Marshal(AppendBegin{Value: 42, Length: 8, Capacity: 8})
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result["type"] != "append_begin" {
		t.Errorf("type = %v, want append_begin", result["type"])
	}

	if result["value"] != float64(42) {
		t.Errorf("value = %v, want 42", result["value"])
	}
}