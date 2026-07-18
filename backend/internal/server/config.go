package server

import "fmt"

type Config struct {
	Port int `mapstructure:"port"`
}

func (c Config) Validate() error {
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("server: port value %v must be in [0,65535]", c.Port)
	}
	return nil
}
