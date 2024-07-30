package wh

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/jimjibone/queue/v2"
	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
	"github.com/jimjibone/woodhouse-4/apitools"
	"github.com/jimjibone/woodhouse-4/log"
	"github.com/jimjibone/woodhouse-4/shared/crypt"
	"github.com/jimjibone/woodhouse-4/shared/random"
	"github.com/jimjibone/woodhouse-4/shared/stores"
	"github.com/jimjibone/woodhouse-4/wh/v1/devices"
	"github.com/jimjibone/woodhouse-4/wh/v1/reactors"
	"github.com/schollz/pake/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

var (
	ErrNotConnected = errors.New("not connected to server")
)

type Client struct {
	log        *log.Context
	store      *clientStore
	serverAddr string
	clientID   string
	stopCtx    context.Context
	stop       func()

	minBackoff  time.Duration
	maxBackoff  time.Duration
	lastBackoff time.Duration

	handlers []ConnectionHandler

	devicesMu sync.RWMutex               // locks the devices map only
	devices   map[string]*devices.Device // key=id
	updates   *queue.Queue[*clientsapi.Device]

	reactorsEnabled bool
	reactorsMu      sync.RWMutex // locks the reactors map only
	reactors        map[string]*reactorDevice

	imagesEnabled bool

	serviceMu sync.RWMutex
	service   clientsapi.ClientServiceClient
}

type reactorDevice struct {
	reactors map[*reactors.Device]struct{}
}

// Create a new woodhouse client. The store is used to keep pairing secrets
// between executions of the client. The serverAddr is the address of the
// woodhouse server.
func NewClient(store stores.Store, serverAddr string, opts ...ClientOption) *Client {
	client := &Client{
		log:        log.NewContext(log.DefaultLogger, "client", log.DebugLevel),
		store:      newClientStore(store),
		serverAddr: serverAddr,
		clientID:   "",

		minBackoff:  time.Second,
		maxBackoff:  32 * time.Second,
		lastBackoff: 0,

		devices: make(map[string]*devices.Device),
		updates: queue.New[*clientsapi.Device](),

		reactors: make(map[string]*reactorDevice),
	}

	client.stopCtx, client.stop = context.WithCancel(context.Background())

	// Discard updates until we're connected to the server.
	client.updates.Discard(true)

	for _, o := range opts {
		o(client)
	}

	return client
}

type ClientOption func(*Client)

type ConnectionHandler func(ctx context.Context, conn *grpc.ClientConn)

// Sets the client ID manually. Overrides the default option of generating one
// automatically.
func WithClientID(id string) ClientOption {
	return func(c *Client) {
		c.clientID = id
	}
}

// Set log level. Overrides the default of warnings and above.
func WithLogLevel(level log.Level) ClientOption {
	return func(c *Client) {
		c.log = log.NewContext(log.DefaultLogger, "client", level)
	}
}

// Set log level. Overrides the default of warnings and above.
func WithConnectionHandler(handler ConnectionHandler) ClientOption {
	return func(c *Client) {
		c.handlers = append(c.handlers, handler)
	}
}

// Enable the reactors functionality.
func WithReactors() ClientOption {
	return func(c *Client) {
		c.reactorsEnabled = true
	}
}

// Enable the images functionality.
func WithImages() ClientOption {
	return func(c *Client) {
		c.imagesEnabled = true
	}
}

// Add a device to the client.
func (client *Client) AddDevice(device *devices.Device) error {
	client.devicesMu.Lock()
	defer client.devicesMu.Unlock()
	if _, found := client.devices[device.ID()]; found {
		return fmt.Errorf("device id already exists in client")
	}
	client.devices[device.ID()] = device
	device.Init(func(state *clientsapi.Device) {
		client.updates.Push(state)
	})
	device.SendFullState()
	return nil
}

