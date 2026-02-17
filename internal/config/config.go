package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env"
	"github.com/creasty/defaults"
)

type Config struct {
	HostAddr    string `json:"hostAddr" yaml:"hostAddr" env:"RUN_ADDRESS" default:"localhost:8080"`
	DatabaseDSN string `json:"dbDSN" yaml:"dbDSN" env:"DATABASE_URI"`
	Key         string `json:"key" env:"KEY"`
}

func (c *Config) ParseConfig() error {
	if err := defaults.Set(c); err != nil {
		return err
	}

	flag.Func("a", "pass ip:port to run server", func(flagArgs string) error {
		return parseAddr(flagArgs, &c.HostAddr)
	})
	flag.StringVar(
		&c.DatabaseDSN,
		"d",
		"",
		"string to connect to database",
	)
	flag.StringVar(
		&c.Key,
		"k",
		"qwerty123",
		"key for JWT",
	)
	flag.Parse()

	if err := env.Parse(c); err != nil {
		return err
	}

	return nil
}

func parseAddr(args string, value *string) error {
	val := strings.Split(args, ":")
	if len(val) != 2 {
		return fmt.Errorf("host address must be ip:port")
	}

	*value = args

	return nil
}
