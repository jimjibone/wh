package auth

import (
	"context"
	"strings"

	"github.com/jimjibone/woodhouse-4/cmd/woodhouse-core/bridges"
	"github.com/jimjibone/woodhouse-4/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	ba                    *bridges.JWTManager
	authorisationRequired map[string]bool // bool indicates if auth required
}

func NewAuthInterceptor(ba *bridges.JWTManager) *AuthInterceptor {
	return &AuthInterceptor{
		ba: ba,
		authorisationRequired: map[string]bool{
			"/woodhouse.api.BridgeAuthService/Pair":    false,
			"/woodhouse.api.BridgeAuthService/Refresh": true,
		},
	}
}

// Unary returns a server interceptor function to authenticate and authorize unary RPC
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Debugf("--> unary: %s", info.FullMethod)

		ctx2, err := i.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(ctx2, req)
	}
}

// Stream returns a server interceptor function to authenticate and authorize stream RPC
func (i *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		log.Debugf("--> stream: %s", info.FullMethod)

		ctx2, err := i.authorize(stream.Context(), info.FullMethod)
		if err != nil {
			return err
		}

		return handler(srv, newWrappedStream(stream, ctx2))
	}
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func newWrappedStream(s grpc.ServerStream, ctx context.Context) grpc.ServerStream {
	return &wrappedStream{s, ctx}
}

func (ws *wrappedStream) Context() context.Context {
	return ws.ctx
}

func (ai *AuthInterceptor) authorize(ctx context.Context, method string) (context.Context, error) {
	// gRPC-Web requests don't have a / prefix.
	if !strings.HasPrefix(method, "/") {
		method = "/" + method
	}

	authRequired, ok := ai.authorisationRequired[method]
	if !ok {
		// method denied by default
		log.Warnf("requested method has no accessible roles: %s", method)
		return nil, status.Errorf(codes.Internal, "requested method has no accessible roles")
	}

	if !authRequired {
		// everyone can access this
		return ctx, nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 || len(values[0]) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	// if accessibleMethod.IsUserMethod() {
	// 	accessToken := values[0]
	// 	claims, err := i.userJWTManager.VerifyAccessToken(accessToken)
	// 	if err != nil {
	// 		return nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	// 	}

	// 	for _, role := range accessibleMethod.Roles {
	// 		if role == claims.Role {
	// 			return context.WithValue(ctx, "claims", claims), nil
	// 		}
	// 	}
	// } else if accessibleMethod.IsBridgeMethod() {
	accessToken := values[0]
	claims, err := ai.ba.VerifyAccessToken(accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

	return context.WithValue(ctx, bridges.AccessTokenClaimsType, claims), nil

	// for _, perm := range accessibleMethod.Perms {
	// 	for _, claimPerm := range claims.Perms {
	// 		if perm == perms.Perm(claimPerm) {
	// 			return context.WithValue(ctx, "claims", claims), nil
	// 		}
	// 	}
	// }
	// }

	// return nil, status.Error(codes.PermissionDenied, "no permission to access this RPC")
}
