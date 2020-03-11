package pifra

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLtsGeneration(t *testing.T) {
	t.Run("lts_generation", func(t *testing.T) {
		// Current working directory.
		pwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		// Test directory.
		testFolder := path.Join(pwd, "test")
		// Output directory.
		outFolder := path.Join(pwd, "test", "out")

		// Get test files, trimming extension ".pi".
		var testFiles []string
		filepath.Walk(testFolder, func(path string, f os.FileInfo, _ error) error {
			if !f.IsDir() {
				if filepath.Ext(path) == ".pi" {
					testFiles = append(testFiles, strings.TrimSuffix(f.Name(), filepath.Ext(f.Name())))
				}
			}
			return nil
		})

		// Remove output directory when finished.
		defer os.RemoveAll(outFolder)

		// Test LTS pretty print output.
		flags := Flags{
			MaxStates:    10,
			RegisterSize: 1073741824,
			Pretty:       true,
		}
		compareLts(t, flags, testFiles, testFolder, outFolder, ".txt")

		// Test LTS GraphViz DOT output.
		flags = Flags{
			MaxStates:    10,
			RegisterSize: 1073741824,
		}
		compareLts(t, flags, testFiles, testFolder, outFolder, ".dot")
	})
}

func compareLts(t *testing.T, flags Flags, testFiles []string, testFolder string, outFolder string, ext string) {
	// Write LTS's to output directory.
	for _, testFile := range testFiles {
		outputPath := path.Join(outFolder, testFile+ext)
		testPath := path.Join(testFolder, testFile+".pi")

		flags.OutputFile = outputPath
		flags.InputFile = testPath
		if err := OutputMode(flags); err != nil {
			t.Error(err)
		}

		flags.OutputFile = outputPath
		flags.InputFile = testPath
		if err := OutputMode(flags); err != nil {
			t.Error(err)
		}
	}

	// Read LTS's from output directory, read LTS's from test directory and compare.
	for _, testFile := range testFiles {
		outputPath := path.Join(outFolder, testFile+ext)
		testOutputPath := path.Join(testFolder, testFile+ext)

		outputFile, err := ioutil.ReadFile(outputPath)
		if err != nil {
			t.Error(err)
		}
		testOutputFile, err := ioutil.ReadFile(testOutputPath)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(outputFile, testOutputFile) {
			t.Errorf("not equal: %s, generated:\n%s", testFile+ext, outputFile)
		}
	}
}
