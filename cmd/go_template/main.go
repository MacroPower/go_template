package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/MacroPower/go_template/internal/log"

	"github.com/alecthomas/kong"
)

const appName = "go_template"

var cli struct {
	Log struct {
		Level  string `help:"Log level." default:"info"`
		Format string `help:"Log format. One of: [logfmt, json]" default:"logfmt"`
	} `prefix:"log." embed:""`
}

func main() {
	cliCtx := kong.Parse(&cli, kong.Name(appName))

	logLevel := &log.AllowedLevel{}
	if err := logLevel.Set(cli.Log.Level); err != nil {
		cliCtx.FatalIfErrorf(err)
	}

	logFormat := &log.AllowedFormat{}
	if err := logFormat.Set(cli.Log.Format); err != nil {
		cliCtx.FatalIfErrorf(err)
	}

	logger := log.New(&log.Config{
		Level:  logLevel,
		Format: logFormat,
	})

	err := log.Info(logger).Log("msg", fmt.Sprintf("Starting %s", appName))
	cliCtx.FatalIfErrorf(err)

	sb := strings.Builder{}
	Hello(&sb)
	cliCtx.Printf("%s", sb.String())
}

func Hello(r io.Writer) {
	_, err := r.Write([]byte("Hello World!"))
	if err != nil {
		panic(err)
	}
}
