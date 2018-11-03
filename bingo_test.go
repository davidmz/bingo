package main_test

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func TestMain(m *testing.M) {
	exeName := "bingo"
	if runtime.GOOS == "windows" {
		exeName = exeName + ".exe"
	}
	build := exec.Command("go", "build", "-o", exeName)

	if err := build.Run(); err != nil {
		fmt.Println("could not make executable file:", err)
		os.Exit(1)
	}

	exitCode := m.Run()

	os.Remove(exeName)
	os.Exit(exitCode)
}

func TestBingo(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		noStdin bool
	}{
		{"no-args", []string{}, false},
		{"pkg-var", []string{"-var", "Bar", "-pkg", "pack"}, false},
		{"from-file", []string{"-in", "_fixtures/lorem-ipsum.txt"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenFile := "_fixtures/" + tt.name + ".golden"
			data, err := ioutil.ReadFile("_fixtures/lorem-ipsum.txt")
			if err != nil {
				t.Fatal("Cannot read data file:", err)
			}

			cmd := exec.Command("./bingo", tt.args...)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal("Cannot open stdin:", err)
			}
			go func() {
				defer stdin.Close()
				stdin.Write(data)
			}()

			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatal("Cannot run bingo:", err)
			}

			if *update {
				err = ioutil.WriteFile(goldenFile, out, 0644)
				if err != nil {
					t.Fatal("Cannot write golden file:", err)
				}
			}

			expected, err := ioutil.ReadFile(goldenFile)
			if err != nil {
				t.Fatal("Cannot read golden file:", err)
			}

			if string(out) != string(expected) {
				t.Errorf("Expected: %q, got %q", string(expected), string(out))
			}
		})
	}
}
