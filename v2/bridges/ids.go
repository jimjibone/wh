package bridges

import (
	"errors"
	"strings"
)

// DeviceID builds a device ID following the convention "<protocol><unique-id>"
// in lowercase [a-z0-9], e.g. DeviceID("zigbee", "0x00158D00045C089F") is
// "zigbee00158d00045c089f". protocol names the bridge's protocol or product
// family ("zigbee", "shelly", "sonos", ...). raw should be a hardware ID (MAC,
// IEEE address, serial) where one exists, otherwise the most stable ID the
// protocol offers. A leading "0x" is dropped and every character outside
// [a-z0-9] is removed after lowercasing.
func DeviceID(protocol, raw string) (string, error) {
	p := clean(protocol)
	if p == "" || p != protocol {
		return "", errors.New("device id: protocol must be non-empty lowercase [a-z0-9]")
	}
	r := strings.ToLower(strings.TrimSpace(raw))
	r = strings.TrimPrefix(r, "0x")
	r = clean(r)
	if r == "" {
		return "", errors.New("device id: unique id has no [a-z0-9] characters")
	}
	return p + r, nil
}

func clean(s string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(s) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			b.WriteRune(c)
		}
	}
	return b.String()
}
