package version

import (
	_ "embed"
	"fmt"
	"runtime/debug"
)

//go:generate sh -c "printf %s $(git describe --tags) > version"
//go:generate sh -c "git status --porcelain > status"

var (
	//go:embed version
	tag string

	//go:embed status
	status string

	buildInfo string
)

func Version() string {
	v := tag
	if status != "" {
		v += "-dirty"
	}
	return fmt.Sprintf("%s %s", v, buildInfo)
}

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("No build info")
		return
	}

	var goos, goarch string
	for _, s := range info.Settings {
		switch s.Key {
		case "GOOS":
			goos = s.Value
		case "GOARCH":
			goarch = s.Value
		}
	}

	buildInfo = fmt.Sprintf("%s/%s", goos, goarch)
}
