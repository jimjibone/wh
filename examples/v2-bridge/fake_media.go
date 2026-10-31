package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"sync"
	"time"

	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/bridges/services"
	"github.com/jimjibone/wh/v2/types"
)

// FakeSpeaker is a speaker: media_playback with transport commands and album
// art, volume with the group feature, and media_sources holding a list of
// favourites (a list of structs) that the play command takes an id from.
type FakeSpeaker struct {
	dev      *bridges.Device
	playback *services.MediaPlayback
	volume   *services.Volume
	sources  *services.MediaSources

	mu    sync.Mutex
	track int
}

var fakeTracks = []struct{ title, artist, album string }{
	{"Morning Light", "The Examples", "Placeholder"},
	{"Second Song", "The Examples", "Placeholder"},
	{"Banana Boat", "Alice and Bob", "Fruit"},
}

func NewFakeSpeaker(uniqueID, name string) *FakeSpeaker {
	f := &FakeSpeaker{
		dev:      newDevice(uniqueID, name, "Fake Speaker"),
		playback: services.NewMediaPlayback(""),
		volume:   services.NewVolume("", services.VolumeFeatureGroup),
		sources:  services.NewMediaSources(""),
	}
	ep := f.dev.Endpoint("1")
	ep.AddService(f.playback, f.volume, f.sources)

	f.playback.OnPlay(func() error { f.playback.State.Set("playing"); return nil })
	f.playback.OnPause(func() error { f.playback.State.Set("paused"); return nil })
	f.playback.OnStop(func() error { f.playback.State.Set("stopped"); return nil })
	f.playback.OnNext(func() error { f.skip(1); return nil })
	f.playback.OnPrevious(func() error { f.skip(-1); return nil })
	f.playback.Art.OnImageRequest(f.art)

	f.volume.Level.OnWrite(func(v int64) error { f.volume.Level.Set(v); return nil })
	f.volume.Muted.OnWrite(func(v bool) error { f.volume.Muted.Set(v); return nil })
	f.volume.GroupLevel.OnWrite(func(v int64) error { f.volume.GroupLevel.Set(v); return nil })
	f.volume.GroupMuted.OnWrite(func(v bool) error { f.volume.GroupMuted.Set(v); return nil })
	f.volume.OnStep(func(p types.VolumeStepParams) error {
		f.volume.Level.Set(min(100, max(0, f.volume.Level.Get()+p.Delta)))
		return nil
	})

	f.sources.SetSources([]types.MediaSourcesSourcesItem{
		{ID: "fav-1", Name: "Radio Example"},
		{ID: "fav-2", Name: "Morning Playlist"},
		{ID: "fav-3", Name: "Rain Sounds"},
	})
	f.sources.OnPlay(func(p types.MediaSourcesPlayParams) error {
		for _, s := range f.sources.SourcesItems() {
			if s.ID == p.ID {
				f.playback.Title.Set(s.Name)
				f.playback.NowPlaying.Set(s.Name)
				f.playback.State.Set("playing")
				return nil
			}
		}
		return fmt.Errorf("no source %q", p.ID)
	})

	f.volume.Level.Set(25)
	f.volume.Muted.Set(false)
	f.volume.GroupLevel.Set(30)
	f.volume.GroupMuted.Set(false)
	f.playback.State.Set("stopped")
	f.skip(0)
	return f
}

func (f *FakeSpeaker) skip(by int) {
	f.mu.Lock()
	f.track = (f.track + by + len(fakeTracks)) % len(fakeTracks)
	t := fakeTracks[f.track]
	f.mu.Unlock()
	f.playback.Title.Set(t.title)
	f.playback.Artist.Set(t.artist)
	f.playback.Album.Set(t.album)
	f.playback.NowPlaying.Set(t.artist + " - " + t.title)
	f.playback.Duration.Set(3*time.Minute + time.Duration(f.track)*20*time.Second)
	f.playback.Elapsed.Set(0)
}

// art draws a small square in a colour per track.
func (f *FakeSpeaker) art() ([]byte, error) {
	f.mu.Lock()
	n := f.track
	f.mu.Unlock()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	c := color.RGBA{R: uint8(80 * n), G: 120, B: uint8(255 - 80*n), A: 255}
	for y := range 64 {
		for x := range 64 {
			img.SetRGBA(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
