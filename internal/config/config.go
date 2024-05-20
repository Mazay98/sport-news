package config

import (
	"errors"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jessevdk/go-flags"
	"go.sport-news/internal/environment"
	"os"
	"time"
)

//nolint:lll
type (
	Config struct {
		Env    environment.Env `yml:"env" env:"ENV" long:"env" description:"Environment application is running in" default:"local"`
		Parser Parser          `yml:"parser" env-namespace:"PARSER" namespace:"parser" group:"Parser options"`
		HTTP   Http            `yml:"http" env-namespace:"HTTP" namespace:"http" group:"Http options"`
		Logger Logger          `yml:"logger" env-namespace:"LOGGER"  namespace:"logger"       group:"Logger options"`
		Mongo  Mongo           `yml:"mongo" env-namespace:"MONGO"   namespace:"mongo"      group:"Mongodb options"`
	}
	Mongo struct {
		URL        string `yml:"url" env:"URL" long:"url" description:"Mongodb url" default:"mongodb://localhost:27017"  `
		Collection string `yml:"collection" env:"C_NAME" long:"collection-name" description:"Mongodb collection name" default:"sport-news"  `
	}
	Parser struct {
		Enable  int8          `yml:"enable" env:"ENABLE" long:"enable" description:"Enable parsing monde" default:"1"`
		URL     string        `yml:"url" env:"URL" long:"url" description:"Parser url" default:"https://www.htafc.com/api/incrowd"`
		Count   int           `yml:"count" env:"COUNT" long:"count"  description:"Count rows from url" default:"50"`
		JobTime time.Duration `yml:"time" env:"JOB_TIME" long:"job-time" description:"Job parser timer" default:"30s"`
	}
	Http struct {
		Port         int           `yml:"port" env:"PORT" long:"port" description:"" default:"8080"`
		ExternalPort int           `yml:"external_port" env:"EXTERNAL_PORT" long:"external_port" description:"" env-default:"8889"`
		WriteTimeout time.Duration `yml:"write_timeout" env:"WRITE_TIMEOUT" long:"write_timeout" description:"Write timeout" default:"100s"`
		ReadTimeout  time.Duration `yml:"read_timeout" env:"READ_TIMEOUT" long:"read_timeout" description:"Read timeout" default:"100s"`
		IdleTimeout  time.Duration `yml:"idle_timeout" env:"IDLE_TIMEOUT" long:"idle_timeout" description:"Idle timeout" default:"100s"`
	}
	Logger struct {
		Level string `env:"LEVEL" long:"level" description:"Log level to use; environment-base level is used when empty" `
	}
)

// MustLoad reads flags and envs and returns Config
// that corresponds to the values read.
func MustLoad() *Config {
	var config Config
	if _, err := flags.Parse(&config); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && flagsErr.Type == flags.ErrHelp {
			panic("help")
		}
		panic("failed to parse config")
	}

	return &config
}

// MustLoadFromYAML read cfg from YAML file
func MustLoadFromYAML(configPath string) *Config {
	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}
