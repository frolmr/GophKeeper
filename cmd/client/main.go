package main

import (
	cmd "github.com/frolmr/GophKeeper/internal/client/commands"
	"github.com/frolmr/GophKeeper/pkg/buildinfo"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

func main() {
	buildinfo.CurrentBuild = buildinfo.BuildInfo{
		Version:   buildVersion,
		BuildDate: buildDate,
	}

	cmd.Execute()
}
