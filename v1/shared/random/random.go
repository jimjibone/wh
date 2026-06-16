package random

import (
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/schollz/mnemonicode"
)

// GenerateRandom returns a securely generated random slice of bytes based
// on the supplied set. It will return an error if the system's secure random
// number generator fails to function correctly, in which case the caller should
// not continue.
// See https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func GenerateRandom(n int, set []byte) ([]byte, error) {
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(set))))
		if err != nil {
			return nil, err
		}
		ret[i] = set[num.Int64()]
	}
	return ret, nil
}

// GenerateRandomString returns a securely generated random slice of bytes.
// See GenerateRandom.
func GenerateRandomBytes(n int) ([]byte, error) {
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(255)))
		if err != nil {
			return nil, err
		}
		ret[i] = byte(num.Int64())
	}
	return ret, nil
}

// GenerateRandomString returns a securely generated random string.
// See GenerateRandom.
func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret, err := GenerateRandom(n, []byte(letters))
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

// GenerateRandomPin returns a securely generated random pin.
// See GenerateRandom.
func GenerateRandomPin(n int) (string, error) {
	const letters = "0123456789"
	ret, err := GenerateRandom(n, []byte(letters))
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

// GenerateRandomName returns mnemoicoded random name containing n words
// separated by hyphens.
func GenerateRandomName(n int) (string, error) {
	bs, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	result := mnemonicode.EncodeWordList(nil, bs)
	return strings.Join(result, "-"), nil
}
