package main

import (
	"fmt"
	"os"

	"github.com/mayuanyuan/gitai/internal/gitai"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:     "gitai",
		Usage:    "AI-native version control system",
		Version:  "0.1.0-day2",
		Commands: []*cli.Command{},
	}

	// Register commands
	registerCommands(app)

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func registerCommands(app *cli.App) {
	app.Commands = append(app.Commands,
		&cli.Command{
			Name:   "init",
			Usage:  "Initialize .gitai/ in current git repository",
			Action: gitai.InitCmd,
		},
		&cli.Command{
			Name:      "track",
			Usage:     "Track a .prompt file",
			ArgsUsage: "<file.prompt>",
			Action:    gitai.TrackCmd,
		},
		&cli.Command{
			Name:  "status",
			Usage: "Show tracked files status",
			Action: func(c *cli.Context) error {
				return gitai.StatusCmd()
			},
		},
		&cli.Command{
			Name:      "show",
			Usage:     "Show detailed information about a tracked file",
			ArgsUsage: "<file.prompt>",
			Action:    gitai.ShowCmd,
		},
	)
}
