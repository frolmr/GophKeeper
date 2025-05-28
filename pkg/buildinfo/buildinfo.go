package buildinfo

type BuildInfo struct {
	Version   string
	BuildDate string
}

var CurrentBuild BuildInfo
