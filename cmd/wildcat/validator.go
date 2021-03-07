package main

import (
	"fmt"
	"strings"
)

func validateOptions(opts *options) error {
	return validateFormat(opts.cli.format)
}

func validateFormat(givenFormat string) error {
	availableFormats := []string{"default", "csv", "json", "xml"}
	format := strings.ToLower(givenFormat)
	return contains(format, availableFormats)
}

func contains(item string, items []string) error {
	for _, f := range items {
		if f == item {
			return nil
		}
	}
	return fmt.Errorf("%s: invalid resultant format", item)
}
