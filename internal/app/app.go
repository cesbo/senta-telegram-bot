package app

import (
	"fmt"
	"log"
	"os"
	"sentabot/internal/config"
	"sentabot/internal/tlgbot"
)

type App struct {
	path string
}

func NewApp(path string) (*App, error) {
	a := &App{
		path: path,
	}

	return a, nil
}

func (a *App) Start() error {
	log.Println("Start server")

	if _, err := os.Stat(a.path); os.IsNotExist(err) {
		return fmt.Errorf("config file not found")
	}

	if err := config.LoadConfig(a.path); err != nil {
		return err
	}

	tlgbot.Pool()

	return nil
}
