package v1

import (
	"net/url"
	"strconv"
)

func GetPortOrDefaultFromScheme(url *url.URL) int {
	port, err := strconv.Atoi(url.Port())
	if err == nil {
		return port
	}

	switch url.Scheme {
	case "http":
		return 80
	case "https":
		return 443
	default:
		return 80
	}
}
