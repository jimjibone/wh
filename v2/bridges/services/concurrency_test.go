package services

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jimjibone/wh/v2/types"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
)

// Bridges set attributes from many goroutines (pollers, MQTT callbacks) while
// actions and full states read them. Run with -race.
func TestServiceConcurrentUse(t *testing.T) {
	light := NewLight("", LightFeatureBrightness, LightFeatureColorTemp, LightFeatureHS)
	button := NewButton("")
	var mu sync.Mutex
	var updates int
	push := func(*clientsapi.Service) { mu.Lock(); updates++; mu.Unlock() }
	emit := func(*clientsapi.Event) {}

	var writes atomic.Int64
	light.On.OnWrite(func(v bool) error { writes.Add(1); light.On.Set(v); return nil })
	light.OnSet(func(p types.LightSetParams) error {
		if p.Brightness != nil {
			light.Brightness.Set(*p.Brightness)
		}
		return nil
	})

	var wg sync.WaitGroup
	run := func(n int, f func(i int)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range n {
				f(i)
			}
		}()
	}
	const n = 200
	run(1, func(int) { light.Push(push, emit); button.Push(push, emit) })
	run(n, func(i int) { light.Brightness.Set(int64(i % 100)) })
	run(n, func(i int) { light.ColorTemp.Set(int64(2000 + i)) })
	run(n, func(i int) { light.Color.SetHueSat(float64(i%360), 50) })
	run(n, func(i int) { _ = light.Pb(); _ = light.Brightness.Get() })
	run(n, func(i int) { _ = write(light, values.Field("on", values.Bool(i%2 == 0))) })
	run(n, func(i int) { _, _ = invoke(light, "set", values.Field("brightness", values.Int(int64(i%100)))) })
	run(n, func(int) { _ = button.EmitPress() })
	run(n, func(i int) { light.SetAlias("alias"); light.Brightness.SetRange(0, 100) })
	wg.Wait()

	if writes.Load() != n {
		t.Fatalf("%d writes handled, want %d", writes.Load(), n)
	}
}
