package reactors

import (
	"sync"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type Device struct {
	mu sync.RWMutex
	id string

	info        *InfoService
	infoHandler func(*InfoService)

	online        *OnlineService
	onlineHandler func(*OnlineService)

	input        *InputService
	inputHandler func(*InputService)

	relay        *RelayService
	relayHandler func(*RelayService)
}

func NewDevice(id string) *Device {
	dev := &Device{
		id: id,
	}
	return dev
}

// Get the ID of the device.
func (dev *Device) ID() string {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.id
}

// Get the info service. Returns nil if the device does not have the info
// service or no updates have been received for the device yet.
func (dev *Device) Info() *InfoService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.info
}

// Set a function to be called when the info service is updated.
func (dev *Device) OnInfoUpdated(handler func(*InfoService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	dev.infoHandler = handler
}

// Get the online service. Returns nil if the device does not have the online
// service or no updates have been received for the device yet.
func (dev *Device) Online() *OnlineService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.online
}

// Set a function to be called when the online service is updated.
func (dev *Device) OnOnlineUpdated(handler func(*OnlineService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	dev.onlineHandler = handler
}

// Get the input service. Returns nil if the device does not have the input
// service or no updates have been received for the device yet.
func (dev *Device) Input() *InputService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.input
}

// Set a function to be called when the input service is updated.
func (dev *Device) OnInputUpdated(handler func(*InputService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	dev.inputHandler = handler
}

// Get the relay service. Returns nil if the device does not have the relay
// service or no updates have been received for the device yet.
func (dev *Device) Relay() *RelayService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.relay
}

// Set a function to be called when the relay service is updated.
func (dev *Device) OnRelayUpdated(handler func(*RelayService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	dev.relayHandler = handler
}

// Internal method to update this reactor device with the new state of the
// device.
func (dev *Device) HandleUpdate(update *clientsapi.Device) {
	for _, service := range update.Services {
		switch service.GetTyp() {
		case clientsapi.Service_INFO:
			dev.mu.Lock()
			if dev.info == nil {
				dev.info = &InfoService{}
			}
			changed := dev.info.handleUpdate(service) && dev.infoHandler != nil
			dev.mu.Unlock()
			if changed {
				dev.infoHandler(dev.info)
			}

		case clientsapi.Service_ONLINE:
			dev.mu.Lock()
			if dev.online == nil {
				dev.online = &OnlineService{}
			}
			changed := dev.online.handleUpdate(service) && dev.onlineHandler != nil
			dev.mu.Unlock()
			if changed {
				dev.onlineHandler(dev.online)
			}

		case clientsapi.Service_INPUT:
			dev.mu.Lock()
			if dev.input == nil {
				dev.input = &InputService{}
			}
			changed := dev.input.handleUpdate(service) && dev.inputHandler != nil
			dev.mu.Unlock()
			if changed {
				dev.inputHandler(dev.input)
			}

		case clientsapi.Service_RELAY:
			dev.mu.Lock()
			if dev.relay == nil {
				dev.relay = &RelayService{}
			}
			changed := dev.relay.handleUpdate(service) && dev.relayHandler != nil
			dev.mu.Unlock()
			if changed {
				dev.relayHandler(dev.relay)
			}
		}
	}
}
