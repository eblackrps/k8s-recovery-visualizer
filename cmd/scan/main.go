package main

import (
	"log"
	"os"

	"k8s-recovery-visualizer/internal/scanapp"
)

func main() {
	log.SetFlags(0)
	os.Exit(scanapp.Main(os.Args[1:]))
}
