package processors

import (
	"fmt"

	"github.com/lycerius/epilguard/tools"
)

//FlashingProcessor Processes a video stream and detects flashing photosensitive content
type FlashingProcessor struct {
	decoder       *tools.FFMPEGDecoder //Decoder to fetch frames from
	JobID         string               //The job assosiated with this request
	HazardReport  HazardReport         //Generated hazard report
	AreaThreshold float32
}

type brightnessFrame struct {
	Index         uint
	data          []int
	Height, Width int
}

type frameDifference struct {
	Index         uint
	Height, Width int
	negatives     map[int]int
	positives     map[int]int
}

//NewFlashingProcessor creates a flashing processor
func NewFlashingProcessor(f *tools.FFMPEGDecoder, jobID string) FlashingProcessor {
	processor := FlashingProcessor{}

	processor.decoder = f
	processor.JobID = jobID

	return processor
}

//Process begins scanning the video for flashing photosensitive content
func (f *FlashingProcessor) Process() error {
	//frameDifferences := list.New()
	frame, err := f.decoder.Next()

	if err != nil {
		return err
	}

	lastFrame := rGBFrameToLuminance(&frame)

	//pixelCountThreshold := int(float32(lastFrame.Height) * float32(lastFrame.Width) * f.AreaThreshold)
	for {
		//Step 1: Compute frame difference
		frame, err = f.decoder.Next()

		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}

		nextFrame := rGBFrameToLuminance(&frame)

		//Step 2: Generate hK+ and hK- of positive and negative differences
		/*difference :=*/
		calculateFrameDifference(lastFrame, nextFrame)
		fmt.Println(nextFrame.Index)
		lastFrame = nextFrame
	}

	return nil
}

func rGBFrameToLuminance(frame *tools.Frame) brightnessFrame {

	var lframe brightnessFrame
	lframe.Height = frame.Height
	lframe.Width = frame.Width
	size := frame.Height * frame.Width
	data := make([]int, size, size)

	for y := 0; y < frame.Height; y++ {
		for x := 0; x < frame.Width; x++ {
			pixel := frame.GetRGB(x, y)
			lum := tools.RGBtoBrightness(float32(pixel.Red), float32(pixel.Green), float32(pixel.Blue))
			data[y*frame.Width+x] = lum
		}
	}
	lframe.Index = frame.Index
	lframe.data = data
	return lframe
}

func calculateFrameDifference(f1, f2 brightnessFrame) frameDifference {
	var frameDifference frameDifference
	positives := make(map[int]int)
	negatives := make(map[int]int)

	for i := 0; i < f1.Height*f1.Width; i++ {
		difference := f2.data[i] - f1.data[i]
		if difference > 0 {
			positives[difference]++
		} else if difference < 0 {
			negatives[(-difference)]++
		}
	}
	frameDifference.Height = f1.Height
	frameDifference.Width = f1.Width
	frameDifference.positives = positives
	frameDifference.negatives = negatives
	frameDifference.Index = f2.Index
	return frameDifference
}
