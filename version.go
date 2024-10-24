package fluid

import (
	"fmt"
	"log"
	"runtime"
)

type Version struct {
	Version      string
	BuildDate    string
	GitCommit    string
	GitTag       string
	GitTreeState string
	GoVersion    string
	Compiler     string
	Platform     string
}

var (
	version      = "0.0.0"                // value from VERSION file
	buildDate    = "1970-01-01T00:00:00Z" // output from `date -u +'%Y-%m-%dT%H:%M:%SZ'`
	gitCommit    = ""                     // output from `git rev-parse HEAD`
	gitTag       = ""                     // output from `git describe --exact-match --tags HEAD` (if clean tree state)
	gitTreeState = ""                     // determined from `git status --porcelain`. either 'clean' or 'dirty'
)

func getVersion() Version {
	var versionStr string
	if gitCommit != "" && gitTag != "" && gitTreeState == "clean" {
		// if we have a clean tree state and the current commit is tagged,
		// this is an official release.
		versionStr = gitTag
	} else {
		// otherwise formulate a queryversion string based on as much metadata
		// information we have available.
		versionStr = version
		if len(gitCommit) >= 7 {
			versionStr += "+" + gitCommit[0:7]
			if gitTreeState != "clean" {
				versionStr += ".dirty"
			}
		} else {
			versionStr += "+unknown"
		}
	}
	return Version{
		Version:      versionStr,
		BuildDate:    buildDate,
		GitCommit:    gitCommit,
		GitTag:       gitTag,
		GitTreeState: gitTreeState,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// PrintVersion info directly by command
func PrintVersion(short bool) {
	v := getVersion()
	if short {
		fmt.Printf("version: %s\n", v.Version)
		return
	}
	fmt.Printf("  BuildDate: %s\n", v.BuildDate)
	fmt.Printf("  GitCommit: %s\n", v.GitCommit)
	fmt.Printf("  GitTreeState: %s\n", v.GitTreeState)
	if v.GitTag != "" {
		fmt.Printf("  GitTag: %s\n", v.GitTag)
	}
	fmt.Printf("  GoVersion: %s\n", v.GoVersion)
	fmt.Printf("  Compiler: %s\n", v.Compiler)
	fmt.Printf("  Platform: %s\n", v.Platform)
}

// LogVersion info in log when start
func LogVersion() {
	v := getVersion()
	log.Printf("BuildDate: %s\n", v.BuildDate)
	log.Printf("GitCommit: %s\n", v.GitCommit)
	log.Printf("GitTreeState: %s\n", v.GitTreeState)
	if v.GitTag != "" {
		log.Printf("GitTag: %s\n", v.GitTag)
	}
	log.Printf("GoVersion: %s\n", v.GoVersion)
	log.Printf("Compiler: %s\n", v.Compiler)
	log.Printf("Platform: %s\n", v.Platform)
}
