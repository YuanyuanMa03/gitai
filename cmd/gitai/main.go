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
		Version:  "1.0.0",
		Commands: []*cli.Command{},
	}

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
		&cli.Command{
			Name:      "diff",
			Usage:     "Compare two versions of a prompt file",
			ArgsUsage: "<file.prompt> [commit-a] [commit-b]",
			Action:    gitai.DiffCmd,
		},
		&cli.Command{
			Name:  "log",
			Usage: "Show commit history with token deltas",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:    "limit",
					Aliases: []string{"n"},
					Value:   10,
					Usage:   "Number of commits to show",
				},
			},
			Action: gitai.LogCmd,
		},
		&cli.Command{
			Name:  "cost",
			Usage: "Show cost analysis for tracked files",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "model",
					Aliases: []string{"m"},
					Value:   "gpt-4o",
					Usage:   "Model for cost calculation",
				},
				&cli.StringFlag{
					Name:    "currency",
					Aliases: []string{"c"},
					Value:   "USD",
					Usage:   "Preferred currency display (USD/CNY)",
				},
				&cli.StringFlag{
					Name:    "by",
					Value:   "file",
					Usage:   "Aggregate by: file, section, role",
				},
				&cli.BoolFlag{
					Name:    "user",
					Aliases: []string{"u"},
					Usage:   "Show cost breakdown by user (overrides --by)",
				},
				&cli.BoolFlag{
					Name:    "branch",
					Aliases: []string{"b"},
					Usage:   "Show cost breakdown by branch (overrides --by)",
				},
				&cli.IntFlag{
					Name:    "limit",
					Aliases: []string{"n"},
					Value:   50,
					Usage:   "Number of commits to analyze (for --user)",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Bool("user") {
					return gitai.CostByUserCmd(c)
				}
				if c.Bool("branch") {
					return gitai.CostByBranchCmd(c)
				}
				return gitai.CostCmd(c)
			},
		},
		&cli.Command{
			Name:  "price",
			Usage: "List model pricing information",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "currency",
					Aliases: []string{"c"},
					Value:   "USD",
					Usage:  "Show prices in currency (USD/CNY)",
				},
			},
			Action: gitai.PriceCmd,
		},
		&cli.Command{
			Name:  "config",
			Usage:  "Manage GitAI configuration",
			ArgsUsage: "[init|set|get|list|edit] [key] [value]",
			Action:    gitai.ConfigCmd,
		},
		&cli.Command{
			Name:  "review",
			Usage: "Review prompt files for quality and improvement suggestions",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "detailed",
					Aliases: []string{"d"},
					Usage:   "Show detailed suggestions for each file",
				},
				&cli.Float64Flag{
					Name:    "min-score",
					Aliases: []string{"m"},
					Value:   0,
					Usage:   "Only show files with score below this threshold",
				},
			},
			Action: gitai.ReviewCmd,
		},
		&cli.Command{
			Name:  "quality",
			Usage: "Show quick quality overview of all prompts",
			Action: gitai.QualityCmd,
		},
	)
}
