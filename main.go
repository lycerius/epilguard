package main

import (
	"log"

	"github.com/epilguard/processors"
	"github.com/epilguard/tools"
)

func main() {
	decoder := tools.NewDecoder("C:\\Users\\Nathan C. Purpura\\Downloads\\bbb_sunflower_1080p_60fps_stereo_abl.mp4")
	decoder.FrameBufferSize = 2
	decoder.Start()

	processor := processors.NewFlashingProcessor(&decoder, "helloworld")

	log.Fatal(processor.Process())

}
