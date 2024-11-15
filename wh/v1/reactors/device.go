package reactors

import (
	"context"
	"sync"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type Requester func(ctx context.Context, req *clientsapi.ActionRequest, handler func(resp *clientsapi.ActionResponse)) error

type Device struct {
	mu        sync.RWMutex
	requester Requester
	id        string
	wait      chan struct{}
	waitDone  bool

	battery        *BatteryService
	batteryHandler func(*BatteryService)

	button        *ButtonService
	buttonHandler func(*ButtonService)

	camera        *CameraService
	cameraHandler func(*CameraService)

	climate        map[string]*ClimateService
	climateHandler map[string]func(*ClimateService)

	contact        *ContactService
	contactHandler func(*ContactService)

	enum        *EnumService
	enumHandler func(*EnumService)

	environment        *EnvironmentService
	environmentHandler func(*EnvironmentService)

	info        *InfoService
	infoHandler func(*InfoService)

	input        *InputService
	inputHandler func(*InputService)

	lightbulb        map[string]*LightbulbService
	lightbulbHandler map[string]func(*LightbulbService)

	online        *OnlineService
	onlineHandler func(*OnlineService)

	relay        map[string]*RelayService
	relayHandler map[string]func(*RelayService)

	update        *UpdateService
	updateHandler func(*UpdateService)
}

// Create a new reactor device. You must add this to the client with client.AddReactor() for it to function.
// Alternatively, use client.NewReactor().
func NewDevice(id string) *Device {
	dev := &Device{
		id:   id,
		wait: make(chan struct{}),
	}
	return dev
}

// Internal use only. Initialises the device.
func (dev *Device) Init(requester Requester) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	dev.requester = requester
}

// Get the ID of the device.
func (dev *Device) ID() string {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.id
}

// Returns a channel which waits for the device to come online before it is closed.
func (dev *Device) WaitForDevice() <-chan struct{} {
	return dev.wait
}

// Get the battery service. Returns nil if the device does not have the battery
// service or no updates have been received for the device yet.
func (dev *Device) Battery() *BatteryService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.battery
}

// Set a function to be called when the battery service is updated.
func (dev *Device) OnBatteryUpdated(handler func(*BatteryService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	dev.batteryHandler = handler
}

// Get the button service. Returns nil if the device does not have the button
// service or no updates have been received for the device yet.
func (dev *Device) Button() *ButtonService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.button
}

// Set a function to be called when the button service is updated.
func (dev *Device) OnButtonUpdated(handler func(*ButtonService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	dev.buttonHandler = handler
}

// Get the camera service. Returns nil if the device does not have the camera
// service or no updates have been received for the device yet.
func (dev *Device) Camera() *CameraService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.camera
}

// Set a function to be called when the camera service is updated.
func (dev *Device) OnCameraUpdated(handler func(*CameraService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	dev.cameraHandler = handler
}

// Get the climate service by ID. Returns nil if the device does not have the
// climate service that ID or no updates have been received for the device yet.
func (dev *Device) Climate(id string) *ClimateService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	if dev.climate != nil {
		return dev.climate[id]
	}
	return nil
}

// Set a function to be called when the climate service is updated.
func (dev *Device) OnClimateUpdated(id string, handler func(*ClimateService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	if dev.climateHandler == nil {
		dev.climateHandler = make(map[string]func(*ClimateService))
	}
	dev.climateHandler[id] = handler
}

// Get the contact service. Returns nil if the device does not have the contact
// service or no updates have been received for the device yet.
func (dev *Device) Contact() *ContactService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.contact
}

// Set a function to be called when the contact service is updated.
func (dev *Device) OnContactUpdated(handler func(*ContactService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	dev.contactHandler = handler
}

// Get the enum service. Returns nil if the device does not have the enum
// service or no updates have been received for the device yet.
func (dev *Device) Enum() *EnumService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.enum
}

// Set a function to be called when the enum service is updated.
func (dev *Device) OnEnumUpdated(handler func(*EnumService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	dev.enumHandler = handler
}

// Get the environment service. Returns nil if the device does not have the environment
// service or no updates have been received for the device yet.
func (dev *Device) Environment() *EnvironmentService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.environment
}

