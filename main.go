package main

import (
	"log"
	"time"

	"github.com/epilguard/tools"
)

func main() {
	decoder := tools.NewDecoder("/Users/longdog/Downloads/bbb_sunflower_1080p_60fps_stereo_abl.mp4")
	err := decoder.Start()
	if err != nil {
		log.Fatal(err)
		return
	}
	for {
		<-decoder.FrameBuffer
		time.Sleep(1 * time.Second)
	}
}
