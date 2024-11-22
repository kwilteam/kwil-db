package version

import (
	"github.com/kwilteam/kwil-db/app/shared/display"
)

func Example_versionInfo_text() {
	display.Print(&respVersionInfo{
		Info: &versionInfo{
			Version:   "0.0.0",
			GitCommit: "00000000",

			BuildTime:  "0001-01-01T00:00:00Z",
			APIVersion: "0.0.0",
			GoVersion:  "unknown",
			Os:         "unknown",
			Arch:       "unknown",
		},
	}, nil, "text")
	//display.PrettyPrint(msg, "text")
	// Output:
	//  Version:	0.0.0
	//  Git commit:	00000000
	//  Built:	0001-01-01T00:00:00Z
	//  API version:	0.0.0
	//  Go version:	unknown
	//  OS/Arch:	unknown/unknown
}

func Example_versionInfo_json() {
	display.Print(&respVersionInfo{
		Info: &versionInfo{
			Version:    "0.0.0",
			GitCommit:  "00000000",
			BuildTime:  "0001-01-01T00:00:00Z",
			APIVersion: "0.0.0",
			GoVersion:  "unknown",
			Os:         "unknown",
			Arch:       "unknown",
		},
	}, nil, "json")
	// Output:
	// {
	//   "result": {
	//     "version": "0.0.0",
	//     "git_commit": "00000000",
	//     "build_time": "0001-01-01T00:00:00Z",
	//     "api_version": "0.0.0",
	//     "go_version": "unknown",
	//     "os": "unknown",
	//     "arch": "unknown"
	//   },
	//   "error": ""
	// }
}
