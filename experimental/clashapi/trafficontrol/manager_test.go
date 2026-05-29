package trafficontrol

import (
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	C "github.com/sagernet/sing-box/constant"

	"github.com/gofrs/uuid/v5"
)

func TestSnapshotForState(t *testing.T) {
	manager := NewManager()
	active := newTestTracker("")
	closed := newTestTracker("")
	dns := newTestTracker(C.TypeDNS)

	manager.Join(active)
	manager.Join(closed)
	manager.Join(dns)
	manager.Leave(closed)
	manager.Leave(dns)

	if got := snapshotConnectionCount(t, manager.Snapshot()); got != 1 {
		t.Fatalf("default snapshot connections = %d, want 1 active connection", got)
	}
	if got := snapshotConnectionCount(t, manager.SnapshotForState(ConnectionStateClosed)); got != 1 {
		t.Fatalf("closed snapshot connections = %d, want 1 closed connection", got)
	}
	if got := snapshotConnectionCount(t, manager.SnapshotForState(ConnectionStateAll)); got != 2 {
		t.Fatalf("all snapshot connections = %d, want active + closed connections", got)
	}
}

func TestClosedConnectionJSONIncludesClosedAt(t *testing.T) {
	manager := NewManager()
	active := newTestTracker("")
	closed := newTestTracker("")

	manager.Join(active)
	manager.Join(closed)
	manager.Leave(closed)

	activePayload := decodeSnapshotPayload(t, manager.Snapshot())
	if _, loaded := activePayload.Connections[0]["closedAt"]; loaded {
		t.Fatal("active connection unexpectedly includes closedAt")
	}

	closedPayload := decodeSnapshotPayload(t, manager.SnapshotForState(ConnectionStateClosed))
	closedAt, loaded := closedPayload.Connections[0]["closedAt"]
	if !loaded {
		t.Fatal("closed connection does not include closedAt")
	}
	closedAtValue, ok := closedAt.(string)
	if !ok {
		t.Fatalf("closedAt type = %T, want string", closedAt)
	}
	if _, err := time.Parse(time.RFC3339Nano, closedAtValue); err != nil {
		t.Fatalf("closedAt is not RFC3339 time: %v", err)
	}
}

type snapshotPayload struct {
	Connections []map[string]any `json:"connections"`
}

func snapshotConnectionCount(t *testing.T, snapshot *Snapshot) int {
	t.Helper()
	return len(decodeSnapshotPayload(t, snapshot).Connections)
}

func decodeSnapshotPayload(t *testing.T, snapshot *Snapshot) snapshotPayload {
	t.Helper()
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	var payload snapshotPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	return payload
}

type testTracker struct {
	metadata TrackerMetadata
}

func newTestTracker(outboundType string) *testTracker {
	id, _ := uuid.NewV4()
	return &testTracker{
		metadata: TrackerMetadata{
			ID:           id,
			CreatedAt:    time.Now(),
			Upload:       new(atomic.Int64),
			Download:     new(atomic.Int64),
			OutboundType: outboundType,
		},
	}
}

func (t *testTracker) Metadata() *TrackerMetadata {
	return &t.metadata
}

func (t *testTracker) Close() error {
	return nil
}
