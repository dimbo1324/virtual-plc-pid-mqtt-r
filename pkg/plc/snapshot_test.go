package plc

import "testing"

func TestSnapshotReturnsDeepCopy(t *testing.T) {
	runtime := newTestRuntime(t)
	first := runtime.Snapshot()
	first.Loops["pressure"] = LoopSnapshot{Name: "mutated"}
	delete(first.Loops, "pressure")

	second := runtime.Snapshot()
	if _, ok := second.Loops["pressure"]; !ok {
		t.Fatal("mutating returned snapshot changed internal loop map")
	}
}
