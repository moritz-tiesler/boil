package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		run(false)
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
	fmt.Println(string(testOutput))

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
		if ta.Action == "build-fail" {
			t.Fatal("tests did not compile")
		}
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

	t.Logf("%d tests ran\n", nRan)
	if nRan == 0 {
		t.Fatal("no generated test were ran")
	}

	if len(runTests) != nRan {
		t.Logf("recoded action: %+v", testActions)
		t.Fatalf("expected output for %d tests, got ouput for %d\n", nRan, len(runTests))
	}

	nFail := 0
	for _, action := range runTests {
		if action != "fail" {
			t.Fatalf("expected test fail, got=%s\n", action)
		}
		nFail++
	}
	clear()
	t.Logf("%d tests failed\n", nRan)

}
func TestTableTestGeneration(t *testing.T) {
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
		run(true)
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
	fmt.Println(string(testOutput))

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
		if ta.Action == "build-fail" {
			t.Fatal("tests did not compile")
		}
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

	t.Logf("%d tests ran\n", nRan)
	if nRan == 0 {
		t.Fatal("no generated test were ran")
	}

	if len(runTests) != nRan {
		t.Logf("recoded action: %+v", testActions)
		t.Fatalf("expected output for %d tests, got ouput for %d\n", nRan, len(runTests))
	}

	nFail := 0
	for _, action := range runTests {
		if action != "fail" {
			t.Fatalf("expected test fail, got=%s\n", action)
		}
		nFail++
	}
	clear()
	t.Logf("%d tests failed\n", nRan)

}
