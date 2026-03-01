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
		Version:  "1.4.0",
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
		&cli.Command{
			Name:  "export",
			Usage: "Export tracked files to JSON or CSV",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "format",
					Aliases: []string{"f"},
					Value:   "json",
					Usage:   "Export format: json or csv",
				},
				&cli.StringFlag{
					Name:    "output",
					Aliases: []string{"o"},
					Value:   "",
					Usage:   "Output filename (default: gitai-export-TIMESTAMP.{json,csv})",
				},
				&cli.BoolFlag{
					Name:    "include-config",
					Aliases: []string{"c"},
					Usage:   "Include configuration in JSON export",
				},
			},
			Action: gitai.ExportCmd,
		},
		&cli.Command{
			Name:  "import",
			Usage: "Import tracked files from export file",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "input",
					Aliases: []string{"i"},
					Usage:   "Input file path (required)",
				},
				&cli.BoolFlag{
					Name:    "merge",
					Aliases: []string{"m"},
					Usage:   "Merge with existing tracked files",
				},
				&cli.BoolFlag{
					Name:    "force",
					Aliases: []string{"f"},
					Usage:   "Force overwrite existing entries",
				},
			},
			Action: gitai.ImportCmd,
		},
		&cli.Command{
			Name:      "search",
			Usage:     "Search across tracked prompt files",
			ArgsUsage: "<query>",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "role",
					Aliases: []string{"r"},
					Usage:   "Filter by role",
				},
				&cli.IntFlag{
					Name:    "max",
					Aliases: []string{"n"},
					Value:   0,
					Usage:   "Maximum results to show (0 = unlimited)",
				},
				&cli.IntFlag{
					Name:    "context",
					Aliases: []string{"C"},
					Value:   0,
					Usage:   "Number of context lines to show",
				},
				&cli.BoolFlag{
					Name:    "regex",
					Aliases: []string{"e"},
					Usage:   "Use regex pattern matching",
				},
				&cli.BoolFlag{
					Name:    "case-sensitive",
					Aliases: []string{"c"},
					Usage:   "Case-sensitive search",
				},
			},
			Action: gitai.SearchCmd,
		},
		&cli.Command{
			Name:      "find-role",
			Usage:     "Find files by role pattern",
			ArgsUsage: "<pattern>",
			Action:    gitai.FindByRoleCmd,
		},
		&cli.Command{
			Name:  "template",
			Usage: "Manage prompt templates",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "name",
					Aliases: []string{"n"},
					Usage:   "Template name",
				},
				&cli.StringFlag{
					Name:    "category",
					Aliases: []string{"c"},
					Usage:   "Template category",
				},
				&cli.StringFlag{
					Name:    "from",
					Aliases: []string{"f"},
					Usage:   "Create from existing file",
				},
				&cli.StringFlag{
					Name:    "description",
					Aliases: []string{"d"},
					Usage:   "Template description",
				},
				&cli.StringFlag{
					Name:    "output",
					Aliases: []string{"o"},
					Usage:   "Output file for 'use' command",
				},
				&cli.StringSliceFlag{
					Name:    "var",
					Usage:   "Variable substitutions (key=value)",
				},
			},
			Action: gitai.TemplateCmd,
		},
		&cli.Command{
			Name:  "tag",
			Usage: "Manage tags for tracked files",
			Subcommands: []*cli.Command{
				{
					Name:      "add",
					Usage:     "Add tags to a file",
					ArgsUsage: "<file> <tag1> [tag2 ...]",
					Action:    gitai.TagCmd,
				},
				{
					Name:      "remove",
					Aliases:   []string{"rm"},
					Usage:     "Remove tags from a file",
					ArgsUsage: "<file> <tag1> [tag2 ...]",
					Action:    gitai.TagCmd,
				},
				{
					Name:      "set",
					Usage:     "Set tags for a file (replaces existing)",
					ArgsUsage: "<file> <tag1> [tag2 ...]",
					Action:    gitai.TagCmd,
				},
				{
					Name:      "list",
					Aliases:   []string{"ls"},
					Usage:     "List tags for a file",
					ArgsUsage: "<file>",
					Action:    gitai.TagCmd,
				},
				{
					Name:      "search",
					Usage:     "Search files by tag",
					ArgsUsage: "<tag>",
					Action:    gitai.TagCmd,
				},
			},
			Action: gitai.TagCmd,
		},
		&cli.Command{
			Name:  "stats",
			Usage: "Show comprehensive statistics dashboard",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "model",
					Aliases: []string{"m"},
					Value:   "gpt-4o",
					Usage:   "Model for cost calculation",
				},
				&cli.BoolFlag{
					Name:    "verbose",
					Aliases: []string{"v"},
					Usage:   "Show verbose output",
				},
			},
			Action: gitai.StatsCmd,
		},
		&cli.Command{
			Name:  "trend",
			Usage: "Show token growth trend over time",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:    "days",
					Aliases: []string{"d"},
					Value:   30,
					Usage:   "Number of days to show",
				},
			},
			Action: gitai.TrendCmd,
		},
		&cli.Command{
			Name:  "compare",
			Usage: "Compare statistics between two points",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "a",
					Usage: "First commit or reference",
				},
				&cli.StringFlag{
					Name:  "b",
					Usage: "Second commit or reference",
				},
			},
			Action: gitai.CompareCmd,
		},
		&cli.Command{
			Name:  "commit",
			Usage: "Create a commit with intent tracking",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "message",
					Aliases: []string{"m"},
					Usage:   "Commit message",
				},
				&cli.BoolFlag{
					Name:    "edit",
					Aliases: []string{"e"},
					Usage:   "Open editor for commit message",
				},
				&cli.BoolFlag{
					Name:    "intent-only",
					Aliases: []string{"i"},
					Usage:   "Save intent without creating commit",
				},
				&cli.BoolFlag{
					Name:    "dry-run",
					Aliases: []string{"n"},
					Usage:   "Show intent without committing",
				},
			},
			Action: gitai.CommitCmd,
		},
		&cli.Command{
			Name:  "intent-log",
			Usage: "Show intent history",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:    "limit",
					Aliases: []string{"n"},
					Value:   20,
					Usage:   "Number of intents to show",
				},
				&cli.StringFlag{
					Name:    "category",
					Aliases: []string{"c"},
					Usage:   "Filter by category",
				},
			},
			Action: gitai.LogIntentsCmd,
		},
		&cli.Command{
			Name:      "impact",
			Usage:     "Analyze impact of changing a prompt",
			ArgsUsage: "<prompt-file>",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "verbose",
					Aliases: []string{"v"},
					Usage:   "Show detailed analysis",
				},
			},
			Action: gitai.ImpactCmd,
		},
		&cli.Command{
			Name:  "scan-deps",
			Usage: "Auto-scan and detect prompt dependencies",
			Action: gitai.ScanDependenciesCmd,
		},
		&cli.Command{
			Name:  "show-deps",
			Usage: "Show all tracked dependencies",
			Action: gitai.ShowDependenciesCmd,
		},
		&cli.Command{
			Name:  "feedback",
			Usage: "Record and manage feedback on commits/intents",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "reason",
					Aliases: []string{"r"},
					Usage:   "Reason for feedback",
				},
			},
			Action: gitai.FeedbackCmd,
		},
	)
}
