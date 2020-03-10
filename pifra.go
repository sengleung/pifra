package pifra

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
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

	inputTimeStart := time.Now()
	input, err := ioutil.ReadFile(flags.InputFile)
	if err != nil {
		return err
	}
	inputTime := time.Since(inputTimeStart)

	programTimeStart := time.Now()
	lts, err := generateLts(input)
	if err != nil {
		return err
	}
	programElapsed := time.Since(programTimeStart)

	var outputTime time.Duration

	if !flags.Quiet {
		if flags.OutputFile == "" {
			// No output file specified. Print LTS.
			output := generatePrettyLts(lts)

			outputTimeStart := time.Now()
			fmt.Println(string(output))
			outputTime = time.Since(outputTimeStart)
		} else {
			// Output file specified. Write to file.
			var output []byte
			if flags.Pretty {
				output = generatePrettyLts(lts)
			} else {
				output = generateGraphVizFile(lts, flags.GVOutputStates)
			}
			outputTimeStart := time.Now()
			if err := writeFile(output, flags.OutputFile); err != nil {
				return err
			}
			outputTime = time.Since(outputTimeStart)
		}
	}

	if flags.Statistics {
		if !flags.Quiet && flags.OutputFile == "" {
			// Print new line if LTS is printed to standard output.
			fmt.Println()
		}
		ioElapsed := inputTime + outputTime
		fmt.Printf("states explored      %d\n", lts.StatesExplored)
		fmt.Printf("states generated     %d\n", lts.StatesGenerated)
		fmt.Printf("states unique        %d\n", len(lts.States))
		fmt.Printf("transitions          %d\n", len(lts.Transitions))
		fmt.Printf("time I/O             %s\n", ioElapsed)
		fmt.Printf("time LTS generation  %s\n", programElapsed)
	}

	return nil
}

func writeFile(output []byte, outputFile string) error {
	dir := path.Dir(outputFile)
	os.MkdirAll(dir, os.ModePerm)
	return ioutil.WriteFile(outputFile, output, 0644)
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
