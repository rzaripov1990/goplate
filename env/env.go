package env

import (
	"time"
)

type (
	BaseConfig struct {
		Port        string      `name:"PORT" default:"8080"`
		Environment string      `name:"ENVIRONMENT" default:"dev" required:"true"`
		App         Application `name:"APP"`
		Log         Log         `name:"LOG"`
		Http        Http
	}

	Application struct {
		Name    string `name:"NAME" required:"true"`
		Version string `name:"VERSION" default:"1.0.0"`
	}

	Log struct {
		Level string `name:"LEVEL" default:"DEBUG"`
	}

	Http struct {
		BaseUrl             string        `name:"BASE_URL"`
		BaseUiUrl           string        `name:"BASE_UI_URL"`
		OwnServiceUrl       string        `name:"OWN_SERVICE_URL"`
		DefaultTimeout      time.Duration `name:"DEFAULT_TIMEOUT" default:"60s"`
		ReadTimeout         time.Duration `name:"READ_TIMEOUT" default:"60s"`
		WriteTimeout        time.Duration `name:"WRITE_TIMEOUT" default:"60s"`
		MaxIdleConns        int           `name:"HTTP_MAX_IDLE_CONNS" default:"10"`
		MaxIdleConnsPerHost int           `name:"HTTP_MAX_IDLE_CONNS_PER_HOST" default:"10"`
	}
)
