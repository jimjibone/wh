package reactors

import (
	"sync"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type Device struct {
	mu sync.RWMutex
	id string

	battery        *BatteryService
	batteryHandler func(*BatteryService)

	button        *ButtonService
	buttonHandler func(*ButtonService)

	camera        *CameraService
	cameraHandler func(*CameraService)

	climate        *ClimateService
	climateHandler func(*ClimateService)

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

	lightbulb        *LightbulbService
	lightbulbHandler func(*LightbulbService)

	online        *OnlineService
	onlineHandler func(*OnlineService)

	relay        *RelayService
	relayHandler func(*RelayService)

	update        *UpdateService
	updateHandler func(*UpdateService)
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
// Get the climate service. Returns nil if the device does not have the climate
// service or no updates have been received for the device yet.
func (dev *Device) Climate() *ClimateService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.climate
}

// Set a function to be called when the climate service is updated.
func (dev *Device) OnClimateUpdated(handler func(*ClimateService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	dev.climateHandler = handler
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

// Get the lightbulb service. Returns nil if the device does not have the lightbulb
// service or no updates have been received for the device yet.
func (dev *Device) Lightbulb() *LightbulbService {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return dev.lightbulb
}

// Set a function to be called when the lightbulb service is updated.
func (dev *Device) OnLightbulbUpdated(handler func(*LightbulbService)) {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	dev.lightbulbHandler = handler
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
				dev.climate = &ClimateService{}
			}
			changed := dev.climate.handleUpdate(service) && dev.climateHandler != nil
			dev.mu.Unlock()
			if changed {
				dev.climateHandler(dev.climate)
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
				dev.lightbulb = &LightbulbService{}
			}
			changed := dev.lightbulb.handleUpdate(service) && dev.lightbulbHandler != nil
			dev.mu.Unlock()
			if changed {
				dev.lightbulbHandler(dev.lightbulb)
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
				dev.relay = &RelayService{}
			}
			changed := dev.relay.handleUpdate(service) && dev.relayHandler != nil
			dev.mu.Unlock()
			if changed {
				dev.relayHandler(dev.relay)
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
}
