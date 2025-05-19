package main

import (
	"log"

	"github.com/frolmr/GophKeeper/internal/client/app"
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

	gk, err := app.NewApplication()
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	cmd.SetApp(gk)
	cmd.Execute()

	if err := gk.Close(); err != nil {
		log.Fatalf("app close failure: %v", err)
	}
}
