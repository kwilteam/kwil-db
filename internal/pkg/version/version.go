package version

import (
	"fmt"
	"runtime/debug"
	"time"
)

var (
	KwilVersion = "0.5.1-pre" // precursor to 0.5.1+release
	Build       = vcsInfo()
)

func init() {
	if Build != nil {
		KwilVersion += "+" + Build.Revision
		if Build.Dirty {
			KwilVersion += ".dirty"
		}
	}
}

type BuildInfo struct {
	GoVersion string
	Revision  string
	RevTime   time.Time
	Dirty     bool
}

func vcsInfo() *BuildInfo {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	buildInfo := new(BuildInfo)
	for _, bs := range bi.Settings {
		switch bs.Key {
		case "vcs.revision":
			buildInfo.Revision = bs.Value
		case "vcs.time":
			revtime, err := time.Parse(time.RFC3339, bs.Value)
			if err != nil {
				fmt.Printf("invalid vcs.time %v: %v", bs.Value, err)
				continue
			}
			buildInfo.RevTime = revtime
		case "vcs.modified":
			buildInfo.Dirty = bs.Value == "true"
		}
	}
	const revLen = 9
	if len(buildInfo.Revision) > revLen {
		buildInfo.Revision = buildInfo.Revision[:revLen]
	}
	return buildInfo
}
