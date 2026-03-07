package main

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func flagValue(name string) (string, bool) {
	for i, a := range os.Args {
		if a == name && i+1 < len(os.Args) {
			return os.Args[i+1], true
		}
		if strings.HasPrefix(a, name+"=") {
			return strings.TrimPrefix(a, name+"="), true
		}
	}
	return "", false
}

func intFlag(name string, fallback int) int {
	if v, ok := flagValue(name); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func timeSecondsFlag(name string) time.Duration {
	if v, ok := flagValue(name); ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 0
}
