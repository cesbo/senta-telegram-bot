package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"sentabot/internal/app"
	"sentabot/internal/help"
)

func start(path string) error {
	a, err := app.NewApp(path)
	if err != nil {
		return err
	}

	if err := a.Start(); err != nil {
		return err
	}

	return nil
}

func main() {
	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "", "help", "-h", "--help", "-help":
		help.PrintUsage()

	case "version", "-v", "--version", "-version":
		help.PrintVersion()

	default:
		if strings.HasPrefix(cmd, "-") {
			fmt.Println("Unknown argument:", cmd)
			fmt.Println("More information:", os.Args[0], "help")
			os.Exit(1)
		}

		if err := start(os.Args[1]); err != nil {
			log.Fatal(err)
		}
	}
}
