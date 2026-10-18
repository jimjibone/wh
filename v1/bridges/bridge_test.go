package bridges

import (
	"slices"
	"testing"
	"time"

	"github.com/jimjibone/wh/v1/clients"
	"github.com/jimjibone/wh/v1/shared/stores"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

func newTestBridge() *Bridge {
	return NewBridge(BridgeConfig{
		ClientConfig: clients.ClientConfig{
			Store:      stores.NewMemStore(),
			ServerAddr: "127.0.0.1:1",
			ClientID:   "client-1",
		},
	})
}

// collect returns the queued status updates, stopping once none arrive for
// quiet.
func collect(b *Bridge, quiet time.Duration) []*clientsapi.StatusUpdate {
	var updates []*clientsapi.StatusUpdate
	for {
		select {
		case update := <-b.updates.Pop():
			updates = append(updates, update)
		case <-time.After(quiet):
			return updates
		}
	}
}

func fullStateIDs(updates []*clientsapi.StatusUpdate) []string {
	var ids []string
	for _, update := range updates {
		for _, dev := range update.DeviceInfo {
			if dev.FullState {
				ids = append(ids, dev.Id)
			}
		}
	}
	slices.Sort(ids)
	return ids
}

func removedIDs(updates []*clientsapi.StatusUpdate) []string {
	var ids []string
	for _, update := range updates {
		ids = append(ids, update.RemovedDeviceIds...)
	}
	return ids
}

func TestRemoveDevice(t *testing.T) {
	b := newTestBridge()
	b.updates.Discard(false) // connected

	dev := NewDevice("dev-a", clientsapi.Device_DEVICE)
	if err := b.AddDevice(dev); err != nil {
		t.Fatalf("AddDevice: %s", err)
	}
	if got := fullStateIDs(collect(b, 200*time.Millisecond)); !slices.Equal(got, []string{"dev-a"}) {
		t.Fatalf("full states after add = %v, want [dev-a]", got)
	}

	// A pending update may or may not be sent, but nothing may follow the
	// removal.
	dev.Info.Name.Set("banana")
	if err := b.RemoveDevice("dev-a"); err != nil {
		t.Fatalf("RemoveDevice: %s", err)
	}
	updates := collect(b, 200*time.Millisecond)
	if len(updates) == 0 {
		t.Fatalf("no updates after removal")
	}
	last := updates[len(updates)-1]
	if !slices.Equal(last.RemovedDeviceIds, []string{"dev-a"}) || len(last.DeviceInfo) != 0 {
		t.Fatalf("last update = %v, want removal of dev-a only", last)
	}

	// Updates to the removed device are dropped without blocking.
	done := make(chan struct{})
	go func() {
		defer close(done)
		dev.Info.Name.Set("apple")
		dev.Info.Name.Set("cherry")
		dev.SendFullState()
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("update to removed device blocked")
	}
	if updates := collect(b, 200*time.Millisecond); len(updates) != 0 {
		t.Fatalf("updates after removal = %v, want none", updates)
	}

	if err := b.RemoveDevice("dev-a"); err == nil {
		t.Errorf("second RemoveDevice of dev-a returned nil error")
	}
	if err := b.RemoveDevice("dev-unknown"); err == nil {
		t.Errorf("RemoveDevice of unknown device returned nil error")
	}
}

func TestRemoveDeviceResentOnConnect(t *testing.T) {
	b := newTestBridge() // disconnected, updates discarded

	devA := NewDevice("dev-a", clientsapi.Device_DEVICE)
	defer devA.Close()
	if err := b.AddDevice(devA); err != nil {
		t.Fatalf("AddDevice: %s", err)
	}
	if err := b.AddDevice(NewDevice("dev-b", clientsapi.Device_DEVICE)); err != nil {
		t.Fatalf("AddDevice: %s", err)
	}
	if err := b.RemoveDevice("dev-b"); err != nil {
		t.Fatalf("RemoveDevice: %s", err)
	}

	// Connect: the removal made while disconnected is sent.
	b.updates.Discard(false)
	b.queueFullState()
	updates := collect(b, 200*time.Millisecond)
	if got := removedIDs(updates); !slices.Equal(got, []string{"dev-b"}) {
		t.Errorf("removed on connect = %v, want [dev-b]", got)
	}
	if got := fullStateIDs(updates); !slices.Equal(got, []string{"dev-a"}) {
		t.Errorf("full states on connect = %v, want [dev-a]", got)
	}

	// Reconnect: still resent, as the server may not have processed it.
	b.queueFullState()
	if got := removedIDs(collect(b, 200*time.Millisecond)); !slices.Equal(got, []string{"dev-b"}) {
		t.Errorf("removed on reconnect = %v, want [dev-b]", got)
	}

	// Re-adding the device stops the removal being resent.
	devB := NewDevice("dev-b", clientsapi.Device_DEVICE)
	defer devB.Close()
	if err := b.AddDevice(devB); err != nil {
		t.Fatalf("AddDevice: %s", err)
	}
	collect(b, 200*time.Millisecond)
	b.queueFullState()
	updates = collect(b, 200*time.Millisecond)
	if got := removedIDs(updates); len(got) != 0 {
		t.Errorf("removed after re-add = %v, want none", got)
	}
	if got := fullStateIDs(updates); !slices.Equal(got, []string{"dev-a", "dev-b"}) {
		t.Errorf("full states after re-add = %v, want [dev-a dev-b]", got)
	}
}
