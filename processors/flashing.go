package processors

import (
	"container/list"
	"math"
	"time"

	"github.com/lycerius/epilguard/decoder"
	"github.com/lycerius/epilguard/equations"
	"github.com/lycerius/epilguard/hazards"
)

//FlashingProcessor Processes a video stream and detects flashing photosensitive content
type FlashingProcessor struct {
	decoder       *decoder.Decoder     //Decoder to fetch frames from
	CSVDirectory  string               //The job assosiated with this request
	HazardReport  hazards.HazardReport //Generated hazard report
	AreaThreshold float32
}

//brightnessFrame Describes the pixel brightness transition between two frames
type brightnessFrame struct {
	Index         uint
	Pixels        []int
	Height, Width int
}

//frameBrightnessDelta is like brightnessFrame, but the pixels are organized into positive/negative bins
type frameBrightnessDelta struct {
	Index                          uint
	Height, Width                  int
	NegativePixels, PositivePixels map[int]int
	MaxPos, MaxNeg                 int
}

type BrightnessAccumulationTable = *list.List

//BrightnessAccumulation holds the average brightness and the total accumulation over time
type BrightnessAccumulation struct {
	Index                    uint
	Brightness, Accumulation int
}

type FlashTable = *list.List

//Flash describes the maximum brightness achieved over a set of frames before an inversion
type Flash struct {
	Brightness, Frames int
}

//NewFlashingProcessor creates a flashing processor
func NewFlashingProcessor(f *decoder.Decoder, csvDir string) FlashingProcessor {
	var processor FlashingProcessor

	processor.decoder = f
	processor.CSVDirectory = csvDir

	return processor
}

//Process begins scanning the video for flashing photosensitive content
func (proc *FlashingProcessor) Process() error {
	now := time.Now()
	brightnessAcc, err := createBrightnessAccumulationTable(proc.decoder)

	if err != nil {
		return err
	}

	err = ExportBrightnessAccumulation(proc.decoder.FileName, proc.CSVDirectory, brightnessAcc, now)

	if err != nil {
		return err
	}

	flashes := createFlashTable(brightnessAcc)

	err = ExportFlashTable(proc.decoder.FileName, proc.CSVDirectory, flashes, now)

	if err != nil {
		return err
	}

	err = ExportFlashTableByFrames(proc.decoder.FileName, proc.CSVDirectory, flashes, now)

	if err != nil {
		return err
	}

	report := createHazardReport(flashes, proc.decoder.FramesPerSecond)

	report.CreatedOn = time.Now()

	proc.HazardReport = report

	return nil
}

func createBrightnessAccumulationTable(decoder *decoder.Decoder) (BrightnessAccumulationTable, error) {
	brightnessAcc := list.New()
	frame, err := decoder.NextFrame()

	if err != nil {
		return nil, err
	}

	var accBrightness int
	firstFrame := rGBFrameToBrightness(frame)
	lastFrame := &firstFrame

	for {
		frame, err := decoder.NextFrame()

		if err != nil {
			if err.Error() == "EOF" {
				break
			} else {
				return nil, err
			}
		}

		brightnessFrame := rGBFrameToBrightness(frame)
		difference := calculateFrameDifference(*lastFrame, brightnessFrame)
		averageBrightness := findAverageBrightness(difference)

		//Check if signs are different and no 0 value
		if (accBrightness < 0) == (averageBrightness < 0) || averageBrightness == 0 {
			accBrightness += averageBrightness
		} else {
			accBrightness = averageBrightness
		}

		lastFrame = &brightnessFrame

		var accumulation BrightnessAccumulation
		accumulation.Index = frame.Index
		accumulation.Accumulation = accBrightness
		accumulation.Brightness = averageBrightness

		brightnessAcc.PushBack(accumulation)
	}

	return brightnessAcc, nil
}

//In the future, we may want to parallize this
func rGBFrameToBrightness(frame decoder.Frame) brightnessFrame {

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
	lframe.Pixels = data
	return lframe
}

