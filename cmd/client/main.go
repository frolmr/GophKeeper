package main

import (
	"log"

	"github.com/frolmr/GophKeeper/internal/client/app"
	cmd "github.com/frolmr/GophKeeper/internal/client/commands"
	"github.com/frolmr/GophKeeper/internal/client/config"
	"github.com/frolmr/GophKeeper/internal/client/crypto"
	"github.com/frolmr/GophKeeper/internal/client/storage"
	"github.com/frolmr/GophKeeper/pkg/buildinfo"
)

const (
	appName = "GophKeeper"
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

	cfg := config.NewConfig(appName)

	cryptoService := crypto.NewCryptoService()

	localStorate, err := storage.NewLocalStorage()
	if err != nil {
		log.Fatal("failed to initialize application dir")
	}

	gk := app.NewApplication(cfg, localStorate, cryptoService)

	cmd.SetApp(gk)
	cmd.Execute()
}
