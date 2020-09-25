package version

import (
	"log"
	"runtime/debug"
)

var (
	// Version is the application version (set in release builds by the command line)
	Version = ""
	// IAM is the version of the IAM API client
	IAM = ""
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		log.Fatalf("Failed to read build info")
	}

	if Version == "" {
		Version = info.Main.Version
	}

	for _, mod := range info.Deps {
		switch mod.Path {
		case "github.com/netsoc/iam/client":
			IAM = mod.Version
		}
	}
}
