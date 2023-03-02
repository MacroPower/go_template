package version

import (
	"runtime"
	"runtime/debug"

	"github.com/MacroPower/go_template/internal/log"

	kitlog "github.com/go-kit/log"
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
func LogInfo(logger kitlog.Logger) {
	if err := log.Info(logger).Log(
		"msg", "info",
		"version", Version,
		"branch", Branch,
		"revision", Revision,
	); err != nil {
		panic(err)
	}
}

// LogBuildContext logs goVersion, platform, buildUser and buildDate.
func LogBuildContext(logger kitlog.Logger) {
	if err := log.Info(logger).Log(
		"msg", "build context",
		"go", GoVersion,
		"platform", GoOS+"/"+GoArch,
		"user", BuildUser,
		"date", BuildDate,
	); err != nil {
		panic(err)
	}
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
