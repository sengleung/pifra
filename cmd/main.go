package main

import (
	"fmt"
	"os"

	"github.com/sengleung/pifra"
	"github.com/spf13/cobra"
)

var flags pifra.Flags

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
		if flags.RegisterSize < 1 {
			fmt.Println("error: register size must be greater or equal to 1")
			os.Exit(1)
		}
		if flags.MaxStates < 0 {
			fmt.Println("error: maximum states explored must be positive")
			os.Exit(1)
		}
		if flags.InteractiveMode {
			pifra.InteractiveMode(flags)
		} else {
			if len(args) < 1 {
				fmt.Println("error: input file required for LTS generation")
				fmt.Println(cmd.UsageString())
				os.Exit(1)
			}
			if len(args) > 1 {
				fmt.Println("error: more than one argument encountered")
				fmt.Println(cmd.UsageString())
				os.Exit(1)
			}
			flags.InputFile = args[0]
			if err := pifra.OutputMode(flags); err != nil {
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

	rootCmd.PersistentFlags().BoolVarP(&flags.InteractiveMode, "interactive", "i", false, "inspect interactively the next transitions after providing input")
	rootCmd.PersistentFlags().StringVarP(&flags.OutputFile, "output", "o", "", "output the LTS to a file in a Graphviz DOT format")
	rootCmd.PersistentFlags().IntVarP(&flags.MaxStates, "max-states", "n", 20, "maximum number of transition states explored")
	rootCmd.PersistentFlags().IntVarP(&flags.RegisterSize, "register-size", "r", 1073741824, "register size")
	rootCmd.PersistentFlags().BoolVarP(&flags.GVOutputStates, "output-states", "s", false, "output state numbers instead of configurations for the Graphviz DOT file")

	rootCmd.PersistentFlags().BoolP("help", "h", false, "show this help message and exit")
}

func main() {
	execute()
}
