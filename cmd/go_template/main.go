package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"charm.land/fang/v2"
	"github.com/spf13/cobra"
	"go.jacobcolvin.com/x/cobras/log"
	"go.jacobcolvin.com/x/cobras/profile"
	"go.jacobcolvin.com/x/version"
)

const appName = "go_template"

// ErrLogHandler indicates an error occurred while creating a log handler.
var ErrLogHandler = errors.New("create log handler")

func main() {
	err := fang.Execute(
		context.Background(),
		newRootCmd(),
		fang.WithVersion(version.GetVersion()),
	)
	if err != nil {
		os.Exit(1)
	}
}

// newRootCmd builds the root [*cobra.Command] for the go_template CLI. Logging
// and profiling are configured from persistent flags in PersistentPreRunE, so
// every subcommand shares the same setup.
func newRootCmd() *cobra.Command {
	logCfg := log.NewConfig()
	profileCfg := profile.NewConfig()

	cmd := &cobra.Command{
		Use:          appName,
		Short:        "A template for my Go projects.",
		SilenceUsage: true,
		Version:      version.GetVersion(),
	}

	logCfg.RegisterFlags(cmd.PersistentFlags())
	logCfg.MustRegisterCompletions(cmd)
	profileCfg.RegisterFlags(cmd.PersistentFlags())
	profileCfg.MustRegisterCompletions(cmd)

	profiler := profileCfg.NewProfiler()

	cmd.PersistentPreRunE = func(cc *cobra.Command, _ []string) error {
		err := profiler.Start()
		if err != nil {
			return fmt.Errorf("start profiler: %w", err)
		}

		h, err := logCfg.NewHandler(cc.ErrOrStderr())
		if err != nil {
			return fmt.Errorf("%w: %w", ErrLogHandler, err)
		}

		slog.SetDefault(slog.New(h))

		return nil
	}

	cmd.PersistentPostRunE = func(_ *cobra.Command, _ []string) error {
		err := profiler.Stop()
		if err != nil {
			return fmt.Errorf("stop profiler: %w", err)
		}

		return nil
	}

	cmd.AddCommand(newHelloCmd())
	cmd.AddCommand(newVersionCmd())

	return cmd
}

// newHelloCmd returns a minimal example subcommand. Replace it with the
// commands your project actually needs.
func newHelloCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hello",
		Short: "Print a greeting",
		RunE: func(cc *cobra.Command, _ []string) error {
			slog.Debug("saying hello")

			return Hello(cc.OutOrStdout())
		},
	}
}

// newVersionCmd returns a command that prints full build information.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		RunE: func(cc *cobra.Command, _ []string) error {
			cc.Println(version.Get().String())

			return nil
		},
	}
}

// Hello writes a greeting to w.
func Hello(w io.Writer) error {
	_, err := io.WriteString(w, "Hello World!")
	if err != nil {
		return fmt.Errorf("write greeting: %w", err)
	}

	return nil
}
