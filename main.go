package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/devopsmakers/iobeat/beater"
)

func main() {
	err := beat.Run("iobeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
