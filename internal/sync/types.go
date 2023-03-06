package sync

// WebhookOptions holds generic webhook modification parameters
type WebhookOptions struct {
	Provider            string
	UseSecret           bool
	SecretName          string
	SecretNamespace     string
	SecretValues        string
	Owner               string
	Repository          string
	Url                 string
	OldUrl              string
	Token               string
	Cleanup             bool
	KubeInClusterConfig bool
	Restart             bool
}

// NgrokTunnelResponse describes the response from the ngrok api
type NgrokTunnelResponse struct {
	Tunnels []NgrokTunnelDefinition `json:"tunnels"`
}

// NgrokTunnelDefinition describes a singleton ngrok tunnel definition
type NgrokTunnelDefinition struct {
	Name       string                 `json:"name"`
	ID         string                 `json:"id"`
	Uri        string                 `json:"uri"`
	Public_url string                 `json:"public_url"`
	Proto      string                 `json:"proto"`
	Config     map[string]interface{} `json:"config"`
	Metrics    map[string]interface{} `json:"metrics"`
}
