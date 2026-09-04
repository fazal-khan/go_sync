package filter

import (
	"testing"
)

func TestApply_CopyValue(t *testing.T) {
	row := map[string]any{"id": "42", "name": "test"}
	filter := Filter{Mutate: Mutate{
		CopyValue: []CopyValue{{From: "id", To: "_id"}},
	}}
	result := Apply(row, filter)
	if result["_id"] != "42" {
		t.Errorf("copy-value failed: _id = %v, want 42", result["_id"])
	}
	if result["id"] != "42" {
		t.Errorf("original field should remain: id = %v", result["id"])
	}
}

func TestApply_CopyMissingFrom(t *testing.T) {
	row := map[string]any{"other": "1"}
	filter := Filter{Mutate: Mutate{
		CopyValue: []CopyValue{{From: "missing", To: "target"}},
	}}
	result := Apply(row, filter)
	if _, exists := result["target"]; exists {
		t.Error("copy from missing field should be a no-op")
	}
}

func TestApply_RemoveFields(t *testing.T) {
	row := map[string]any{"id": "42", "name": "test", "keep": "yes"}
	filter := Filter{Mutate: Mutate{
		RemoveFields: RemoveFields{Field: []string{"name"}},
	}}
	result := Apply(row, filter)
	if _, exists := result["name"]; exists {
		t.Error("remove-field failed: name still present")
	}
	if result["keep"] != "yes" {
		t.Errorf("unrelated field removed: keep = %v", result["keep"])
	}
}

func TestApply_RemoveFields_IgnoreCase(t *testing.T) {
	row := map[string]any{"NAME": "value", "keep": "yes"}
	filter := Filter{Mutate: Mutate{
		RemoveFields: RemoveFields{
			Field:      []string{"name"},
			IgnoreCase: "true",
		},
	}}
	result := Apply(row, filter)
	if _, exists := result["NAME"]; exists {
		t.Error("ignore-case remove failed: NAME still present")
	}
	if result["keep"] != "yes" {
		t.Errorf("unrelated field removed: keep = %v", result["keep"])
	}
}

func TestApply_AddFields(t *testing.T) {
	row := map[string]any{"id": "42"}
	filter := Filter{Mutate: Mutate{
		AddFields: AddFields{Field: []Field{{Key: "type", Value: "A"}}},
	}}
	result := Apply(row, filter)
	if result["type"] != "A" {
		t.Errorf("add-field failed: type = %v, want A", result["type"])
	}
}

func TestApply_LowercaseFields(t *testing.T) {
	row := map[string]any{"name": "HeLLo"}
	filter := Filter{Mutate: Mutate{
		LowercaseFields: LowercaseFields{Field: []string{"name"}},
	}}
	result := Apply(row, filter)
	if result["name"] != "hello" {
		t.Errorf("lowercase failed: name = %v, want hello", result["name"])
	}
}

func TestApply_LowercaseFields_NonString(t *testing.T) {
	row := map[string]any{"name": 42}
	filter := Filter{Mutate: Mutate{
		LowercaseFields: LowercaseFields{Field: []string{"name"}},
	}}
	result := Apply(row, filter)
	if result["name"] != 42 {
		t.Errorf("non-string value should be unchanged: name = %v", result["name"])
	}
}

func TestApply_AllMutations(t *testing.T) {
	row := map[string]any{"id": "42", "name": "MiXeD", "keep": "yes"}
	filter := Filter{Mutate: Mutate{
		CopyValue: []CopyValue{{From: "id", To: "_id"}},
		RemoveFields: RemoveFields{
			Field:      []string{"name"},
			IgnoreCase: "",
		},
		AddFields: AddFields{
			Field: []Field{{Key: "type", Value: "A"}},
		},
		LowercaseFields: LowercaseFields{Field: []string{"name"}},
	}}
	result := Apply(row, filter)
	if result["_id"] != "42" {
		t.Errorf("copy-value failed: _id = %v", result["_id"])
	}
	if _, exists := result["name"]; exists {
		t.Error("remove-field failed: name still present")
	}
	if result["type"] != "A" {
		t.Errorf("add-field failed: type = %v, want A", result["type"])
	}
	// name was removed, so lowercase shouldn't apply
	if _, exists := result["name"]; exists {
		t.Error("name should have been removed before lowercase")
	}
}

func TestApply_OriginalRowUnchanged(t *testing.T) {
	row := map[string]any{"id": "42", "name": "test"}
	filter := Filter{Mutate: Mutate{
		CopyValue: []CopyValue{{From: "id", To: "_id"}},
	}}
	Apply(row, filter)
	if len(row) != 2 {
		t.Errorf("original row should not be modified, got %d keys", len(row))
	}
}

func TestApply_EmptyFilter(t *testing.T) {
	row := map[string]any{"id": "42"}
	filter := Filter{}
	result := Apply(row, filter)
	if len(result) != 1 || result["id"] != "42" {
		t.Errorf("empty filter should return unchanged row")
	}
}
