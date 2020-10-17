package fluid

import "runtime"
import "fmt"
import "github.com/golang/glog"

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
		versionStr = "v" + version
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

// Print version info directly by command
func PrintVersion(short bool){
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
	return
}

// Print version info in log when start
func LogVersion(){
	v := getVersion()
	glog.Infof("BuildDate: %s\n", v.BuildDate)
	glog.Infof("GitCommit: %s\n", v.GitCommit)
	glog.Infof("GitTreeState: %s\n", v.GitTreeState)
	if v.GitTag != "" {
		glog.Infof("GitTag: %s\n", v.GitTag)
	}
	glog.Infof("GoVersion: %s\n", v.GoVersion)
	glog.Infof("Compiler: %s\n", v.Compiler)
	glog.Infof("Platform: %s\n", v.Platform)
	return
}