package bridges

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/jimjibone/log"
	"github.com/jimjibone/queue/v2"
	"github.com/jimjibone/wh/v1/apitools"
	"github.com/jimjibone/wh/v1/clients"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Bridge struct {
	log    *log.Context
	client *clients.Client
	config BridgeConfig

	devicesMu sync.RWMutex        // locks the devices and removed maps only
	devices   map[string]*Device  // key=id
	removed   map[string]struct{} // key=id, resent on each connect until re-added
	updates   *queue.Queue[*clientsapi.StatusUpdate]
}

type BridgeConfig struct {
	clients.ClientConfig

	// Enables the image stream. Set to true if this bridge provides devices
	// with the Camera service.
	ImagesEnabled bool
}

func NewBridge(config BridgeConfig, opts ...clients.ClientOption) *Bridge {
	bridge := &Bridge{
		log:    log.NewContext(log.DefaultLogger, "bridge", log.DebugLevel),
		config: config,

		devices: make(map[string]*Device),
		removed: make(map[string]struct{}),
		updates: queue.New[*clientsapi.StatusUpdate](),
	}

	opts = append(opts,
		clients.WithConnectionHandler(bridge.deviceFeedback),
		clients.WithConnectionHandler(bridge.deviceControl),
	)
	if config.ImagesEnabled {
		opts = append(opts, clients.WithConnectionHandler(bridge.imagesControl))
	}
	bridge.client = clients.NewClient(config.ClientConfig, opts...)

	// Discard updates until we're connected to the server.
	bridge.updates.Discard(true)

	return bridge
}

func (bridge *Bridge) Client() *clients.Client {
	return bridge.client
}

// Add a device to the bridge.
func (bridge *Bridge) AddDevice(device *Device) error {
	bridge.devicesMu.Lock()
	defer bridge.devicesMu.Unlock()
	if _, found := bridge.devices[device.ID()]; found {
		return fmt.Errorf("device id already exists in bridge")
	}
	bridge.devices[device.ID()] = device
	delete(bridge.removed, device.ID())
	device.Init(func(state *clientsapi.Device) {
		bridge.updates.Push(&clientsapi.StatusUpdate{
			DeviceInfo: []*clientsapi.Device{state},
		})
	})
	device.SendFullState()
	return nil
}

// Remove a device from the bridge and tell the server to remove it, even if
// online. The device is closed and must not be reused; to add it again create
// a new Device. If the bridge is not connected the removal is sent on the next
// connect.
func (bridge *Bridge) RemoveDevice(id string) error {
	bridge.devicesMu.Lock()
	defer bridge.devicesMu.Unlock()
	device, found := bridge.devices[id]
	if !found {
		return fmt.Errorf("device id not found in bridge")
	}
	delete(bridge.devices, id)
	bridge.removed[id] = struct{}{}

	// Close before queueing the removal so no update from the device can
	// follow it.
	device.Close()
	bridge.updates.Push(&clientsapi.StatusUpdate{
		RemovedDeviceIds: []string{id},
	})
	return nil
}

// Stops the bridge (and client) from running.
func (bridge *Bridge) Stop() {
	bridge.client.Stop()
}

// Run the bridge (and client).
func (bridge *Bridge) Run() error {
	return bridge.client.Run()
}

func (bridge *Bridge) deviceFeedback(inctx context.Context, conn *grpc.ClientConn) {
	bridge.log.Infof("device feedback started")
	defer bridge.log.Infof("device feedback finished")

	ctx, close := context.WithCancel(inctx)
	defer close()

	service := clientsapi.NewClientServiceClient(conn)
	stream, err := service.StatusStream(ctx)
	if err != nil {
		bridge.log.Errorf("failed to start status stream: %s", err)
		return
	}

	// Send the client info to the server.
	err = stream.Send(&clientsapi.StatusUpdate{
		ClientInfo: &clientsapi.ClientInfo{
			Id:          bridge.config.ClientID,
			Name:        bridge.config.ClientName,
			Description: bridge.config.ClientDescription,
			Version:     bridge.config.ClientVersion,
		},
	})
	if err != nil {
		bridge.log.Errorf("failed to send client info: %s", err)
	}

	// Stop discarding updates until we exit.
	bridge.updates.Discard(false)
	defer bridge.updates.Discard(true)

	bridge.queueFullState()

	// Wait for errors.
	go func() {
		err := stream.RecvMsg(nil)
		bridge.log.Errorf("failed to send device update: %s", err)
		close()
	}()

	// Now wait for updates.
	for {
		select {
		case <-ctx.Done():
			return

		case update := <-bridge.updates.Pop():
			// Send the update to the server.
			err := stream.Send(update)
			if err != nil {
				bridge.log.Errorf("failed to send status update: %s", err)
			}
		}
	}
}

// Queues the removed device IDs and the full state of every device. Called on
// each connect so removals made while disconnected still reach the server.
func (bridge *Bridge) queueFullState() {
	bridge.devicesMu.RLock()
	defer bridge.devicesMu.RUnlock()
	if len(bridge.removed) > 0 {
		bridge.updates.Push(&clientsapi.StatusUpdate{
			RemovedDeviceIds: slices.Sorted(maps.Keys(bridge.removed)),
		})
	}
	for _, dev := range bridge.devices {
		dev.SendFullState()
	}
}

