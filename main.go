package main

import (
	"log"
	"os"

	"github.com/lycerius/epilguard/decoder"
	"github.com/lycerius/epilguard/processors"
)

//main Main entry point
func main() {

	checkArgumentsLength()
	path := os.Args[1]
	reportDirectory := os.Args[2]

	//blank path? use cwd
	if path == "" {
		var err error
		reportDirectory, err = os.Getwd()
		if err != nil {
			log.Fatal("Could not use current working directory as csv directory '", err, "'")
		}
	}

	//Video must exist at path
	if _, err := os.Stat(path); err != nil {
		log.Fatal("Could not open '", path, "', ", err)
	}

	//Create decoder
	decoder := decoder.NewDecoder(path)
	decoder.FrameBufferCacheSize = 15
	decoder.Start()

	//Attatch to new processor
	processor := processors.NewFlashingProcessor(&decoder, reportDirectory)

	//Look for hazards
	err := processor.Process()

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}

func checkArgumentsLength() {
	if len(os.Args) != 3 {
		if len(os.Args) == 1 {
			printHelp()
			os.Exit(1)
		} else {
			log.Fatal("Expected 2 arg, got ", len(os.Args)-1)
		}
	}
}

func printHelp() {
	println("USAGE: epilguard [input-file] [csv-export-directory]")
}
