package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"math/rand"
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/bridges/services"
	"github.com/jimjibone/wh/v2/types"
)

// FakeCamera is a camera with snapshots, a motion sensor and object
// detection events, all on endpoint 1.
type FakeCamera struct {
	dev    *bridges.Device
	camera *services.Camera
	motion *services.Motion
}

func NewFakeCamera(uniqueID, name string, sim bool) *FakeCamera {
	f := &FakeCamera{
		dev:    newDevice(uniqueID, name, "Fake Camera"),
		camera: services.NewCamera("", services.CameraFeatureSnapshot),
		motion: services.NewMotion(""),
	}
	f.dev.Endpoint("1").AddService(f.camera, f.motion)
	f.camera.Image.OnImageRequest(f.snapshot)
	f.motion.Motion.Set(false)
	if sim {
		go f.simulate()
	}
	return f
}

// snapshot generates a 1920x1080 JPEG filled with a random solid colour.
func (f *FakeCamera) snapshot() ([]byte, error) {
	r, g, b := uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256))
	const width, height = 1920, 1080
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	c := color.RGBA{R: r, G: g, B: b, A: 255}
	for y := range height {
		for x := range width {
			img.SetRGBA(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// simulate sees motion now and then and detects what moved.
func (f *FakeCamera) simulate() {
	labels := []string{"person", "car", "cat"}
	for {
		time.Sleep(time.Duration(20+rand.Intn(40)) * time.Second)
		f.motion.Motion.Set(true)
		err := f.motion.EmitDetected()
		if err == nil {
			zone := "driveway"
			err = f.camera.EmitDetection(types.CameraDetectionEvent{
				Label: labels[rand.Intn(len(labels))],
				Score: ptr(0.6 + rand.Float64()*0.4),
				Zone:  &zone,
			})
		}
		if err != nil {
			log.Errorf("%s: %s", f.dev.ID(), err)
		}
		time.Sleep(5 * time.Second)
		f.motion.Motion.Set(false)
	}
}
