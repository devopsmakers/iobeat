// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

// Config - Type for IObeat Config
type Config struct {
	Period time.Duration `config:"period"`
	Disks  *[]string     `config:"disks"`
}

// DefaultConfig - setup a default config if none exists
var DefaultConfig = Config{
	Period: 5 * time.Second,
	Disks:  nil,
}
