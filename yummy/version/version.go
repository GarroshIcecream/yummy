package version

import "runtime/debug"

var Version = "dev"

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	mainVersion := info.Main.Version
	if mainVersion == "" || mainVersion == "(devel)" {
		return
	}

	Version = mainVersion
}
