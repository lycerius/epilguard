package processors

import (
	"container/list"
	"fmt"

	"github.com/lycerius/epilguard/decoder"
	"github.com/lycerius/epilguard/equations"
)

//FlashingProcessor Processes a video stream and detects flashing photosensitive content
type FlashingProcessor struct {
	decoder       *decoder.Decoder //Decoder to fetch frames from
	JobID         string           //The job assosiated with this request
	HazardReport  HazardReport     //Generated hazard report
	AreaThreshold float32
}

type brightnessFrame struct {
	Index         uint
	Data          []int
	Height, Width int
}

type frameDifference struct {
	Index          uint
	Height, Width  int
	Negatives      map[int]int
	Positives      map[int]int
	MaxPos, MaxNeg int
}

//NewFlashingProcessor creates a flashing processor
func NewFlashingProcessor(f *decoder.Decoder, jobID string) FlashingProcessor {
	processor := FlashingProcessor{}

	processor.decoder = f
	processor.JobID = jobID

	return processor
}

//Process begins scanning the video for flashing photosensitive content
func (proc *FlashingProcessor) Process() error {

	frame, err := proc.decoder.NextFrame()

	if err != nil {
		return err
	}

	frameDifferences := list.New()
	accFrameDifferences := list.New()
	accLuminance := 0
	lFrame := rGBFrameToLuminance(&frame)
	lastFrame := &lFrame

	//pixelCountThreshold := int(float32(lastFrame.Height) * float32(lastFrame.Width) * equations.PercentageFlashArea)
	for {
		//Step 1: Compute frame difference
		frame, err = proc.decoder.NextFrame()

		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}

		nextFrame := rGBFrameToLuminance(&frame)

		//Step 2: Generate hK+ and hK- of positive and negative differences
		difference := calculateFrameDifference(*lastFrame, nextFrame)

		//Step 3: Scan from right to left, filling bins until the number of elements
		//Equals tha area neeeded for flashing content (pixelCountThreshold)

		luminanceDifference := findAverageLuminance(difference)

		//Signs are the same? accumulate
		if (accLuminance < 0) == (luminanceDifference < 0) {
			accLuminance += luminanceDifference
		} else { //Reset
			accLuminance = luminanceDifference
		}

		frameDifferences.PushBack(luminanceDifference)
		accFrameDifferences.PushBack(accLuminance)
		//fmt.Println(nextFrame.Index)
		lastFrame = &nextFrame
	}

	index := 1
	fmt.Println("Index\tLuminance\taccLuminance")
	acc := accFrameDifferences.Front()
	for ele := frameDifferences.Front(); ele != nil; ele = ele.Next() {
		val := ele.Value.(int)
		fmt.Println(index, "\t", val, "\t\t", acc.Value.(int))
		index++
		acc = acc.Next()
	}
	return nil
}

//In the future, we may want to parallize this
func rGBFrameToLuminance(frame *decoder.Frame) brightnessFrame {

	var lframe brightnessFrame
	lframe.Height = frame.Height
	lframe.Width = frame.Width
	size := frame.Height * frame.Width
	data := make([]int, size, size)

	for y := 0; y < frame.Height; y++ {
		for x := 0; x < frame.Width; x++ {
			pixel := frame.GetRGB(x, y)
			lum := equations.RGBtoBrightness(pixel.Red, pixel.Green, pixel.Blue)
			data[y*frame.Width+x] = lum
		}
	}
	lframe.Index = frame.Index
	lframe.Data = data
	return lframe
}

func calculateFrameDifference(f1, f2 brightnessFrame) frameDifference {
	var frameDifference frameDifference
	var maxpos, maxneg int
	positives := make(map[int]int)
	negatives := make(map[int]int)

	for i := 0; i < f1.Height*f1.Width; i++ {
		difference := f2.Data[i] - f1.Data[i]
		if difference > 0 {
			positives[difference]++
			if difference > maxpos {
				maxpos = difference
			}
		} else if difference < 0 {
			difference = -difference
			negatives[difference]++
			if difference > maxneg {
				maxneg = difference
			}
		}
	}
	frameDifference.Height = f1.Height
	frameDifference.Width = f1.Width
	frameDifference.Positives = positives
	frameDifference.Negatives = negatives
	frameDifference.Index = f2.Index
	frameDifference.MaxNeg = maxneg
	frameDifference.MaxPos = maxpos
	return frameDifference
}

func findAverageLuminance(fd frameDifference) int {
	var positive, negative int
	elementsRequired := int(float32(fd.Height*fd.Width) * equations.PercentageFlashArea)
	positive = calculateAverageLuminance(fd.Positives, elementsRequired, fd.MaxPos)
	negative = calculateAverageLuminance(fd.Negatives, elementsRequired, fd.MaxNeg)
	if positive >= negative {
		return positive
	}
	return -negative
}

/*
Calculates the average luminance with a given histogram
The algorithm for calculating the average luminance for a given histogram:
Take the top value elements in the histogram until you have an amount of pixels required for a flash
Then compute the average value of those elements used:
averageLuminance = Sum(luminanceOfElement * amountofElementsWithLuminance) / Sum(amountOfElementsWithLuminance)
*/
func calculateAverageLuminance(histogram map[int]int, elementsRequired, maxLuminance int) int {
	averageDifference := 0

	elementsScanned := 0
	numerator, denominator := 0, 0
	for lumMagnitude := maxLuminance; lumMagnitude >= 0 && elementsScanned != elementsRequired; lumMagnitude-- {
		numberOfPixels, pixelsWithLum := histogram[lumMagnitude]

		if !pixelsWithLum {
			continue
		}

		//If we would go over, just take enough elements to satisfy the requirement
		if numberOfPixels+elementsScanned > elementsRequired {
			numberOfPixels = numberOfPixels - elementsRequired
		}

		numerator += numberOfPixels * lumMagnitude
		denominator += numberOfPixels

		elementsScanned += numberOfPixels
	}

	if denominator == 0 {
		denominator = 1
	}
	averageDifference = numerator / denominator

	return averageDifference
}
