package pifra

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// Flags are the user-specified flags for the command line.
type Flags struct {
	InteractiveMode bool

	RegisterSize int
	MaxStates    int
	DisableGC    bool

	InputFile  string
	OutputFile string

	GVLayout       string
	GVOutputStates bool

	Pretty     bool
	Statistics bool

	Quiet bool
}

func initFlags(flags Flags) {
	maxStatesExplored = flags.MaxStates
	registerSize = flags.RegisterSize
	disableGarbageCollection = flags.DisableGC
}

// InteractiveMode allows the user to inspect interactively the next transition(s)
// after providing a pi-calculus syntax input.
func InteractiveMode(flags Flags) {
	initFlags(flags)
	for {
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		proc, err := InitProgram([]byte(text))
		if err != nil {
			fmt.Println("error:", err)
		} else {
			fmt.Println(PrettyPrintAst(proc))
			root := newRootConf(proc)
			confs := trans(root)
			for _, conf := range confs {
				fmt.Println(PrettyPrintConfiguration(conf))
			}
		}
	}
}

// OutputMode generates an LTS from the pi-calculus program file and either writes
// the output to a file, or prints the output if an output file is not specified.
func OutputMode(flags Flags) error {
	initFlags(flags)
	gvLayout = flags.GVLayout

	start := time.Now()

	input, err := ioutil.ReadFile(flags.InputFile)
	if err != nil {
		return err
	}
	lts, err := generateLts(input)
	if err != nil {
		return err
	}

	if !flags.Quiet {
		if flags.OutputFile == "" {
			// No output file specified. Print LTS.
			output := generatePrettyLts(lts)
			fmt.Println(string(output))
		} else {
			// Output file specified. Write to file.
			var output []byte
			if flags.Pretty {
				output = generatePrettyLts(lts)
			} else {
				output = generateGraphVizFile(lts, flags.GVOutputStates)
			}
			return writeFile(output, flags.OutputFile)
		}
	}

	elapsed := time.Since(start)
	if flags.Statistics {
		if !flags.Quiet && flags.OutputFile == "" {
			// Print new line if LTS is printed to standard output.
			fmt.Println()
		}
		fmt.Printf("states explored   %d\n", lts.StatesExplored)
		fmt.Printf("states generated  %d\n", lts.StatesGenerated)
		fmt.Printf("states unique     %d\n", len(lts.States))
		fmt.Printf("transitions       %d\n", len(lts.Transitions))
		fmt.Printf("time              %s\n", elapsed)
	}

	return nil
}

func writeFile(output []byte, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	_, err = file.Write(output)
	if err != nil {
		file.Close()
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

func generateLts(input []byte) (Lts, error) {
	proc, err := InitProgram(input)
	if err != nil {
		return Lts{}, err
	}
	root := newRootConf(proc)
	lts := explore(root)
	return lts, nil
}
