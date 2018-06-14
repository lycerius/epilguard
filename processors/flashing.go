package processors

import (
	"container/list"

	"github.com/epilguard/tools"
)

//FlashingProcessor Processes a video stream and detects flashing photosensitive content
type FlashingProcessor struct {
	decoder       *tools.FFMPEGDecoder //Decoder to fetch frames from
	JobID         string               //The job assosiated with this request
	HazardReport  HazardReport         //Generated hazard report
	AreaThreshold float32
}

type LuminanceFrame struct {
	Index         uint
	data          []int
	Height, Width int
}

type TransitionFrame struct {
	Index         uint
	Height, Width int
	negatives     list.List
	positives     list.List
}

func NewFlashingProcessor(f *tools.FFMPEGDecoder, jobId string) FlashingProcessor {
	processor := FlashingProcessor{}

	processor.decoder = f
	processor.JobID = jobId

	return processor
}

//Process begins scanning the video for flashing photosensitive content
func (f *FlashingProcessor) Process() error {
	//frameDifferences := list.New()
	lastFrame := RGBFrameToLuminance(f.decoder.Next())
	pixelCountThreshold := int(float32(lastFrame.Height) * float32(lastFrame.Width) * f.AreaThreshold)
	for {
		//Get frame difference
		nextFrame := RGBFrameToLuminance(f.decoder.Next())
		difference := calculateFrameDifference(lastFrame, nextFrame)

		//Generate positive and negative histos
		var histoPos, histoNeg [255]int

		//frameDifferences.PushBack(difference)
		lastFrame = nextFrame

		/*for k, v := range difference.negatives {
			fmt.Println(k, ": ", v)
		}
		for k, v := range difference.positives {
			fmt.Println(k, ": ", v)
		}*/

	}
	//Step 1: Calculate Frame differences

	//Step 2: Generate positive and negative histograms

	//Step 3: Scan both histograms from right to left until you have percentage of pixels required for hazard id

	//Step 4: Compute positive and negative averages
	return nil
}

func RGBFrameToLuminance(frame *tools.Frame) LuminanceFrame {

	var lframe LuminanceFrame
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

func calculateFrameDifference(f1, f2 LuminanceFrame) TransitionFrame {
	var frameDifference TransitionFrame
	positives := list.New()
	negatives := list.New()
	for i := 0; i < f1.Height*f1.Width; i++ {
		difference := f2.data[i] - f1.data[i]
		if difference > 0 {
			positives.PushBack(difference)
		} else if difference < 0 {
			negatives.PushBack(difference)
		}
	}
	frameDifference.Height = f1.Height
	frameDifference.Width = f1.Width
	frameDifference.positives = positives
	frameDifference.negatives = negatives
	frameDifference.Index = f2.Index
	return frameDifference
}
