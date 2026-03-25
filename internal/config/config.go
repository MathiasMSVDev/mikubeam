package config

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/creasty/defaults"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	ProxiesFile    string `toml:"proxies_file" default:"data/proxies.txt"`
	UserAgentsFile string `toml:"user_agents_file" default:"data/uas.txt"`
	ServerPort     int    `toml:"server_port" default:"3000"`
	AllowedOrigin  string `toml:"allowed_origin" default:"http://localhost:5173"`
}

func Default() Config {
	var c Config
	_ = defaults.Set(&c)
	return c
}

func Load(path string) (Config, error) {
	c := Default()

	if path != "" {
		b, err := os.ReadFile(path)
		if err == nil {
			_ = toml.Unmarshal(b, &c)
		}
	}

	// Environment variable overrides — useful for Docker / Dokploy deployments.
	// Priority: env var > config file > compiled default.

	// PORT is the conventional cloud env var; SERVER_PORT is the project-specific one.
	// SERVER_PORT wins over PORT when both are set.
	if v := os.Getenv("PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.ServerPort = n
		}
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.ServerPort = n
		}
	}

	// ALLOWED_ORIGIN controls the CORS allowed-origins header.
	// Use "*" to allow all origins when the server itself serves the web client.
	if v := os.Getenv("ALLOWED_ORIGIN"); v != "" {
		c.AllowedOrigin = v
	}

	// PROXIES_FILE / USER_AGENTS_FILE let you point to files outside the default
	// data/ directory, e.g. when using a Docker bind-mount on a different path.
	if v := os.Getenv("PROXIES_FILE"); v != "" {
		c.ProxiesFile = v
	}
	if v := os.Getenv("USER_AGENTS_FILE"); v != "" {
		c.UserAgentsFile = v
	}

	return c, nil
}

func ResolvePath(baseDir, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	if baseDir == "" {
		wd, _ := os.Getwd()
		baseDir = wd
	}
	return filepath.Join(baseDir, p)
}
