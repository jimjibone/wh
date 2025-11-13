package reactors

import (
	"context"
	"sync"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
	"github.com/jimjibone/woodhouse-4/log"
	"github.com/jimjibone/woodhouse-4/shared/stores"
	"github.com/jimjibone/woodhouse-4/wh/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	log    *log.Context
	client *wh.Client

	mu         sync.RWMutex
	reactables map[string]*reactable
}

// Collection of things which react to updates from a particular device ID.
type reactable struct {
	devices  map[*Device]struct{}
	services map[Service]reactableService
}

type reactableService struct {
	typ clientsapi.Service_ServiceType
	id  string
}

func newReactable() *reactable {
	return &reactable{
		devices:  make(map[*Device]struct{}),
		services: make(map[Service]reactableService),
	}
}

func (r *reactable) addDevice(device *Device) {
	r.devices[device] = struct{}{}
}

func (r *reactable) addService(service Service, typ clientsapi.Service_ServiceType, id string) {
	r.services[service] = reactableService{
		typ: typ,
		id:  id,
	}
}

func NewClient(store stores.Store, serverAddr string, opts ...wh.ClientOption) *Client {
	rc := &Client{
		log:        log.NewContext(log.DefaultLogger, "reactor", log.DebugLevel),
		reactables: make(map[string]*reactable),
	}
	opts = append(opts, wh.WithConnectionHandler(rc.runloop))
	rc.client = wh.NewClient(store, serverAddr, opts...)
	return rc
}

func (rc *Client) Client() *wh.Client {
	return rc.client
}

func (rc *Client) GetDevice(deviceID string) *Device {
	if deviceID == "" {
		panic("device id must be defined")
	}
	rc.mu.Lock()
	defer rc.mu.Unlock()
	reactable, found := rc.reactables[deviceID]
	if !found {
		reactable = newReactable()
		rc.reactables[deviceID] = reactable
	}
	device := newDevice(deviceID, func(ctx context.Context, req *clientsapi.ActionRequest, handler func(resp *clientsapi.ActionResponse)) error {
		req.DeviceId = deviceID
		return rc.client.RequestAction(ctx, req, handler)
	})
	reactable.addDevice(device)
	return device
}

func (rc *Client) addService(deviceID string, serviceIDs []string, defaultID string, typ clientsapi.Service_ServiceType, service Service) {
	if deviceID == "" {
		panic("device id must be defined")
	}
	if len(serviceIDs) > 1 {
		panic("only one serviceID argument is allowed")
	}
	serviceID := defaultID
	if len(serviceIDs) == 1 {
		serviceID = serviceIDs[0]
	}
	if serviceID == "" {
		panic("service id must be defined")
	}
	reactable, found := rc.reactables[deviceID]
	if !found {
		reactable = newReactable()
		rc.reactables[deviceID] = reactable
	}
	service.init(serviceID, func(ctx context.Context, req *clientsapi.ActionRequest, handler func(resp *clientsapi.ActionResponse)) error {
		req.DeviceId = deviceID
		return rc.client.RequestAction(ctx, req, handler)
	})
	reactable.addService(service, typ, serviceID)
}

// Get a battery service reactor for the specified device ID. If serviceID is not defined the default of "battery" will be used.
func (rc *Client) GetBattery(deviceID string, serviceID ...string) *BatteryService {
	service := &BatteryService{}
	rc.addService(deviceID, serviceID, "battery", clientsapi.Service_BATTERY, service)
	return service
}

// Get a button service reactor for the specified device ID. If serviceID is not defined the default of "button" will be used.
func (rc *Client) GetButton(deviceID string, serviceID ...string) *ButtonService {
	service := &ButtonService{}
	rc.addService(deviceID, serviceID, "button", clientsapi.Service_BATTERY, service)
	return service
}

// Get a camera service reactor for the specified device ID. If serviceID is not defined the default of "camera" will be used.
func (rc *Client) GetCamera(deviceID string, serviceID ...string) *CameraService {
	service := &CameraService{}
	rc.addService(deviceID, serviceID, "camera", clientsapi.Service_BATTERY, service)
	return service
}

// Get a cliamte service reactor for the specified device ID. If serviceID is not defined the default of "cliamte" will be used.
func (rc *Client) GetClimate(deviceID string, serviceID ...string) *ClimateService {
	service := &ClimateService{}
	rc.addService(deviceID, serviceID, "climate", clientsapi.Service_BATTERY, service)
	return service
}

// Get a contact service reactor for the specified device ID. If serviceID is not defined the default of "contact" will be used.
func (rc *Client) GetContact(deviceID string, serviceID ...string) *ContactService {
	service := &ContactService{}
	rc.addService(deviceID, serviceID, "contact", clientsapi.Service_BATTERY, service)
	return service
}

