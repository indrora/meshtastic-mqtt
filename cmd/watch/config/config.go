package config

import (
	"io"

	"github.com/pelletier/go-toml/v2"
)

type Channel struct {
	Name string    `toml:"name"`
	Key  CryptoKey `toml:"key"`
}

type Config struct {
	Channels  []Channel `toml:"channels"`
	DeviceIDs []string  `toml:"device_ids"`
}

func Load(r io.Reader) (*Config, error) {
	var cfg Config
	decoder := toml.NewDecoder(r)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
