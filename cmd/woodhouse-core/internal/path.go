package internal

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func AbsPathify(parts ...string) string {
	inPath := filepath.Join(parts...)

	// log.Println("DEBUG: Trying to resolve absolute path to", inPath)

	if inPath == "$HOME" || strings.HasPrefix(inPath, "$HOME"+string(os.PathSeparator)) {
		inPath = UserHomeDir() + inPath[5:]
	}

	if inPath == "~" || strings.HasPrefix(inPath, "~"+string(os.PathSeparator)) {
		inPath = UserHomeDir() + inPath[1:]
	}

	if strings.HasPrefix(inPath, "$") {
		end := strings.Index(inPath, string(os.PathSeparator))

		var value, suffix string
		if end == -1 {
			value = os.Getenv(inPath[1:])
		} else {
			value = os.Getenv(inPath[1:end])
			suffix = inPath[end:]
		}

		inPath = value + suffix
	}

	if filepath.IsAbs(inPath) {
		return filepath.Clean(inPath)
	}

	p, err := filepath.Abs(inPath)
	if err == nil {
		return filepath.Clean(p)
	}

	log.Println("ERROR: Couldn't discover absolute path")
	log.Println("ERROR: ", err)
	return ""
}

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
