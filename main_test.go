package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCycleVisitsEveryModeAndReturns(t *testing.T) {
	mode := modeAll
	var seen []string
	for i := 0; i < 3; i++ {
		mode = nextMode[mode]
		seen = append(seen, mode)
	}
	want := []string{modeCurrent, modeMachine, modeAll}
	if !reflect.DeepEqual(seen, want) {
		t.Fatalf("cycle = %v, want %v", seen, want)
	}
}

func TestWorkspaceIDs(t *testing.T) {
	var resp map[string]any
	_ = json.Unmarshal([]byte(`{"result":{"workspaces":[{"workspace_id":"w1"},{"label":"x"},{"workspace_id":"w7"}]}}`), &resp)
	if got := workspaceIDs(resp); !reflect.DeepEqual(got, []string{"w1", "w7"}) {
		t.Fatalf("got %v", got)
	}
	if got := workspaceIDs(map[string]any{}); got == nil || len(got) != 0 {
		t.Fatalf("empty response should give non-nil empty slice, got %#v", got)
	}
}