// Set a function to be called when the environment service is updated.
func (dev *Device) OnEnvironmentUpdated(handler func(*EnvironmentService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	dev.environmentHandler = handler
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

// Get the lightbulb service by ID. Returns nil if the device does not have the
// lightbulb service that ID or no updates have been received for the device yet.
func (dev *Device) Lightbulb(id string) *LightbulbService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	if dev.lightbulb != nil {
		return dev.lightbulb[id]
	}
	return nil
}

// Set a function to be called when the lightbulb service is updated.
func (dev *Device) OnLightbulbUpdated(id string, handler func(*LightbulbService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	if dev.lightbulbHandler == nil {
		dev.lightbulbHandler = make(map[string]func(*LightbulbService))
	}
	dev.lightbulbHandler[id] = handler
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

// Get the relay service by ID. Returns nil if the device does not have the
// relay service that ID or no updates have been received for the device yet.
func (dev *Device) Relay(id string) *RelayService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	if dev.relay != nil {
		return dev.relay[id]
	}
	return nil
}

// Set a function to be called when the relay service is updated.
func (dev *Device) OnRelayUpdated(id string, handler func(*RelayService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	if dev.relayHandler == nil {
		dev.relayHandler = make(map[string]func(*RelayService))
	}
	dev.relayHandler[id] = handler
}

// Get the update service. Returns nil if the device does not have the update
// service or no updates have been received for the device yet.
func (dev *Device) Update() *UpdateService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.update
}

// Set a function to be called when the update service is updated.
func (dev *Device) OnUpdateUpdated(handler func(*UpdateService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	dev.updateHandler = handler
}

// Internal method to update this reactor device with the new state of the
// device.
func (dev *Device) HandleUpdate(update *clientsapi.Device) {
	for _, service := range update.Services {
		id := service.GetId()
		switch service.GetTyp() {
		case clientsapi.Service_BATTERY:
			dev.mu.Lock()
			if dev.battery == nil {
				dev.battery = &BatteryService{}
			}
			changed := dev.battery.handleUpdate(service) && dev.batteryHandler != nil
			dev.mu.Unlock()
			if changed {
				dev.batteryHandler(dev.battery)
			}

		case clientsapi.Service_BUTTON:
			dev.mu.Lock()
			if dev.button == nil {
				dev.button = &ButtonService{}
			}
			changed := dev.button.handleUpdate(service) && dev.buttonHandler != nil
			dev.mu.Unlock()
			if changed {
				dev.buttonHandler(dev.button)
			}

		case clientsapi.Service_CAMERA:
			dev.mu.Lock()
			if dev.camera == nil {
				dev.camera = &CameraService{}
			}
			changed := dev.camera.handleUpdate(service) && dev.cameraHandler != nil
			dev.mu.Unlock()
			if changed {
				dev.cameraHandler(dev.camera)
			}

		case clientsapi.Service_CLIMATE:
			dev.mu.Lock()
			if dev.climate == nil {
				dev.climate = make(map[string]*ClimateService)
			}
			if dev.climate[id] == nil {
				dev.climate[id] = &ClimateService{requester: func(ctx context.Context, req *clientsapi.ActionRequest, handler func(resp *clientsapi.ActionResponse)) error {
					req.DeviceId = dev.id
					return dev.requester(ctx, req, handler)
				}}
			}
			changed := dev.climate[id].handleUpdate(service)
			dev.mu.Unlock()
			if changed && dev.climateHandler[id] != nil {
				dev.climateHandler[id](dev.climate[id])
			}

		case clientsapi.Service_CONTACT:
			dev.mu.Lock()
			if dev.contact == nil {
				dev.contact = &ContactService{}
			}
			changed := dev.contact.handleUpdate(service) && dev.contactHandler != nil
			dev.mu.Unlock()
			if changed {
				dev.contactHandler(dev.contact)
			}

		case clientsapi.Service_ENUM:
			dev.mu.Lock()
			if dev.enum == nil {
				dev.enum = &EnumService{}
			}
			changed := dev.enum.handleUpdate(service) && dev.enumHandler != nil
			dev.mu.Unlock()
			if changed {
				dev.enumHandler(dev.enum)
			}

		case clientsapi.Service_ENVIRONMENT:
			dev.mu.Lock()
			if dev.environment == nil {
				dev.environment = &EnvironmentService{}
			}
			changed := dev.environment.handleUpdate(service) && dev.environmentHandler != nil
			dev.mu.Unlock()
			if changed {
				dev.environmentHandler(dev.environment)
			}

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

		case clientsapi.Service_LIGHTBULB:
			dev.mu.Lock()
			if dev.lightbulb == nil {
				dev.lightbulb = make(map[string]*LightbulbService)
			}
			if dev.lightbulb[id] == nil {
				dev.lightbulb[id] = &LightbulbService{requester: func(ctx context.Context, req *clientsapi.ActionRequest, handler func(resp *clientsapi.ActionResponse)) error {
					req.DeviceId = dev.id
					return dev.requester(ctx, req, handler)
				}}
			}
			changed := dev.lightbulb[id].handleUpdate(service)
			dev.mu.Unlock()
			if changed && dev.lightbulbHandler[id] != nil {
				dev.lightbulbHandler[id](dev.lightbulb[id])
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

		case clientsapi.Service_RELAY:
			dev.mu.Lock()
			if dev.relay == nil {
				dev.relay = make(map[string]*RelayService)
			}
			if dev.relay[id] == nil {
				dev.relay[id] = &RelayService{requester: func(ctx context.Context, req *clientsapi.ActionRequest, handler func(resp *clientsapi.ActionResponse)) error {
					req.DeviceId = dev.id
					return dev.requester(ctx, req, handler)
				}}
			}
			changed := dev.relay[id].handleUpdate(service)
			dev.mu.Unlock()
			if changed && dev.relayHandler[id] != nil {
				dev.relayHandler[id](dev.relay[id])
			}

		case clientsapi.Service_UPDATE:
			dev.mu.Lock()
			if dev.update == nil {
				dev.update = &UpdateService{}
			}
			changed := dev.update.handleUpdate(service) && dev.updateHandler != nil
			dev.mu.Unlock()
			if changed {
				dev.updateHandler(dev.update)
			}
		}
	}

	if !dev.waitDone {
		dev.waitDone = true
		close(dev.wait)
	}
}
