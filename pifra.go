package pifra

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
)

// InteractiveMode allows the user to inspect interactively the next transition(s)
// after providing a pi-calculus syntax input.
func InteractiveMode() {
	for {
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		proc, err := InitProgram([]byte(text))
		if err != nil {
			fmt.Println("error:", err)
		} else {
			fmt.Println(PrettyPrintAst(proc))
			state := newTransitionStateRoot(proc)
			confs := trans(state.Configuration)
			for _, conf := range confs {
				fmt.Println(PrettyPrintConfiguration(conf))
			}
		}
	}
}

// OutputMode generates an LTS from the pi-calculus program file and either writes
// the output to a file, or prints the output if an output file is not specified.
func OutputMode(maxStates int, inputFile string, outputFile string) error {
	maxStatesExplored = maxStates

	input, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return err
	}
	output, err := generateLts(input)
	if err != nil {
		return err
	}

	if outputFile == "" {
		fmt.Println(string(output))
	} else {
		return writeFile(output, outputFile)
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

func generateLts(input []byte) ([]byte, error) {
	proc, err := InitProgram(input)
	if err != nil {
		return nil, err
	}
	root := newTransitionStateRoot(proc)
	vertices, edges := exploreTransitions(root)
	output := generateGraphVizFile(vertices, edges)
	return output, nil
}