// Add a reactor to the client.
func (client *Client) AddReactor(device *reactors.Device) {
	if !client.reactorsEnabled {
		panic("reactors are not enabled")
	}
	if device.ID() == "" {
		panic("reactor device has empty ID")
	}
	client.reactorsMu.Lock()
	defer client.reactorsMu.Unlock()
	if devs, found := client.reactors[device.ID()]; !found {
		client.reactors[device.ID()] = &reactorDevice{reactors: map[*reactors.Device]struct{}{
			device: {},
		}}
	} else {
		devs.reactors[device] = struct{}{}
	}
}

// Remove a reactor from the client.
func (client *Client) RemoveReactor(device *reactors.Device) {
	if !client.reactorsEnabled {
		panic("reactors are not enabled")
	}
	client.reactorsMu.Lock()
	defer client.reactorsMu.Unlock()
	if devs, found := client.reactors[device.ID()]; found {
		delete(devs.reactors, device)
		if len(devs.reactors) == 0 {
			delete(client.reactors, device.ID())
		}
	}
}

func (client *Client) RequestAction(ctx context.Context, req *clientsapi.ActionRequest, handler func(resp *clientsapi.ActionResponse)) error {
	if handler == nil {
		panic("handler is nil")
	}
	client.serviceMu.RLock()
	defer client.serviceMu.RUnlock()
	if client.service == nil {
		return ErrNotConnected
	}
	stream, err := client.service.SendAction(ctx, req)
	if err != nil {
		return err
	}
	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		handler(resp)
	}
}

func (client *Client) RequestImage(ctx context.Context, req *clientsapi.ImageRequest, handler func(resp *clientsapi.ImageResponse)) error {
	if handler == nil {
		panic("handler is nil")
	}
	client.serviceMu.RLock()
	defer client.serviceMu.RUnlock()
	if client.service == nil {
		return ErrNotConnected
	}
	stream, err := client.service.SendImageRequest(ctx, req)
	if err != nil {
		return err
	}
	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		handler(resp)
	}
}

// Stops the client from running.
func (client *Client) Stop() {
	client.stop()
}

func (client *Client) Run() error {
	// Upgrade the store to the latest schema.
	err := client.store.Upgrade(client.log)
	if err != nil {
		return fmt.Errorf("failed to upgrade the store: %w", err)
	}

	// Get the client id.
	if client.clientID == "" {
		// If the store doesn't have an id then generate a new one.
		if !client.store.HasID() {
			name, err := random.GenerateRandomName(2)
			if err != nil {
				client.log.Fatalf("failed to generate client ID: %s", err)
			}
			client.clientID = name

			// Write it to the store.
			err = client.store.SetID([]byte(client.clientID))
			if err != nil {
				return fmt.Errorf("failed to write client id to store: %s", err)
			}
		} else {
			// Read it from the store.
			data, err := client.store.GetID()
			if err != nil {
				return fmt.Errorf("failed to read client id from store: %s", err)
			}
			client.clientID = string(data)
		}
	}

	// Log useful info.
	client.log.Debugf("run started")
	defer client.log.Debugf("run finished")
	client.log.Debugf("server addr: %s", client.serverAddr)
	client.log.Debugf("client id: %s", client.clientID)

	// Listen for interrupts.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(client.stopCtx)
	defer cancel()
	go func() {
		<-c
		// Stop delivering signals.
		signal.Stop(c)
		// Cancel the context to stop the server.
		cancel()
	}()

	// Now run forever or until the context is done.
	done := false
	for !done {
		// Start by pinging the server. This prevents us showing pairing error
		// messages if the server is offline.
		online := client.ping(ctx)

		connected := false
		if online {
			// Now do the pairing. If we're already paired then this will
			// instantly return true.
			paired := client.pair(ctx)

			if paired {
				// Connect to the server with the pairing credentials. If this
				// actually connected to the server then it will return true.
				connected = client.connect(ctx)
			}
		}

		// If the client didn't connect then implement exponential backoff.
		client.backoff(ctx, connected)

		// Exit the loop if the context is done.
		select {
		case <-ctx.Done():
			done = true
		default:
		}
	}

	return nil
}

