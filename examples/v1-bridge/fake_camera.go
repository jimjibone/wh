package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"math/rand"
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v1/devices"
	"github.com/jimjibone/wh/v1/devices/services"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

type FakeCamera struct {
	dev    *devices.Device
	info   *services.Info
	online *services.Online
	camera *services.Camera
}

func NewFakeCamera(id, name string) *FakeCamera {
	dev := &FakeCamera{
		dev:    devices.NewDevice(id, clientsapi.Device_DEVICE),
		info:   services.NewInfo(),
		online: services.NewOnline(),
		camera: services.NewCamera(""),
	}
	dev.dev.AddService(dev.info, dev.online, dev.camera)

	dev.info.Name.Set(name)
	dev.info.Model.Set("Fake Camera")
	dev.info.Manufacturer.Set("Fake Things Inc")
	dev.online.Online.Set(true)
	dev.online.LastSeen.Set(time.Now())

	dev.camera.Image.OnImageRequest(dev.handleImageRequest)

	return dev
}

// handleImageRequest generates a 1920×1080 JPEG filled with a random solid colour.
func (dev *FakeCamera) handleImageRequest() ([]byte, error) {
	log.Infof("fake camera: generating image")

	r := uint8(rand.Intn(256))
	g := uint8(rand.Intn(256))
	b := uint8(rand.Intn(256))

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

	log.Infof("fake camera: generated %d byte image (r=%d g=%d b=%d)", buf.Len(), r, g, b)
	return buf.Bytes(), nil
}
