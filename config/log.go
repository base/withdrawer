// config/log.go
package config

import "flag"

var LogLevel string

func init() {
    flag.StringVar(&LogLevel, "log-level", "info", "set log level (debug, info, warn, error)")
}
