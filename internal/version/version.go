package version

import "runtime/debug"

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	mainVersion := info.Main.Version
	if mainVersion != "" && mainVersion != "(devel)" {
		// bin built using `go install`
		Version = mainVersion
	}

	// Get build info from settings
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			if GitCommit == "unknown" {
				GitCommit = setting.Value
			}
		case "vcs.time":
			if BuildTime == "unknown" {
				BuildTime = setting.Value
			}
		case "vcs.modified":
			if setting.Value == "true" {
				GitCommit += "+dirty"
			}
		}
	}
}

func GetVersionInfo() string {
	return Version
}

func GetFullVersionInfo() string {
	return Version + " (commit: " + GitCommit + ", built: " + BuildTime + ")"
}