// Ping the server and return true if the server responded.
func (client *Client) ping(ctx context.Context) bool {
	client.log.Debugf("ping started")
	defer client.log.Debugf("ping finished")

	// Require TLS but we don't care about trusting it, we'll sort that out in a
	// moment.
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})

	// Connect to the server.
	connCtx, connCancel := context.WithCancel(ctx)
	defer connCancel()
	conn, err := grpc.DialContext(
		connCtx,
		client.serverAddr,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		client.log.Errorf("pairing connection failed: %s", err)
		return false
	}
	defer conn.Close()

	// Create the service and send the ping.
	service := clientsapi.NewAuthServiceClient(conn)

	// Send the ping until successful.
	firstLog := true
	for {
		pingCtx, pingCancel := context.WithTimeout(ctx, 10*time.Second)
		defer pingCancel()
		_, err = service.Ping(pingCtx, &clientsapi.PingRequest{})
		if err != nil {
			if code := status.Code(err); code == codes.Unavailable {
				client.log.Debugf("ping: server offline: %s", err)
			} else {
				client.log.Errorf("ping: server offline: %s", err)
			}
		} else {
			return true
		}

		// If the server didn't respond on the first attempt then mention that
		// it's offline.
		if firstLog {
			firstLog = false
			client.log.Infof("waiting for server to come online")
		}

		// If the client didn't connect then implement exponential backoff.
		client.backoff(ctx, false)

		// Exit the loop if the context is done.
		select {
		case <-ctx.Done():
			return false
		default:
		}
	}
}

