package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/booscaaa/observability-go-example/internals/domain/detection"
)

func main() {
	inputFile := flag.String("input", "data/large_file_transfer_logs.json", "JSON file containing transfer events")
	inputDir := flag.String("input-dir", "", "directory containing behavioral JSON logs")
	format := flag.String("format", "json", "output format: json or prometheus")
	listen := flag.String("listen", "", "serve /metrics, /report and /healthz on this address")
	flag.Parse()
	if *inputDir != "" {
		report, err := detection.AnalyzeDirectory(*inputDir)
		if err != nil {
			fail("analyze directory", err)
		}
		if *listen != "" {
			serve(report, *listen)
			return
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

	if *listen != "" {
		serve(report, *listen)
		return
	}
	writeReport(report, *format)
}

func serve(report detection.Report, address string) {
	log.Printf("insider threat metrics listening on %s", address)
	if err := http.ListenAndServe(address, detection.Handler(report)); err != nil {
		fail("serve metrics", err)
	}
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
