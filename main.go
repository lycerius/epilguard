package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/lycerius/epilguard/decoder"
	"github.com/lycerius/epilguard/processors"
)

var reportDirectory string
var videoFile string
var frameBufferLength uint

//main Main entry point
func main() {
	processArguments()

	//Video must exist at path
	if _, err := os.Stat(videoFile); err != nil {
		log.Fatal("Could not open '", videoFile, "', ", err)
	}

	//Create decoder
	decoder := decoder.NewDecoder(videoFile)
	decoder.FrameBufferCacheSize = int(frameBufferLength)
	decoder.Start()

	//Attatch to new processor
	processor := processors.NewFlashingProcessor(&decoder, reportDirectory)

	//Look for hazards
	err := processor.Process()

	if err != nil {
		log.Fatal(err)
	}
}

func processArguments() {
	flag.StringVar(&reportDirectory, "report-dir", "", "directory to write report files to (default $cwd)")
	flag.UintVar(&frameBufferLength, "buffer-size", 30, "Sets the size of the lookahead framebuffer, must be > 0")

	flag.Usage = func() {
		fmt.Println("epilguard [options] video")
		fmt.Println()
		flag.PrintDefaults()
	}

	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if reportDirectory == "" {
		rDir, err := os.Getwd()

		if err != nil {
			log.Fatal(err)
		}

		reportDirectory = rDir
	}

	if frameBufferLength <= 0 {
		flag.Usage()
		os.Exit(1)
	}

	videoFile = flag.Arg(0)

}
