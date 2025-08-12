package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Debug         bool   `env:"DEBUG" env-default:"false"`
	LogLevel      string `env:"LOG_LEVEL" env-default:"info"`
	ServerAddress string `env:"HOST" env-default:"localhost:9090"`

	Database struct {
		URL      string        `env:"URL" env-required:"true"`
		Timeout  time.Duration `env:"TIMEOUT" env-default:"5s"`
		PoolSize int           `env:"POOL_SIZE" env-default:"10"`
	} `env-prefix:"DB_"`

	Features []string `env:"FEATURES" env-separator:","`
}

// Loader loads configuration from environment variables and optional .env file
// Supports automatic reloading if watch is true
func NewConfig[T any](path string) (*T, error) {
	var cfg *T

	// Read configuration
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return cfg, nil
}

// func ExampleUsage() {
// 	//Missing required field
// 	_, err := NewConfig[AppConfig]("m.toml")
// 	fmt.Println(err)
// 	//err: environment variable "DB_URL" is not set

// 	//Invalid file path
// 	c, err := NewConfig[AppConfig]("m.toml")
// 	fmt.Println(err)
// 	fmt.Println(c)
// 	//err: failed to read config file

// 	fmt.Printf("Server: %s:%d\n", c.ServerAddress)
// 	fmt.Printf("DB Timeout: %v\n", c.Database.Timeout)
// 	fmt.Printf("Features: %v\n", c.Features)

// 	f, err := os.Create("postgre.toml")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer f.Close()

// 	err = toml.NewEncoder(f).Encode(&configStruct)
// 	if err != nil {
// 		panic(err)
// 	}
// }
