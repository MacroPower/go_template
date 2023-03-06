package version

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/MacroPower/go_template/internal/log"
)

var (
	Version   string // Set via ldflags.
	Branch    string
	BuildUser string
	BuildDate string

	Revision  = getRevision()
	GoVersion = runtime.Version()
	GoOS      = runtime.GOOS
	GoArch    = runtime.GOARCH
)

// LogInfo logs version, branch and revision.
func LogInfo(logger log.Logger) error {
	if err := log.Info(logger).Log(
		"msg", "info",
		"version", Version,
		"branch", Branch,
		"revision", Revision,
	); err != nil {
		return fmt.Errorf("log error: %w", err)
	}

	return nil
}

// LogBuildContext logs goVersion, platform, buildUser and buildDate.
func LogBuildContext(logger log.Logger) error {
	if err := log.Info(logger).Log(
		"msg", "build context",
		"go", GoVersion,
		"platform", GoOS+"/"+GoArch,
		"buildUser", BuildUser,
		"buildDate", BuildDate,
	); err != nil {
		return fmt.Errorf("log error: %w", err)
	}

	return nil
}

func getRevision() string {
	rev := "unknown"

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return rev
	}

	modified := false
	for _, v := range buildInfo.Settings {
		switch v.Key {
		case "vcs.revision":
			rev = v.Value
		case "vcs.modified":
			if v.Value == "true" {
				modified = true
			}
		}
	}
	if modified {
		return rev + "-dirty"
	}

	return rev
}
