// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period time.Duration `config:"period"`
	Redis []string `config:"redis"`
}

var DefaultConfig = Config{
	Period: 1 * time.Second,
	Redis:  []string{"192.168.33.10:16379"},
}
