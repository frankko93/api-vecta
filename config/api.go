package config

import (
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type API struct {
	Name              string        `default:"go8_api"`
	Host              string        `default:"0.0.0.0"`
	Port              string        `default:"3080"`
	ReadHeaderTimeout time.Duration `split_words:"true" default:"60s"`
	GracefulTimeout   time.Duration `split_words:"true" default:"8s"`

	RequestLog bool `split_words:"true" default:"false"`
	RunSwagger bool `split_words:"true" default:"true"`
}

func NewAPI() API {
	var api API
	envconfig.MustProcess("NewAPI", &api)

	// Render (and other PaaS) injects PORT env var — use it as override
	if port := os.Getenv("PORT"); port != "" {
		api.Port = port
	}

	return api
}
