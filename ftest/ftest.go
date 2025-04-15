package ftest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type TotalStats struct {
	countOk      int
	countFail    int
	failMessages []string
}

type Decision interface {
	Process(stats *TotalStats, tc *TestCase)
}

type OkDecision struct {
}

func (d *OkDecision) Process(stats *TotalStats, tc *TestCase) {
	stats.countOk++
}

type FailDecision struct {
}

func (d *FailDecision) Process(stats *TotalStats, tc *TestCase) {
	stats.countFail++
	stats.failMessages = append(stats.failMessages,
		strings.Repeat("-", 10)+fmt.Sprintf("[FAIL] test №%d", tc.idx+1)+strings.Repeat("-", 10)+"\n"+
			fmt.Sprintf("Input:\n%s\nExpected:\n%s\nActual:\n%s\n",
				strings.TrimSpace(tc.input),
				strings.TrimSpace(tc.expectedOut),
				strings.TrimSpace(tc.actualOut),
			))
}

const (
	testsFileSeparator = "@TEST SEPARATOR@"
)

type TestCase struct {
	idx         int
	input       string
	expectedOut string
	actualOut   string
	decision    Decision
}

func extractTestCasesFromFile(file string) []string {
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("[WARN] - Error with file reading %s: %v\n", file, err)
		return nil
	}

	return strings.Split(string(data), testsFileSeparator)
}

func extractTestCases(folderPath string, filesPattern string) []TestCase {
	var testCases []TestCase

	inputFiles, err := filepath.Glob(filepath.Join(folderPath, filesPattern))
	if err != nil {
		fmt.Println("[WARN] - Finding files error:", err)
		return testCases
	}

	if len(inputFiles) == 0 {
		return testCases
	}

	for _, inputFile := range inputFiles {
		base := filepath.Base(inputFile)
		outputFile := filepath.Join(folderPath, strings.Replace(base, "input", "output", 1))

		testCasesIn := extractTestCasesFromFile(inputFile)
		testCasesOut := extractTestCasesFromFile(outputFile)
		if len(testCasesIn) != len(testCasesOut) {
			fmt.Printf("[WARN] - Count of tests in the files does not match: %s - %d, %s - %d\n", inputFile, len(testCasesIn), outputFile, len(testCasesOut))
			continue
		}

		for idx, testIn := range testCasesIn {
			testCases = append(testCases, TestCase{
				idx,
				testIn,
				testCasesOut[idx],
				"",
				&OkDecision{},
			})
		}
	}
	return testCases
}

func runTest(testCase *TestCase, process func(io.Reader) string, chTestIdxes chan<- int) bool {
	readerIn := strings.NewReader(testCase.input)
	testCase.actualOut = process(readerIn)

	actualClean := strings.TrimSpace(testCase.actualOut)
	expectedClean := strings.TrimSpace(testCase.expectedOut)

	if actualClean != expectedClean {
		testCase.decision = &FailDecision{}
	}

	chTestIdxes <- testCase.idx

	return actualClean != expectedClean
}

func runWithGorutines(testCases []TestCase, process func(io.Reader) string, chTestIdxes chan<- int) {
	var wg sync.WaitGroup
	wg.Add(len(testCases))
	for idx := range testCases {
		go func(tc *TestCase) {
			defer wg.Done()

			runTest(tc, process, chTestIdxes)
		}(&testCases[idx])
	}
	wg.Wait()
}

func RunTests(folderPath string, filesPattern string, process func(io.Reader) string, isParallel bool, stopOnFirstFail bool) {
	start := time.Now()

	testCases := extractTestCases(folderPath, filesPattern)

	if len(testCases) == 0 {
		fmt.Println("[WARN] - Tests aren't found")
		return
	}

	chTestIdxes := make(chan int, len(testCases))

	go func() {
		if isParallel {
			runWithGorutines(testCases, process, chTestIdxes)
		} else {
			for idx := range testCases {
				if stopOnFirstFail && runTest(&testCases[idx], process, chTestIdxes) {
					break
				}
			}
		}
		close(chTestIdxes)
	}()

	stats := &TotalStats{}
	for idx := range chTestIdxes {
		testCase := &testCases[idx]
		testCase.decision.Process(stats, testCase)

		if stats.countFail > 0 {
			fmt.Print(stats.failMessages[stats.countFail-1])

			if stopOnFirstFail {
				break
			}
		}
	}

	duration := time.Since(start)

	if stats.countFail == 0 {
		fmt.Printf("[OK] - All tests (%d) successfully passed!\nTotal duration: %v\n", stats.countOk, duration)
	}
}