func calculateFrameDifference(f1, f2 brightnessFrame) frameBrightnessDelta {
	var frameDifference frameBrightnessDelta
	var maxpos, maxneg int
	positives := make(map[int]int)
	negatives := make(map[int]int)

	for i := 0; i < f1.Height*f1.Width; i++ {
		difference := f2.Pixels[i] - f1.Pixels[i]
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
	frameDifference.PositivePixels = positives
	frameDifference.NegativePixels = negatives
	frameDifference.Index = f2.Index
	frameDifference.MaxNeg = maxneg
	frameDifference.MaxPos = maxpos
	return frameDifference
}

func findAverageBrightness(fd frameBrightnessDelta) int {
	elementsRequired := int(float32(fd.Height*fd.Width) * equations.PercentageFlashArea)
	positive := calculateAverageLuminance(fd.PositivePixels, elementsRequired, fd.MaxPos)
	negative := calculateAverageLuminance(fd.NegativePixels, elementsRequired, fd.MaxNeg)

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
func calculateAverageLuminance(histogram map[int]int, pixelsRequired, maxBrightness int) int {

	var pixelsScanned, accBrightness, averageDifference int

	for brightness := maxBrightness; brightness > 0 && pixelsScanned < pixelsRequired; brightness-- {
		pixelsWithBrightness, brightnessExists := histogram[brightness]

		if !brightnessExists {
			continue
		}

		//If we would go over, just take enough elements to satisfy the requirement
		if pixelsWithBrightness+pixelsScanned > pixelsRequired {
			pixelsWithBrightness = pixelsRequired - pixelsScanned
		}

		accBrightness += pixelsWithBrightness * brightness
		pixelsScanned += pixelsWithBrightness
	}

	averageDifference = accBrightness / pixelsRequired
	return averageDifference
}

func createFlashTable(brightnessAcc BrightnessAccumulationTable) FlashTable {
	flashTable := list.New()

	localMaxima := (brightnessAcc.Front().Value.(BrightnessAccumulation)).Accumulation
	var amountOfFrames int

	brightnessElement := brightnessAcc.Front()

	for {
		if brightnessElement == nil {
			var extreme Flash
			extreme.Frames = amountOfFrames
			extreme.Brightness = localMaxima
			flashTable.PushBack(extreme)
			break
		}

		accumulation := brightnessElement.Value.(BrightnessAccumulation)

		brightness := accumulation.Accumulation

		if (brightness < 0) == (localMaxima < 0) {
			amountOfFrames++
			if math.Abs(float64(localMaxima)) < math.Abs(float64(brightness)) {
				localMaxima = brightness
			}
		} else {
			var extreme Flash
			extreme.Frames = amountOfFrames
			extreme.Brightness = localMaxima
			flashTable.PushBack(extreme)
			amountOfFrames = 1
			localMaxima = brightness
		}

		brightnessElement = brightnessElement.Next()
	}

	return flashTable
}

func createHazardReport(lumExtTab FlashTable, fps int) hazards.HazardReport {
	var hazardReport hazards.HazardReport

	flashesPerSecondThreshold := 3
	frameCounter := 0
	countedFlashes := 0
	currentFrameIndex := 0
	flashStartIndex := -1
	previousLuminance := (lumExtTab.Front().Value.(Flash)).Brightness
	for lumExtremeElement := lumExtTab.Front(); lumExtremeElement != nil; lumExtremeElement = lumExtremeElement.Next() {

		lumExtreme := lumExtremeElement.Value.(Flash)

		currentFrameIndex += lumExtreme.Frames
		previousLuminanceAbs := int(math.Abs(float64(previousLuminance)))
		currentLuminance := lumExtreme.Brightness
		currentLuminanceAbs := int(math.Abs(float64(currentLuminance)))

		var darkerLuminance int
		if previousLuminance < 0 {
			darkerLuminance = previousLuminanceAbs
		} else {
			darkerLuminance = currentLuminanceAbs
		}

		//We are currently checking for flashes
		if flashStartIndex != -1 {
			frameCounter += lumExtreme.Frames
		}

		//Has to be a difference of 20 or more candellas, and darker frame must be below 160
		if math.Abs(float64(currentLuminance-previousLuminance)) > 20 && darkerLuminance < 160 {
			if flashStartIndex == -1 {
				//Start detecting flashes
				flashStartIndex = currentFrameIndex
			}
			countedFlashes++
		}

		//We have surpassed 1 second after checking for flashes, check to see if we need to make a report
		if frameCounter >= fps || lumExtremeElement.Next() == nil {

			//Crossed threshold
			if countedFlashes >= flashesPerSecondThreshold {
				var hazard hazards.Hazard
				hazard.Start = uint(float64(flashStartIndex) / float64(fps))
				hazard.End = uint(float64(currentFrameIndex) / float64(fps))
				hazard.HazardType = "Flash"
				hazardReport.Hazards.PushBack(hazard)
			}

			//Reset
			flashStartIndex = -1
			frameCounter = 0
			countedFlashes = 0
		}

		previousLuminance = currentLuminance
	}

	hazardReport.Hazards = consolidateHazardList(hazardReport.Hazards)

	return hazardReport
}

func consolidateHazardList(li hazards.HazardList) hazards.HazardList {

	if (li.Front()) == nil {
		return li
	}

	var consolidated hazards.HazardList

	ele := li.Front()
	val := ele.Value.(hazards.Hazard)

	var temp hazards.Hazard
	temp.Start = val.Start
	temp.End = val.End
	temp.HazardType = val.HazardType

	for ele = ele.Next(); ele != nil; ele = ele.Next() {
		val = ele.Value.(hazards.Hazard)

		if val.Start == temp.End {
			temp.End = val.End
		} else {
			consolidated.PushBack(temp)
			temp = hazards.Hazard{}
			temp.Start = val.Start
			temp.End = val.End
			temp.HazardType = val.HazardType
		}
	}
	consolidated.PushBack(temp)
	return consolidated
}
