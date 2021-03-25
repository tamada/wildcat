package wildcat

import (
	"github.com/tamada/wildcat/errors"
)

// Config is the configuration object for counting.
type Config struct {
	ignore Ignore
	opts   *ReadOptions
	ec     *errors.Center
}

// NewConfig creates an instance of Config.
func NewConfig(ignore Ignore, opts *ReadOptions, ec *errors.Center) *Config {
	return &Config{ignore: ignore, opts: opts, ec: ec}
}

func (config *Config) updateOpts(newOpts *ReadOptions) *Config {
	return NewConfig(config.ignore, newOpts, config.ec)
}

func (config *Config) updateIgnore(newIgnore Ignore) *Config {
	return NewConfig(newIgnore, config.opts, config.ec)
}

// IsIgnore checks given line is the ignored file or not.
func (config *Config) IsIgnore(line string) bool {
	return config.ignore != nil && config.ignore.IsIgnore(line)
}
