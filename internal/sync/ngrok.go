package sync

import (
	"encoding/json"
	"io"
	"net/http"
)

const (
	ngrokAPIAddr string = "http://ngrok:4040/api/tunnels"
)

// GetNgrokTunnelURL
func GetNgrokTunnelURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tunnels := NgrokTunnelResponse{}
	jsonErr := json.Unmarshal(body, &tunnels)
	if jsonErr != nil {
		return "", err
	}

	var publicURL string
	for _, tunnel := range tunnels.Tunnels {
		publicURL = tunnel.Public_url
	}

	return publicURL, nil
}