// Attempt to pair with the server. If the client is already paired it will
// return true instantly. If not it will try to pair with the server and
// eventually return true. If the pair was unsuccessful this will return false.
func (client *Client) pair(ctx context.Context) bool {
	if client.store.HasToken() && client.store.HasCert() {
		client.log.Debugf("using previous token and cert")
		return true
	}

	client.log.Infof("pairing started")
	defer client.log.Infof("pairing finished")

	// Require TLS but we don't care about trusting it, we'll sort that out in a
	// moment.
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})

	// Connect to the server.
	connCtx, connCancel := context.WithTimeout(ctx, 10*time.Second)
	defer connCancel()
	conn, err := grpc.DialContext(
		connCtx,
		client.serverAddr,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		client.log.Errorf("pairing connection failed: %s", err)
		return false
	}
	defer conn.Close()

	service := clientsapi.NewAuthServiceClient(conn)

	pairCtx, pairCancel := context.WithCancel(ctx)
	defer pairCancel()
	rpc, err := service.Pair(pairCtx)
	if err != nil {
		code := status.Code(err)
		if code == codes.Unavailable {
			client.log.Debugf("pairing failed to start: server offline")
		} else {
			client.log.Errorf("pairing failed to start: %s", err)
		}
		return false
	}

	// 1. Send our ID to the server.
	err = rpc.Send(&clientsapi.PairRequest{
		ClientId: client.clientID,
	})
	if err != nil {
		client.log.Errorf("pairing failed to send client id: %s", err)
		return false
	}

	// 2. Wait for server to move on from pending.
	var res *clientsapi.PairResponse
	pending := true
	for pending {
		res, err = rpc.Recv()
		if err != nil {
			code := status.Code(err)
			if code == codes.PermissionDenied {
				client.log.Errorf("pairing denied by server")
				return false
			}
			client.log.Errorf("pairing failed to receive message: %s", err)
			return false
		}
		switch res.State {
		case clientsapi.PairResponse_Handshake:
			client.log.Debugf("pairing handshake started...")
			pending = false
		case clientsapi.PairResponse_Pending:
			client.log.Debugf("pairing pending...")
		default:
			client.log.Errorf("pairing receive unexpected state: %s", res.State)
			return false
		}
	}

	// Start the PAKE handshake using the key generated by us.
	client.log.Debugf("pairing initialising pake")
	pakep, err := pake.InitCurve([]byte("redacted"), 1, "p521")
	if err != nil {
		client.log.Errorf("pairing failed to init pake: %s", err)
		return false
	}

	// 3. Receive the first PAKE handshake blob from the server.
	// NOTE: we already received the first blob in the for loop above.
	client.log.Debugf("pairing received first handshake blob")
	err = pakep.Update(res.Data)
	if err != nil {
		client.log.Errorf("pairing failed to update handshake: %s", err)
		return false
	}

	// 4. Send the second handshake blob to the server.
	client.log.Debugf("pairing sending second handshake blob")
	err = rpc.Send(&clientsapi.PairRequest{
		Data: pakep.Bytes(),
	})
	if err != nil {
		client.log.Errorf("pairing failed to send handshake: %s", err)
		return false
	}

	// 5. We now have the session key. The server will now test us to
	// make sure we both have the same session key.
	client.log.Debugf("pairing waiting for handshake test")
	res, err = rpc.Recv()
	if err != nil {
		client.log.Errorf("pairing failed to receive test: %s", err)
		return false
	}
	key, err := pakep.SessionKey()
	if err != nil {
		client.log.Errorf("pairing failed to get session key: %s", err)
		return false
	}
	decrypted, err := crypt.Decrypt(res.Data, key)
	if err != nil {
		client.log.Errorf("pairing failed to decrypt test: %s", err)
		return false
	}
	client.log.Debugf("pairing received test")

	// 6. Reverse the bytes, then re-encrypt and send back the test.
	slices.Reverse(decrypted)
	encrypted, err := crypt.Encrypt(decrypted, key)
	if err != nil {
		client.log.Errorf("pairing failed to encrypt test: %s", err)
		return false
	}
	err = rpc.Send(&clientsapi.PairRequest{
		Data: encrypted,
	})
	if err != nil {
		client.log.Errorf("pairing failed to send test: %s", err)
		return false
	}

	// 7. Receive the server's certificate and save it.
	res, err = rpc.Recv()
	if err != nil {
		client.log.Errorf("pairing failed to receive cert: %s", err)
		return false
	}
	decrypted, err = crypt.Decrypt(res.Data, key)
	if err != nil {
		client.log.Errorf("pairing failed to decrypt cert: %s", err)
		return false
	}
	cert := decrypted
	client.log.Debugf("pairing server cert is %s", decrypted)

	// 8. Receive our new auth tokens and save them.
	res, err = rpc.Recv()
	if err != nil {
		client.log.Errorf("pairing failed to receive token: %s", err)
		return false
	}
	decrypted, err = crypt.Decrypt(res.Data, key)
	if err != nil {
		client.log.Errorf("pairing failed to decrypt token: %s", err)
		return false
	}
	token := string(decrypted)
	client.log.Debugf("pairing client token is %s", token)

	// Save token and cert to the store.
	err = client.store.SetToken([]byte(token))
	if err != nil {
		client.log.Errorf("pairing failed to write token: %s", err)
		return false
	}
	err = client.store.SetCert(cert)
	if err != nil {
		client.log.Errorf("pairing failed to write cert: %s", err)
		return false
	}

	return true
}