// Get a enum service reactor for the specified device ID. If serviceID is not defined the default of "enum" will be used.
func (rc *Client) GetEnum(deviceID string, serviceID ...string) *EnumService {
	service := &EnumService{}
	rc.addService(deviceID, serviceID, "enum", clientsapi.Service_BATTERY, service)
	return service
}

// Get a environment service reactor for the specified device ID. If serviceID is not defined the default of "environment" will be used.
func (rc *Client) GetEnvironment(deviceID string, serviceID ...string) *EnvironmentService {
	service := &EnvironmentService{}
	rc.addService(deviceID, serviceID, "environment", clientsapi.Service_BATTERY, service)
	return service
}

// Get a generic service reactor for the specified device ID. If serviceID is not defined the default of "generic" will be used.
func (rc *Client) GetGeneric(deviceID string, serviceID ...string) *GenericService {
	service := &GenericService{}
	rc.addService(deviceID, serviceID, "generic", clientsapi.Service_BATTERY, service)
	return service
}

// Get a info service reactor for the specified device ID. If serviceID is not defined the default of "info" will be used.
func (rc *Client) GetInfo(deviceID string, serviceID ...string) *InfoService {
	service := &InfoService{}
	rc.addService(deviceID, serviceID, "info", clientsapi.Service_BATTERY, service)
	return service
}

// Get a input service reactor for the specified device ID. If serviceID is not defined the default of "input" will be used.
func (rc *Client) GetInput(deviceID string, serviceID ...string) *InputService {
	service := &InputService{}
	rc.addService(deviceID, serviceID, "input", clientsapi.Service_BATTERY, service)
	return service
}

// Get a lightbulb service reactor for the specified device ID. If serviceID is not defined the default of "lightbulb" will be used.
func (rc *Client) GetLightbulb(deviceID string, serviceID ...string) *LightbulbService {
	service := &LightbulbService{}
	rc.addService(deviceID, serviceID, "lightbulb", clientsapi.Service_LIGHTBULB, service)
	return service
}

// Get a online service reactor for the specified device ID. If serviceID is not defined the default of "online" will be used.
func (rc *Client) GetOnline(deviceID string, serviceID ...string) *OnlineService {
	service := &OnlineService{}
	rc.addService(deviceID, serviceID, "online", clientsapi.Service_BATTERY, service)
	return service
}

// Get a relay service reactor for the specified device ID. If serviceID is not defined the default of "relay" will be used.
func (rc *Client) GetRelay(deviceID string, serviceID ...string) *RelayService {
	service := &RelayService{}
	rc.addService(deviceID, serviceID, "relay", clientsapi.Service_BATTERY, service)
	return service
}

// Get a update service reactor for the specified device ID. If serviceID is not defined the default of "update" will be used.
func (rc *Client) GetUpdate(deviceID string, serviceID ...string) *UpdateService {
	service := &UpdateService{}
	rc.addService(deviceID, serviceID, "update", clientsapi.Service_BATTERY, service)
	return service
}

func (rc *Client) Run() error {
	return rc.client.Run()
}

func (rc *Client) runloop(ctx context.Context, conn *grpc.ClientConn) {
	rc.log.Infof("stream started")
	defer rc.log.Infof("stream finished")

	service := clientsapi.NewClientServiceClient(conn)
	stream, err := service.DeviceStream(ctx, &clientsapi.DeviceStreamRequest{})
	if err != nil {
		rc.log.Errorf("failed to start reactor stream: %s", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		update, err := stream.Recv()
		if err != nil {
			code := status.Code(err)
			if code == codes.Unavailable || code == codes.Canceled {
				rc.log.Debugf("stream closed: %s", err)
			} else {
				rc.log.Errorf("failed to recv reactor request: %s", err)
			}
			return
		} else {
			// rc.log.Debugf("received reactor update: %s", update)

			// Find the reactors for this device update and let them handle it.
			rc.mu.RLock()
			if reactable, found := rc.reactables[update.GetId()]; found {
				for device := range reactable.devices {
					if device != nil {
						device.HandleUpdate(update)
					}
				}
				for _, update := range update.Services {
					for service, info := range reactable.services {
						if service != nil {
							if update.GetTyp() == info.typ && update.GetId() == info.id {
								service.handleUpdate(update)
							} else if update.GetTyp() == clientsapi.Service_INFO {
								service.handleInfo(update)
							} else if update.GetTyp() == clientsapi.Service_ONLINE {
								service.handleOnline(update)
							}
						}
					}
				}
			}
			rc.mu.RUnlock()
		}
	}
}
