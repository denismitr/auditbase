package rest

import "strings"

type Config struct {
	Port      string
	BodyLimit string
}

func ResolvePort(port string) string {
	if port == "" {
		port = "3000"
	}

	if !strings.HasPrefix(port, ":") {
		return ":" + port
	}

	return port
}