// Connects to the server using the stored secrets gathered during pairing. If
// the connection was successful it will return true, otherwise it will return
// false if the connection failed instantly.
func (client *Client) connect(ctx context.Context) bool {
	token, err := client.store.GetToken()
	if err != nil {
		client.log.Errorf("failed to read token from store: %s", err)

		// Delete the token from the store to trigger pairing.
		err = client.store.DelToken()
		if err != nil {
			client.log.Errorf("failed to delete token from store: %s", err)
		}
		return false
	}
	cert, err := client.store.GetCert()
	if err != nil {
		client.log.Errorf("failed to read cert from store: %s", err)

		// Delete the cert from the store to trigger pairing.
		err = client.store.DelCert()
		if err != nil {
			client.log.Errorf("failed to delete cert from store: %s", err)
		}
		return false
	}

	client.log.Infof("connection started")
	defer client.log.Infof("connection finished")

	// Require TLS and now we care about trusting it. Use the server cert we
	// got previously.
	certpool := x509.NewCertPool()
	ok := certpool.AppendCertsFromPEM(cert)
	if !ok {
		// The cert is probably bad, so trigger pairing by deleting it.
		client.log.Errorf("failed to load server cert")
		client.store.DelCert()
		return false
	}
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            certpool,
		ServerName:         "woodhouse",
	})

	// Intercept server requests and add auth tokens.
	auth := NewAuthInterceptor(token, func(token []byte) {
		if err := client.store.SetToken(token); err != nil {
			client.log.Errorf("failed to save token: %s", err)
		}
	})
	defer auth.Close()

	// Create a connection to the server for regular requests.
	connCtx, connCancel := context.WithTimeout(ctx, 10*time.Second)
	defer connCancel()
	conn, err := grpc.DialContext(
		connCtx,
		client.serverAddr,
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(auth.Unary()),
		grpc.WithStreamInterceptor(auth.Stream()),
	)
	if err != nil {
		client.log.Errorf("connection failed: %s", err)
		return false
	}
	defer conn.Close()

	// Create a client service client which can be used by request methods.
	client.serviceMu.Lock()
	client.service = clientsapi.NewClientServiceClient(conn)
	client.serviceMu.Unlock()
	defer func() {
		client.serviceMu.Lock()
		client.service = nil
		client.serviceMu.Unlock()
	}()

	// Start the auth (fetches a new access token).
	err = auth.Start(conn)
	if err != nil {
		client.log.Errorf("failed to create auth: %s", err)

		// If we've been unauthenticated delete the token from the store to
		// trigger pairing.
		if code := status.Code(err); code == codes.Unauthenticated {
			client.log.Infof("resetting auth to trigger pairing")
			auth.Reset()
		}
		return false
	}

	// Start connection handlers.
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	handlerCtx, handlerCancel := context.WithCancel(context.Background())
	defer handlerCancel()
	for _, handler := range client.handlers {
		wg.Add(1)
		go func(handler ConnectionHandler) {
			handler(handlerCtx, conn)
			wg.Done()
		}(handler)
	}

	// Start the device control/feedback loop.
	wg.Add(2)
	go client.deviceFeedback(handlerCtx, handlerCancel, wg, conn)
	go client.deviceControl(handlerCtx, handlerCancel, wg, conn)

	// Start the images comms if enabled.
	if client.imagesEnabled {
		wg.Add(1)
		go client.imagesControl(handlerCtx, handlerCancel, wg, conn)
	}

	// Start the reactor comms if enabled.
	if client.reactorsEnabled {
		wg.Add(1)
		go client.reactorControl(handlerCtx, handlerCancel, wg, conn)
	}

	// Monitor the connection and return if it closes.
	client.log.Debugf("connection complete")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	done := false
	for !done {
		select {
		case <-ctx.Done():
			// Exit if the context is closed.
			done = true

		case <-handlerCtx.Done():
			// Exit if the context is closed.
			done = true

		case <-ticker.C:
			// Check the connection.
			if err := auth.Ping(handlerCtx); err != nil {
				if code := status.Code(err); code == codes.Unavailable {
					client.log.Debugf("server went offline: %s", err)
				} else {
					client.log.Errorf("server went offline - ping error: %s", err)
				}
				done = true
			}
		}
	}
	client.log.Debugf("connection finishing")

	return true
}

