package config

type Config struct {
	Version    string                 `json:"version"`
	Properties map[string]interface{} `json:"properties"`
}
