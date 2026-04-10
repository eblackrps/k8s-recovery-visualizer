package main

import (
	"log"
	"os"

	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/scanapp"
)

func main() {
	log.SetFlags(0)
	os.Exit(scanapp.Main(os.Args[1:]))
}

// write is kept as a small compatibility shim for artifact tests.
func write(bundle *model.Bundle, outDir string, quiet bool, minScore int, csvExport, summaryOut, redactOut, runbookOut bool) (string, int) {
	trendLabel, trendDelta, err := scanapp.WriteOutputs(bundle, outDir, quiet, minScore, csvExport, summaryOut, redactOut, runbookOut)
	if err != nil {
		log.Fatalf("write outputs failed: %v", err)
	}
	return trendLabel, trendDelta
}