// Implements an exponential backoff by sleeping the goroutine for an increasing
// amount of time, up to the maxBackoff, unless reset is true when it will
// return the backoff to minBackoff.
func (client *Client) backoff(ctx context.Context, reset bool) {
	if reset {
		client.lastBackoff = client.minBackoff
	} else {
		client.lastBackoff = client.lastBackoff * 2
	}
	if client.lastBackoff <= 0 {
		client.lastBackoff = client.minBackoff
	}
	if client.lastBackoff > client.maxBackoff {
		client.lastBackoff = client.maxBackoff
	}
	client.log.Debugf("backoff for %s", client.lastBackoff)
	timer := time.NewTimer(client.lastBackoff)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func (client *Client) deviceFeedback(ctx context.Context, close func(), wg *sync.WaitGroup, conn *grpc.ClientConn) {
	defer close()
	defer wg.Done()

	client.log.Infof("device feedback started")
	defer client.log.Infof("device feedback finished")

	service := clientsapi.NewClientServiceClient(conn)
	stream, err := service.StatusStream(ctx)
	if err != nil {
		client.log.Errorf("failed to start status stream: %s", err)
		return
	}

	// Stop discarding updates until we exit.
	client.updates.Discard(false)
	defer client.updates.Discard(true)

	// Get all devices to send their full states.
	client.devicesMu.RLock()
	for _, dev := range client.devices {
		dev.SendFullState()
	}
	client.devicesMu.RUnlock()

	// Now wait for updates.
	for {
		select {
		case <-ctx.Done():
			return

		case update := <-client.updates.Pop():
			// Send the update to the server.
			err := stream.Send(&clientsapi.StatusUpdate{
				DeviceInfo: []*clientsapi.Device{
					update,
				},
			})
			if err != nil {
				client.log.Errorf("failed to send device update: %s", err)
			}
		}
	}
}

func (client *Client) deviceControl(ctx context.Context, close func(), wg *sync.WaitGroup, conn *grpc.ClientConn) {
	defer close()
	defer wg.Done()

	client.log.Infof("device control started")
	defer client.log.Infof("device control finished")

	service := clientsapi.NewClientServiceClient(conn)
	stream, err := service.ActionStream(ctx)
	if err != nil {
		client.log.Errorf("failed to start action stream: %s", err)
		return
	}

	for {
		req, err := stream.Recv()
		if err != nil {
			code := status.Code(err)
			if code == codes.Unavailable || code == codes.Canceled {
				client.log.Debugf("action stream closed: %s", err)
			} else {
				client.log.Errorf("failed to recv action request: %s", err)
			}
			return
		} else {
			client.log.Debugf("received action: %s", req)

			// Find the device.
			client.devicesMu.RLock()
			dev, found := client.devices[req.GetDeviceId()]
			client.devicesMu.RUnlock()
			if !found {
				err := stream.Send(&clientsapi.ActionResponse{
					ActionId: req.GetActionId(),
					Status:   clientsapi.ActionResponse_ERROR,
					Details:  "device not found",
				})
				if err != nil {
					client.log.Errorf("failed to send action response: %s", err)
				}
			}

			// Let the device handle it in another goroutine.
			go func() {
				lastStatus := clientsapi.ActionResponse_UNDEFINED
				err := dev.HandleAction(req, func(res *clientsapi.ActionResponse) {
					client.log.Debugf("sending action response: %s", res)
					lastStatus = res.Status
					err := stream.Send(res)
					if err != nil {
						client.log.Errorf("failed to send action response: %s", err)
					}
				})
				if err != nil {
					client.log.Debugf("sending action error response: %s", err)
					err := stream.Send(&clientsapi.ActionResponse{
						ActionId: req.GetActionId(),
						Status:   clientsapi.ActionResponse_ERROR,
						Details:  err.Error(),
					})
					if err != nil {
						client.log.Errorf("failed to send action error response: %s", err)
					}
				} else {
					// Auto return complete if no other final status was sent.
					if lastStatus < clientsapi.ActionResponse_COMPLETE {
						res := &clientsapi.ActionResponse{
							ActionId: req.GetActionId(),
							Status:   clientsapi.ActionResponse_COMPLETE,
							Details:  "",
						}
						client.log.Debugf("sending action auto response: %s", res)
						err := stream.Send(res)
						if err != nil {
							client.log.Errorf("failed to send action auto response: %s", err)
						}
					}
				}
			}()
		}
	}
}

func (client *Client) reactorControl(ctx context.Context, close func(), wg *sync.WaitGroup, conn *grpc.ClientConn) {
	defer close()
	defer wg.Done()

	client.log.Infof("reactor control started")
	defer client.log.Infof("reactor control finished")

	service := clientsapi.NewClientServiceClient(conn)
	stream, err := service.DeviceStream(ctx, &clientsapi.DeviceStreamRequest{})
	if err != nil {
		client.log.Errorf("failed to start reactor stream: %s", err)
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
				client.log.Debugf("reactor stream closed: %s", err)
			} else {
				client.log.Errorf("failed to recv reactor request: %s", err)
			}
			return
		} else {
			// client.log.Debugf("received reactor update: %s", update)

			// Find the reactors for this device update.
			client.reactorsMu.RLock()
			reacts, found := client.reactors[update.GetId()]
			client.reactorsMu.RUnlock()

			// If found then pass the update to the reactors.
			if found {
				for reactor := range reacts.reactors {
					if reactor != nil {
						reactor.HandleUpdate(update)
					}
				}
			}
		}
	}
}

