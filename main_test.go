package main

import (
	"bytes"
	"encoding/json"
	"maps"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestTestGeneration(t *testing.T) {
	pkgName := "test"
	clear := func() {
		entries, _ := os.ReadDir(".")
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), "_test.go") {
				os.Remove(e.Name())
			}
		}
	}
	cd := func(path string) error {
		return os.Chdir(path)
	}

	generate := func() {
		main()
	}

	runTest := func() ([]byte, error) {
		cmd := exec.Command("go", "test", ".", "-json")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return out, err
		}
		return out, nil
	}

	cd(pkgName)
	clear()
	generate()
	testOutput, err := runTest()
	if err == nil {
		t.Fatalf("expected generated tests to fail")
	}

	type testAction struct {
		Action string
		Test   string
	}

	var testActions []testAction

	decoder := json.NewDecoder(bytes.NewReader(testOutput))
	for decoder.More() {
		var testAction testAction
		if err = decoder.Decode(&testAction); err != nil {
			t.Fatalf("failed to unmarshal test output: %v\n", err)
		}
		testActions = append(testActions, testAction)
	}

	runTests := make(map[string]string)
	nRan := 0
	for _, ta := range testActions {
		if ta.Action == "output" {
			continue
		}
		if ta.Test == "" {
			// package fail message
			continue
		}
		if ta.Action == "run" {
			nRan++
		}
		runTests[ta.Test] = ta.Action
	}

	if len(runTests) != nRan {
		t.Logf("recoded action: %+v", testActions)
		t.Fatalf("expected output for %d tests, got ouput for %d\n", nRan, len(runTests))
	}

	for action := range maps.Values(runTests) {
		if action != "fail" {
			t.Fatalf("expected test fail, got=%s\n", action)
		}
	}
	// 	{"Time":"2025-07-13T21:34:58.164318375+02:00","Action":"fail","Package":"github.com/moritz-tiesler/tscaff/test","Test":"TestADD","Elapsed":0}
	// {"Time":"2025-07-13T21:34:58.164321512+02:00","Action":"run","Package":"github.com/moritz-tiesler/tscaff/test","Test":"TestDO"}
	// {"Time":"2025-07-13T21:34:58.164323228+02:00","Action":"output","Package":"github.com/moritz-tiesler/tscaff/test","Test":"TestDO","Output":"=== RUN   TestDO\n"}

}
