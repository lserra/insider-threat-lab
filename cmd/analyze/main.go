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
	format := flag.String("format", "json", "output format: json or prometheus")
	flag.Parse()
	if *inputDir != "" {
		report, err := detection.AnalyzeDirectory(*inputDir)
		if err != nil {
			fail("analyze directory", err)
		}
		writeReport(report, *format)
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

	writeReport(report, *format)
}

func writeReport(report detection.Report, format string) {
	if format == "prometheus" {
		fmt.Print(detection.Prometheus(report))
		return
	}
	if format != "json" {
		fail("format", fmt.Errorf("unsupported format %q", format))
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