func (client *Client) imagesControl(ctx context.Context, close func(), wg *sync.WaitGroup, conn *grpc.ClientConn) {
	defer close()
	defer wg.Done()

	client.log.Infof("image control started")
	defer client.log.Infof("image control finished")

	service := clientsapi.NewClientServiceClient(conn)
	stream, err := service.ImageStream(ctx)
	if err != nil {
		client.log.Errorf("failed to start image stream: %s", err)
		return
	}

	for {
		req, err := stream.Recv()
		if err != nil {
			code := status.Code(err)
			if code == codes.Unavailable || code == codes.Canceled {
				client.log.Debugf("image stream closed: %s", err)
			} else {
				client.log.Errorf("failed to recv image request: %s", err)
			}
			return
		} else {
			client.log.Debugf("received image request: %s", req)

			// Find the device.
			client.devicesMu.RLock()
			dev, found := client.devices[req.GetDeviceId()]
			client.devicesMu.RUnlock()
			if !found {
				err := stream.Send(&clientsapi.ImageResponse{
					RequestId: req.GetRequestId(),
					Status:    clientsapi.ImageResponse_ERROR,
					Details:   "device not found",
				})
				if err != nil {
					client.log.Errorf("failed to send image response: %s", err)
				}
			}

			// Let the device handle it in another goroutine.
			go func() {
				lastStatus := clientsapi.ImageResponse_UNDEFINED
				data, err := dev.HandleImage(req, func(res *clientsapi.ImageResponse) {
					client.log.Debugf("sending image response: %s", apitools.ImageResponseString(res))
					lastStatus = res.Status
					err := stream.Send(res)
					if err != nil {
						client.log.Errorf("failed to send image response: %s", err)
					}
				})
				if err != nil {
					client.log.Debugf("sending image error response: %s", err)
					err := stream.Send(&clientsapi.ImageResponse{
						RequestId: req.GetRequestId(),
						Status:    clientsapi.ImageResponse_ERROR,
						Details:   err.Error(),
					})
					if err != nil {
						client.log.Errorf("failed to send image error response: %s", err)
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
						client.log.Debugf("sending image auto response: %s", apitools.ImageResponseString(res))
						err := stream.Send(res)
						if err != nil {
							client.log.Errorf("failed to send image auto response: %s", err)
						}
					}
				}
			}()
		}
	}
}
