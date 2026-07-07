package clients

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v1/shared/crypt"
	"github.com/jimjibone/wh/v1/shared/random"
	"github.com/jimjibone/wh/v1/shared/sas"
	"github.com/jimjibone/wh/v1/shared/stores"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
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

	clientID          string
	clientName        string
	clientDescription string
	clientVersion     string

	stopCtx context.Context
	stop    func()

	minBackoff  time.Duration
	maxBackoff  time.Duration
	lastBackoff time.Duration

	handlers []ConnectionHandler

	serviceMu sync.RWMutex
	service   clientsapi.ClientServiceClient
}

type ClientConfig struct {
	// Store in which to keep persistent data.
	Store stores.Store

	// Server address to connect to.
	ServerAddr string

	// Optional ID for this client. If not set a random name is generate on
	// first run and saved in the store.
	ClientID string

	// Optional name for this client.
	ClientName string

	// Optional description for this client.
	ClientDescription string

	// Optional version string for this client.
	ClientVersion string
}

// Create a new woodhouse client. The store is used to keep pairing secrets
// between executions of the client. The serverAddr is the address of the
// woodhouse server.
func NewClient(config ClientConfig, opts ...ClientOption) *Client {
	// Just panic if the configuration isn't valid.
	if config.Store == nil {
		log.Fatalf("store not defined in client config")
	}
	if config.ServerAddr == "" {
		log.Fatalf("server address not defined in client config")
	}
	if _, _, err := net.SplitHostPort(config.ServerAddr); err != nil {
		log.Fatalf("server address not valid in client config: %s", err)
	}

	client := &Client{
		log:        log.NewContext(log.DefaultLogger, "client", log.DebugLevel),
		store:      newClientStore(config.Store),
		serverAddr: config.ServerAddr,

		clientID:          config.ClientID,
		clientName:        config.ClientName,
		clientDescription: config.ClientDescription,
		clientVersion:     config.ClientVersion,

		minBackoff:  time.Second,
		maxBackoff:  32 * time.Second,
		lastBackoff: 0,
	}

	client.stopCtx, client.stop = context.WithCancel(context.Background())

	for _, o := range opts {
		o(client)
	}

	return client
}

type ClientOption func(*Client)

type ConnectionHandler func(ctx context.Context, conn *grpc.ClientConn)

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

	// Generate our ephemeral key (PKa) for the SAS key agreement.
	clientPriv, err := sas.GenerateKey()
	if err != nil {
		client.log.Errorf("pairing failed to generate key: %s", err)
		return false
	}
	pka := clientPriv.PublicKey().Bytes()

	// 1. Send our ID and ephemeral public key to the server.
	err = rpc.Send(&clientsapi.PairRequest{
		ClientId:     client.clientID,
		ClientPubkey: pka,
		// TODO: send client name, description and version when pairing
	})
	if err != nil {
		client.log.Errorf("pairing failed to send client id: %s", err)
		return false
	}

	// 2. Receive the server's public key (PKb) and its commitment to Nb.
	res, err := rpc.Recv()
	if err != nil {
		code := status.Code(err)
		switch code {
		case codes.Unavailable:
			client.log.Debugf("pairing failed to start: server offline")
		case codes.AlreadyExists:
			client.log.Errorf("pairing already in progress for this client id")
		default:
			client.log.Errorf("pairing failed to receive key exchange: %s", err)
		}
		return false
	}
	if res.State != clientsapi.PairResponse_KeyExchange {
		client.log.Errorf("pairing unexpected state, want key exchange: %s", res.State)
		return false
	}
	pkb := res.ServerPubkey
	commitment := res.Commitment
	serverPub, err := sas.ParsePublicKey(pkb)
	if err != nil {
		client.log.Errorf("pairing server sent an invalid public key: %s", err)
		return false
	}

	// 3. Send our nonce (Na).
	na, err := sas.Nonce()
	if err != nil {
		client.log.Errorf("pairing failed to generate nonce: %s", err)
		return false
	}
	err = rpc.Send(&clientsapi.PairRequest{ClientNonce: na})
	if err != nil {
		client.log.Errorf("pairing failed to send nonce: %s", err)
		return false
	}

	// 4. Receive the server's revealed nonce (Nb) and verify the commitment.
	res, err = rpc.Recv()
	if err != nil {
		client.log.Errorf("pairing failed to receive reveal: %s", err)
		return false
	}
	if res.State != clientsapi.PairResponse_Reveal {
		client.log.Errorf("pairing unexpected state, want reveal: %s", res.State)
		return false
	}
	nb := res.ServerNonce
	if !sas.VerifyCommit(commitment, pkb, pka, client.clientID, nb) {
		client.log.Errorf("pairing commitment verification failed - possible man-in-the-middle, aborting")
		return false
	}

	// 5. Derive the SAS and AES-256 session key.
	sasCode, key, err := sas.Derive(clientPriv, serverPub, pka, pkb, client.clientID, na, nb)
	if err != nil {
		client.log.Errorf("pairing failed to derive sas: %s", err)
		return false
	}

	// 6. Show the SAS for the user to compare against the woodhouse web UI.
	client.log.Infof("pairing code: %s (confirm it matches the code shown in the woodhouse web UI)", sas.Grouped(sasCode))

	// 7. Wait for the server to deliver the credentials once the user confirms,
	// ignoring Pending keepalives. A denial or timeout ends the stream with an
	// error.
	for {
		res, err = rpc.Recv()
		if err != nil {
			code := status.Code(err)
			switch code {
			case codes.PermissionDenied:
				client.log.Errorf("pairing denied by server")
			case codes.DeadlineExceeded:
				client.log.Errorf("pairing timed out awaiting confirmation")
			default:
				client.log.Errorf("pairing failed to receive cert: %s", err)
			}
			return false
		}
		if res.State == clientsapi.PairResponse_Pending {
			continue
		}
		if res.State == clientsapi.PairResponse_Confirmed {
			break
		}
		client.log.Errorf("pairing unexpected state, want confirmed: %s", res.State)
		return false
	}

	// 8. Decrypt the server's certificate.
	cert, err := crypt.Decrypt(res.Data, key)
	if err != nil {
		client.log.Errorf("pairing failed to decrypt cert: %s", err)
		return false
	}

	// 9. Receive and decrypt our new refresh token.
	res, err = rpc.Recv()
	if err != nil {
		client.log.Errorf("pairing failed to receive token: %s", err)
		return false
	}
	tokenBytes, err := crypt.Decrypt(res.Data, key)
	if err != nil {
		client.log.Errorf("pairing failed to decrypt token: %s", err)
		return false
	}
	token := string(tokenBytes)

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
			handlerCancel()
			wg.Done()
		}(handler)
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
