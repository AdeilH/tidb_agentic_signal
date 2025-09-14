package config

type Config struct {
	KimiKey string
	DBDSN   string
}

func Load() (*Config, error) {
	return &Config{
		KimiKey: "test-key",
		DBDSN:   ":memory:",
	}, nil
}
