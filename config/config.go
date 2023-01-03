package config

type Config struct {
	// web
	Addr string

	// server monitoring
	MetricsEnabled bool
}

var config Config

func LoadConfig(c *Config) {
	config = *c
}

func GetConfig() *Config {
	return &config
}
