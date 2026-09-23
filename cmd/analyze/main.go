package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/booscaaa/observability-go-example/internals/domain/detection"
)

func main() {
	inputFile := flag.String("input", "data/large_file_transfer_logs.json", "JSON file containing transfer events")
	inputDir := flag.String("input-dir", "", "directory containing behavioral JSON logs")
	flag.Parse()
	if *inputDir != "" {
		report, err := detection.AnalyzeDirectory(*inputDir)
		if err != nil {
			fail("analyze directory", err)
		}
		writeReport(report)
		return
	}

	file, err := os.Open(*inputFile)
	if err != nil {
		fail("open input", err)
	}
	defer file.Close()

	report, err := detection.Analyze(file, *inputFile)
	if err != nil {
		fail("analyze input", err)
	}

	writeReport(report)
}

func writeReport(report detection.Report) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fail("encode report", err)
	}
}

func fail(operation string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", operation, err)
	os.Exit(1)
}
