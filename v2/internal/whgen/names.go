package main

import "strings"

// initialisms are written in upper case in Go names.
var initialisms = map[string]string{
	"id":  "ID",
	"url": "URL",
	"qr":  "QR",
	"xy":  "XY",
	"hs":  "HS",
}

// camel converts snake_case to CamelCase, e.g. "device_id" -> "DeviceID".
func camel(s string) string {
	var b strings.Builder
	for _, part := range strings.Split(s, "_") {
		if part == "" {
			continue
		}
		if up, ok := initialisms[part]; ok {
			b.WriteString(up)
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]) + part[1:])
	}
	return b.String()
}

// lowerCamel converts snake_case to lowerCamelCase, e.g. "set_schedule" ->
// "setSchedule".
func lowerCamel(s string) string {
	c := camel(s)
	for word, up := range initialisms {
		if strings.HasPrefix(c, up) && strings.HasPrefix(s, word) {
			return word + c[len(up):]
		}
	}
	return strings.ToLower(c[:1]) + c[1:]
}
