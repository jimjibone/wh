package auth

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jimjibone/woodhouse-4/cmd/woodhouse-core/internal"
	"github.com/jimjibone/woodhouse-4/cmd/woodhouse-core/internal/jsonfile"
	"github.com/jimjibone/woodhouse-4/shared/paths"
)

const (
	refreshTokenDuration = 30 * 24 * time.Hour
	accessTokenDuration  = 15 * time.Minute
)

type BridgeAuth struct {
	wg             sync.WaitGroup
	close          func()
	mu             sync.RWMutex
	storageEnabled bool
	storagePath    string
	refreshSecret  string
	accessSecret   string
	changed        bool
}

func NewBridgeAuth(storageEnabled bool, storagePath string) (*BridgeAuth, error) {
	storagePath = paths.AbsPathify(storagePath)

	ctx, cancel := context.WithCancel(context.Background())
	ba := &BridgeAuth{
		close:          cancel,
		storageEnabled: storageEnabled,
		storagePath:    storagePath,
	}

	// Load the previous store from file.
	err := ba.loadStore(storagePath)
	if err != nil {
		return nil, err
	}

	// Generate refresh/access token keys if not already defined.
	if ba.refreshSecret == "" {
		key, err := internal.GenerateRandomString(64)
		if err != nil {
			return nil, fmt.Errorf("failed to generate refresh secret: %w", err)
		}
		ba.refreshSecret = key
		ba.changed = true
	}
	if ba.accessSecret == "" {
		key, err := internal.GenerateRandomString(64)
		if err != nil {
			return nil, fmt.Errorf("failed to generate access secret: %w", err)
		}
		ba.accessSecret = key
		ba.changed = true
	}

	// Save an empty version of the store if the file doesn't exist.
	if _, err := os.Stat(storagePath); errors.Is(err, fs.ErrNotExist) {
		err = ba.saveStore(storagePath)
		if err != nil {
			return nil, err
		}
	}

	// Periodically save the store if it has changed.
	ba.wg.Add(1)
	go func() {
		defer ba.wg.Done()
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ba.mu.RLock()
				if ba.changed {
					ba.changed = false
					err := ba.saveStore(storagePath)
					if err != nil {
						log.Fatalf("failed to save device store: %s", err)
					}
				}
				ba.mu.RUnlock()
			}
		}
	}()

	return ba, nil
}

func (ba *BridgeAuth) Close() error {
	ba.close()
	ba.wg.Wait()
	return ba.saveStore(ba.storagePath)
}

func (ba *BridgeAuth) loadStore(filename string) error {
	if ba.storageEnabled {
		store := struct {
			RefreshSecret string `json:"refresh-secret"`
			AccessSecret  string `json:"access-secret"`
		}{}

		err := jsonfile.LoadFile(&store, filename)
		if err != nil {
			return err
		}

		ba.refreshSecret = store.RefreshSecret
		ba.accessSecret = store.AccessSecret
	}

	return nil
}

func (ba *BridgeAuth) saveStore(filename string) error {
	if ba.storageEnabled {
		store := struct {
			RefreshSecret string `json:"refresh-secret"`
			AccessSecret  string `json:"access-secret"`
		}{
			RefreshSecret: ba.refreshSecret,
			AccessSecret:  ba.accessSecret,
		}
		return jsonfile.SaveFile(store, filename)
	}
	return nil
}

type TokenAllocation struct {
	ID             string
	RefreshExpires time.Time
}

type TokenDetails struct {
	AccessToken    string
	RefreshToken   string
	AccessUUID     string
	RefreshUUID    string
	AccessExpires  time.Time
	RefreshExpires time.Time
}

type RefreshTokenClaims struct {
	jwt.RegisteredClaims
	RefreshUUID string `json:"refresh_uuid"`
	BridgeID    string `json:"bridge_id"`
}

type AccessTokenClaims struct {
	jwt.RegisteredClaims
	AccessUUID string `json:"access_uuid"`
	BridgeID   string `json:"bridge_id"`
}

func (manager *BridgeAuth) GenerateTokens(id string) (*TokenDetails, error) {
	u1, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %v", err)
	}
	u2, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %v", err)
	}

	td := &TokenDetails{}
	td.AccessExpires = time.Now().Add(accessTokenDuration)
	td.AccessUUID = u1.String()
	td.RefreshExpires = time.Now().Add(refreshTokenDuration)
	td.RefreshUUID = u2.String()

	// Create Access Token
	atClaims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(td.AccessExpires),
		},
		AccessUUID: td.AccessUUID,
		BridgeID:   id,
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(manager.accessSecret))
	if err != nil {
		return nil, err
	}

	// Create Refresh Token
	rtClaims := RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(td.RefreshExpires),
		},
		RefreshUUID: td.RefreshUUID,
		BridgeID:    id,
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(manager.refreshSecret))
	if err != nil {
		return nil, err
	}

	// Add the new token allocation.
	// manager.tokenAllocations[td.RefreshUUID] = TokenAllocation{
	// 	Username:       bridge.Username,
	// 	RefreshExpires: td.RefreshExpires,
	// }

	return td, nil
}

// func (manager *BridgeAuth) RevokeToken(refreshUUID string) {
// 	delete(manager.tokenAllocations, refreshUUID)
// }

// func (manager *BridgeAuth) Generate(bridge *User) (string, error) {
// 	claims := AccessTokenClaims{
// 		StandardClaims: jwt.StandardClaims{
// 			ExpiresAt: time.Now().Add(manager.accessTokenDuration).Unix(),
// 		},
// 		Username: bridge.Username,
// 		Role:     bridge.Role,
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	return token.SignedString([]byte(manager.accessSecretKey))
// }

func (manager *BridgeAuth) VerifyRefreshToken(refreshToken string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		refreshToken,
		&RefreshTokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected refresh token signing method")
			}

			return []byte(manager.refreshSecret), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok {
		return nil, fmt.Errorf("invalid refresh token claims")
	}

	// Check that the refresh UUID is in the allocated (allowed) list.
	// if _, found := manager.tokenAllocations[claims.RefreshUUID]; !found {
	// 	return nil, fmt.Errorf("refresh token revoked")
	// }

	return claims, nil
}

func (manager *BridgeAuth) VerifyAccessToken(accessToken string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&AccessTokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected access token signing method")
			}

			return []byte(manager.accessSecret), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok {
		return nil, fmt.Errorf("invalid access token claims")
	}

	return claims, nil
}
