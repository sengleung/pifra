package main

import (
	"fmt"
	"os"

	"github.com/sengleung/pifra"
	"github.com/spf13/cobra"
)

var maxStates int
var registerSize int
var maxProcessDepth int
var interactiveMode bool
var outputFile string

var usageTemplate = []byte(`Usage:{{if .Runnable}}
{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
{{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
{{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Options:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
{{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)

var rootCmd = &cobra.Command{
	Use:   "pifra [OPTION...] [FILE]",
	Short: "LTS generator for the pi-calculus represented by FRA.",
	Long: `Labelled transition system (LTS) generation for the
pi-calculus represented by fresh-register automata.`,
	Run: func(cmd *cobra.Command, args []string) {
		if registerSize < 1 {
			fmt.Println("error: register size must be greater or equal to 1")
			os.Exit(1)
		}
		if maxProcessDepth < 0 {
			fmt.Println("error: maximum process depth must be positive")
			os.Exit(1)
		}
		if maxStates < 0 {
			fmt.Println("error: error: maximum states explored must be positive")
			os.Exit(1)
		}
		if interactiveMode {
			if err := pifra.InteractiveMode(registerSize); err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}
		} else {
			if len(args) < 1 {
				fmt.Println("error: input file required for LTS generation")
				os.Exit(1)
			}
			if len(args) > 1 {
				fmt.Println("error: more than one argument encountered")
				os.Exit(1)
			}
			inputFile := args[0]
			if err := pifra.OutputMode(maxProcessDepth, registerSize, maxStates, inputFile, outputFile); err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}
		}
	},
}

func execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.DisableFlagsInUseLine = true
	rootCmd.SetUsageTemplate(string(usageTemplate))

	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false

	rootCmd.PersistentFlags().BoolVarP(&interactiveMode, "interactive", "i", false, "inspect interactively the next transition(s) after providing input")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "output the LTS to a Graphviz DOT file")
	rootCmd.PersistentFlags().IntVarP(&maxProcessDepth, "depth", "d", 50, "maximum process depth during parsing\nexample: P = a(b).P will only resolve to the maximum depth and\nthen assign the nil process 0, i.e. a(b)...a(b).0")
	rootCmd.PersistentFlags().IntVarP(&maxStates, "max-states", "s", 50, "maximum number of transition states explored")
	rootCmd.PersistentFlags().IntVarP(&registerSize, "register-size", "r", 10000, "register size")

	rootCmd.PersistentFlags().BoolP("help", "h", false, "show this help message and exit")
}

func main() {
	execute()
}