func (bridge *Bridge) deviceControl(ctx context.Context, conn *grpc.ClientConn) {
	bridge.log.Infof("device control started")
	defer bridge.log.Infof("device control finished")

	service := clientsapi.NewClientServiceClient(conn)
	stream, err := service.ActionStream(ctx)
	if err != nil {
		bridge.log.Errorf("failed to start action stream: %s", err)
		return
	}

	for {
		req, err := stream.Recv()
		if err != nil {
			code := status.Code(err)
			if code == codes.Unavailable || code == codes.Canceled {
				bridge.log.Debugf("action stream closed: %s", err)
			} else {
				bridge.log.Errorf("failed to recv action request: %s", err)
			}
			return
		} else {
			bridge.log.Debugf("received action: %s", req)

			// Find the device.
			bridge.devicesMu.RLock()
			dev, found := bridge.devices[req.GetDeviceId()]
			bridge.devicesMu.RUnlock()
			if !found {
				bridge.log.Errorf("device not found for action: %s", req)
				err := stream.Send(&clientsapi.ActionResponse{
					ActionId: req.GetActionId(),
					Status:   clientsapi.ActionResponse_ERROR,
					Details:  "device not found",
				})
				if err != nil {
					bridge.log.Errorf("failed to send action response: %s", err)
				}
				continue
			}

			// Send an initial QUEUED response so the requester know's it got
			// somewhere.
			err := stream.Send(&clientsapi.ActionResponse{
				ActionId: req.GetActionId(),
				Status:   clientsapi.ActionResponse_QUEUED,
				Details:  "",
			})
			if err != nil {
				bridge.log.Errorf("failed to send action queued response: %s", err)
			}

			// Let the device handle it in another goroutine.
			go func() {
				lastStatus := clientsapi.ActionResponse_UNDEFINED
				err := dev.HandleAction(req, func(res *clientsapi.ActionResponse) {
					str := res.String()
					if len(str) > 1000 {
						str = str[:1000]
					}
					bridge.log.Debugf("sending action response: %s", str)
					lastStatus = res.Status
					err := stream.Send(res)
					if err != nil {
						bridge.log.Errorf("failed to send action response: %s", err)
					}
				})
				if err != nil {
					bridge.log.Debugf("sending action error response: %s", err)
					err := stream.Send(&clientsapi.ActionResponse{
						ActionId: req.GetActionId(),
						Status:   clientsapi.ActionResponse_ERROR,
						Details:  err.Error(),
					})
					if err != nil {
						bridge.log.Errorf("failed to send action error response: %s", err)
					}
				} else {
					// Auto return complete if no other final status was sent.
					if lastStatus < clientsapi.ActionResponse_COMPLETE {
						res := &clientsapi.ActionResponse{
							ActionId: req.GetActionId(),
							Status:   clientsapi.ActionResponse_COMPLETE,
							Details:  "",
						}
						bridge.log.Debugf("sending action auto response: %s", res)
						err := stream.Send(res)
						if err != nil {
							bridge.log.Errorf("failed to send action auto response: %s", err)
						}
					}
				}
			}()
		}
	}
}

func (bridge *Bridge) imagesControl(ctx context.Context, conn *grpc.ClientConn) {
	bridge.log.Infof("image control started")
	defer bridge.log.Infof("image control finished")

	service := clientsapi.NewClientServiceClient(conn)
	stream, err := service.ImageStream(ctx)
	if err != nil {
		bridge.log.Errorf("failed to start image stream: %s", err)
		return
	}

	for {
		req, err := stream.Recv()
		if err != nil {
			code := status.Code(err)
			if code == codes.Unavailable || code == codes.Canceled {
				bridge.log.Debugf("image stream closed: %s", err)
			} else {
				bridge.log.Errorf("failed to recv image request: %s", err)
			}
			return
		} else {
			bridge.log.Debugf("received image request: %s", req)

			// Find the device.
			bridge.devicesMu.RLock()
			dev, found := bridge.devices[req.GetDeviceId()]
			bridge.devicesMu.RUnlock()
			if !found {
				err := stream.Send(&clientsapi.ImageResponse{
					RequestId: req.GetRequestId(),
					Status:    clientsapi.ImageResponse_ERROR,
					Details:   "device not found",
				})
				if err != nil {
					bridge.log.Errorf("failed to send image response: %s", err)
				}
				continue
			}

			// Let the device handle it in another goroutine.
			go func() {
				lastStatus := clientsapi.ImageResponse_UNDEFINED
				data, err := dev.HandleImage(req, func(res *clientsapi.ImageResponse) {
					bridge.log.Debugf("sending image response: %s", apitools.ImageResponseString(res))
					lastStatus = res.Status
					err := stream.Send(res)
					if err != nil {
						bridge.log.Errorf("failed to send image response: %s", err)
					}
				})
				if err != nil {
					bridge.log.Debugf("sending image error response: %s", err)
					err := stream.Send(&clientsapi.ImageResponse{
						RequestId: req.GetRequestId(),
						Status:    clientsapi.ImageResponse_ERROR,
						Details:   err.Error(),
					})
					if err != nil {
						bridge.log.Errorf("failed to send image error response: %s", err)
					}
				} else {
					// Auto return complete if no other final status was sent.
					if lastStatus < clientsapi.ImageResponse_COMPLETE {
						res := &clientsapi.ImageResponse{
							RequestId: req.GetRequestId(),
							Status:    clientsapi.ImageResponse_COMPLETE,
							Details:   "",
							Data:      data,
						}
						bridge.log.Debugf("sending image auto response: %s", apitools.ImageResponseString(res))
						err := stream.Send(res)
						if err != nil {
							bridge.log.Errorf("failed to send image auto response: %s", err)
						}
					}
				}
			}()
		}
	}
}
