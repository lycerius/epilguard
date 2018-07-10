package main

import (
	"log"
	"os"

	"github.com/lycerius/epilguard/hazards"

	"github.com/lycerius/epilguard/decoder"
	"github.com/lycerius/epilguard/processors"
)

func main() {
	if len(os.Args) != 3 {
		if len(os.Args) == 1 {
			printHelp()
			os.Exit(1)
		} else {
			log.Fatal("Expected 2 arg, got ", len(os.Args)-1)
		}
	}

	path := os.Args[1]
	jobID := os.Args[2]

	if _, err := os.Stat(path); err != nil {
		log.Fatal("Could not open '", path, "', ", err)
	}

	decoder := decoder.NewDecoder(path)
	decoder.FrameBufferCacheSize = 15
	decoder.Start()
	processor := processors.NewFlashingProcessor(&decoder, jobID)
	err := processor.Process()

	if err != nil {
		log.Fatal(err)
	}

	report := processor.HazardReport

	decoder.Close()
	hazards.UploadHazardReport(report)
	os.Exit(0)
}

func printHelp() {
	println("USAGE: epilguard [input-file] [job-id]")
}
