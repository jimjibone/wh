package wh

import (
	"context"
	"fmt"
	"sync"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
	"github.com/jimjibone/woodhouse-4/apitools"
	"github.com/jimjibone/woodhouse-4/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	RefreshTokensInterval = 14 * time.Minute
)

type AuthInterceptor struct {
	log          *log.Context
	ctx          context.Context
	cancel       func()
	wg           sync.WaitGroup
	client       clientsapi.AuthServiceClient
	refreshToken string
	accessToken  string
	changed      bool
	saver        func(token []byte)
}

func NewAuthInterceptor(refreshToken []byte, saver func(token []byte)) *AuthInterceptor {
	auth := &AuthInterceptor{
		log:          log.NewContext(log.DefaultLogger, "auth", log.DebugLevel),
		client:       nil,
		refreshToken: string(refreshToken),
		saver:        saver,
	}

	return auth
}

func (auth *AuthInterceptor) Close() {
	if auth.cancel != nil {
		auth.cancel()
	}
	auth.wg.Wait()
	if auth.changed {
		auth.changed = false
		auth.saver([]byte(auth.refreshToken))
	}
}

func (auth *AuthInterceptor) Start(conn *grpc.ClientConn) error {
	auth.client = clientsapi.NewAuthServiceClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	auth.cancel = cancel

	err := auth.refresh(ctx)
	if err != nil {
		return err
	}

	auth.wg.Add(1)
	go func() {
		defer auth.wg.Done()
		ticker := time.NewTicker(RefreshTokensInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := auth.refresh(ctx); err != nil {
					auth.log.Errorf("refresh failed: %s", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Reset auth tokens to force pairing.
func (auth *AuthInterceptor) Reset() {
	auth.refreshToken = ""
	auth.accessToken = ""
	auth.changed = true
}

func (auth *AuthInterceptor) Ping(ctx context.Context) error {
	if auth.client != nil {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		_, err := auth.client.Ping(ctx, &clientsapi.PingRequest{})
		return err
	}
	return fmt.Errorf("no client")
}

func (auth *AuthInterceptor) refresh(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	res, err := auth.client.Refresh(ctx, &clientsapi.RefreshRequest{
		RefreshToken: auth.refreshToken,
	})
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			// Erase the refresh token if unauthenticated.
			auth.Reset()
		}
		return err
	}

	if auth.refreshToken != res.GetRefreshToken() {
		auth.refreshToken = res.GetRefreshToken()
		auth.changed = true
	}
	auth.accessToken = res.GetAccessToken()

	return nil
}

func (auth *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", auth.accessToken)
}

func (interceptor *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if apitools.RequiresAuth(method) {
			interceptor.log.Debugf("--> unary: %s (with auth)", method)
			return invoker(interceptor.attachToken(ctx), method, req, reply, cc, opts...)
		}

		if method != "/woodhouse.api.v1.clients.AuthService/Ping" {
			interceptor.log.Debugf("--> unary: %s (no auth)", method)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (interceptor *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		if apitools.RequiresAuth(method) {
			interceptor.log.Debugf("--> stream: %s (with auth)", method)
			return streamer(interceptor.attachToken(ctx), desc, cc, method, opts...)
		}

		interceptor.log.Debugf("--> stream: %s (no auth)", method)
		return streamer(ctx, desc, cc, method, opts...)
	}
}
