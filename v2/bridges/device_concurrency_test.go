package bridges

import (
	"sync"
	"testing"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

// Full states, label changes and actions run while attributes push updates
// through the device. Run with -race; a lock-order cycle would hang it.
func TestDeviceConcurrentUse(t *testing.T) {
	dev, outs, ins := twoChannelRelay()
	defer dev.Close()
	sent := make(chan struct{}, 1)
	dev.Init(func(*clientsapi.Device) {
		select {
		case sent <- struct{}{}:
		default:
		}
	})
	for _, out := range outs {
		out.On.OnWrite(func(v bool) error { out.On.Set(v); return nil })
	}

	var wg sync.WaitGroup
	run := func(f func(i int)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 200 {
				f(i)
			}
		}()
	}
	run(func(i int) { outs[0].On.Set(i%2 == 0) })
	run(func(i int) { ins[1].On.Set(i%2 == 0) })
	run(func(int) { dev.SendFullState() })
	run(func(i int) { dev.Endpoint("2").SetLabel("Output 2") })
	run(func(i int) {
		_, _ = dev.HandleAction(&clientsapi.ActionRequest{
			EndpointId: "1", ServiceId: "on_off",
			Action: &clientsapi.ActionRequest_Write{Write: &clientsapi.WriteAction{Values: []*clientsapi.FieldValue{values.Field("on", values.Bool(i%2 == 0))}}},
		}, nil)
	})

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("deadlock: concurrent use did not finish")
	}
	select {
	case <-sent:
	case <-time.After(time.Second):
		t.Fatal("nothing sent")
	}
}
