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
	flag.Parse()

	file, err := os.Open(*inputFile)
	if err != nil {
		fail("open input", err)
	}
	defer file.Close()

	report, err := detection.Analyze(file, *inputFile)
	if err != nil {
		fail("analyze input", err)
	}

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