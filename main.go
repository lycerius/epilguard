package main

import (
	"log"
	"os"

	"github.com/epilguard/processors"
	"github.com/epilguard/tools"
)

func main() {
	if len(os.Args) != 2 {
		if len(os.Args) == 1 {
			printHelp()
			os.Exit(1)
		} else {
			log.Fatal("Expected 1 arg, got ", len(os.Args)-1)
		}
	}

	path := os.Args[1]

	if _, err := os.Stat(path); err != nil {
		log.Fatal("Could not open '", path, "', ", err)
	}

	decoder := tools.NewDecoder(path)
	decoder.FrameBufferSize = 2
	decoder.Start()
	processor := processors.NewFlashingProcessor(&decoder, "helloworld")
	processor.Process()
}

func printHelp() {
	println("USAGE: epilguard [input-file]")
}
