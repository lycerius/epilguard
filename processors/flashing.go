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
	decoder         *decoder.Decoder     //Decoder to fetch frames from
	ReportDirectory string               //The job assosiated with this request
	HazardReport    hazards.HazardReport //Generated hazard report
	AreaThreshold   float32
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

//BrightnessAccumulationTable a list of Brightness Accumulations
type BrightnessAccumulationTable = *list.List

//BrightnessAccumulation holds the average brightness and the total accumulation over time
type BrightnessAccumulation struct {
	Index                    uint
	Brightness, Accumulation int
}

//FlashTable is a list of flashes
type FlashTable = *list.List

//Flash describes the maximum brightness achieved over a set of frames before an inversion
type Flash struct {
	Brightness, Frames int
}

//NewFlashingProcessor creates a flashing processor
func NewFlashingProcessor(f *decoder.Decoder, reportDir string) FlashingProcessor {
	var processor FlashingProcessor

	processor.decoder = f
	processor.ReportDirectory = reportDir

	return processor
}

//Process scans a video for photosensitive content and exports it to reportDir
func (proc *FlashingProcessor) Process() error {

	brightnessAcc, err := createBrightnessAccumulationTable(proc.decoder)
	if err != nil {
		return err
	}

	flashes := createFlashTable(brightnessAcc)

	report := createHazardReport(flashes, proc.decoder.FramesPerSecond)
	report.CreatedOn = time.Now()

	proc.exportReport(brightnessAcc, flashes, report)
	proc.HazardReport = report

	return nil
}

//exportReport exports the report to ReportDirectory
func (proc *FlashingProcessor) exportReport(brightnessAcc BrightnessAccumulationTable, flashes FlashTable, report hazards.HazardReport) error {
	now := time.Now()

	err := ExportBrightnessAccumulation(proc.decoder.FileName, proc.ReportDirectory, brightnessAcc, now)
	if err != nil {
		return err
	}

	err = ExportFlashTable(proc.decoder.FileName, proc.ReportDirectory, flashes, now)
	if err != nil {
		return err
	}

	err = ExportFlashTableByFrames(proc.decoder.FileName, proc.ReportDirectory, flashes, now)
	if err != nil {
		return err
	}

	err = ExportHazardReport(proc.decoder.FileName, proc.ReportDirectory, report, now)
	if err != nil {
		return err
	}

	return nil
}

//createBrightnessAccumulationTable decodes all frames and creates a brightness accumulation table
func createBrightnessAccumulationTable(decoder *decoder.Decoder) (BrightnessAccumulationTable, error) {
	brightnessAcc := list.New()

	//First frame for baseline brightness
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
			//This is an OK error, just EOF
			if err.Error() == "EOF" {
				break
			} else {
				return nil, err
			}
		}

		//Calculations
		brightnessFrame := rGBFrameToBrightness(frame)
		difference := calculateFrameDifference(*lastFrame, brightnessFrame)
		averageBrightness := findAverageBrightness(difference)

		//If signs are equal, or no change, accumulate
		if (accBrightness < 0) == (averageBrightness < 0) || averageBrightness == 0 {
			accBrightness += averageBrightness
		} else {
			//If signs are different, then an inversion occured. Start new accumulation
			accBrightness = averageBrightness
		}

		lastFrame = &brightnessFrame

		//Create new entry
		var accumulation BrightnessAccumulation
		accumulation.Index = frame.Index
		accumulation.Accumulation = accBrightness
		accumulation.Brightness = averageBrightness
		brightnessAcc.PushBack(accumulation)
	}

	return brightnessAcc, nil
}

//rGBFrameToBrightness converts an RGB frame to brightness
func rGBFrameToBrightness(frame decoder.Frame) brightnessFrame {

	var lframe brightnessFrame
	lframe.Height = frame.Height
	lframe.Width = frame.Width
	size := frame.Height * frame.Width

	pixelBuffer := make([]int, size, size)

	for y := 0; y < frame.Height; y++ {
		for x := 0; x < frame.Width; x++ {
			pixel := frame.GetRGB(x, y)
			brightness := equations.RGBtoBrightness(pixel.Red, pixel.Green, pixel.Blue)
			pixelBuffer[y*frame.Width+x] = brightness
		}
	}

	lframe.Index = frame.Index
	lframe.Pixels = pixelBuffer
	return lframe
}

//calculateFrameDifference takes 2 brightness frames and calculates brightness change per pixel
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

//findAverageBrightness takes the calculated brightness differences and chooses the positive or negative bin
//depending on which bin has the largest magnitude
func findAverageBrightness(fd frameBrightnessDelta) int {
	elementsRequired := int(float32(fd.Height*fd.Width) * equations.PercentageFlashArea)
	positive := calculateAverageBrightness(fd.PositivePixels, elementsRequired, fd.MaxPos)
	negative := calculateAverageBrightness(fd.NegativePixels, elementsRequired, fd.MaxNeg)

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
func calculateAverageBrightness(histogram map[int]int, pixelsRequired, maxBrightness int) int {

	var pixelsScanned, accBrightness, averageDifference int

	//Start at the top, going down in brightness checks
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

//createFlashTable takes brightness accumulations and compresses it to just inversions
//and how many frames the brightness trend lasted before the inversion
func createFlashTable(brightnessAcc BrightnessAccumulationTable) FlashTable {
	flashTable := list.New()

	localMaxima := (brightnessAcc.Front().Value.(BrightnessAccumulation)).Accumulation
	var amountOfFrames int

	brightnessElement := brightnessAcc.Front()

	for {
		//No more elements, push the last entry
		if brightnessElement == nil {
			var extreme Flash
			extreme.Frames = amountOfFrames
			extreme.Brightness = localMaxima
			flashTable.PushBack(extreme)
			break
		}

		accumulation := brightnessElement.Value.(BrightnessAccumulation)
		brightness := accumulation.Accumulation

		//Signs are equal, trend continues
		if (brightness < 0) == (localMaxima < 0) {
			amountOfFrames++
			if math.Abs(float64(localMaxima)) < math.Abs(float64(brightness)) {
				localMaxima = brightness
			}
		} else {
			//Inversion occured
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

	flashesPerSecondThreshold := equations.FlashFrequencyMax
	frameCounter := 0
	countedFlashes := 0
	currentFrameIndex := 1
	flashStartIndex := -1
	previousLuminance := (lumExtTab.Front().Value.(Flash)).Brightness
	for lumExtremeElement := lumExtTab.Front(); lumExtremeElement != nil; lumExtremeElement = lumExtremeElement.Next() {

		lumExtreme := lumExtremeElement.Value.(Flash)

		currentFrameIndex += lumExtreme.Frames
		currentLuminance := lumExtreme.Brightness
		currentLuminanceAbs := int(math.Abs(float64(currentLuminance)))

		var darkerLuminance int
		if previousLuminance < 0 {
			darkerLuminance = int(math.Abs(float64(previousLuminance)))
		} else {
			darkerLuminance = currentLuminanceAbs
		}

		//We are currently checking for flashes
		if flashStartIndex != -1 {
			frameCounter += lumExtreme.Frames
		}

		//Has to be a difference of 20 or more candellas, and darker frame must be below 160
		if currentLuminanceAbs >= 20 && darkerLuminance < 160 {
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

//consolidateHazardList compresses consecutive entries down into a hazard
//consecutive entries are hazards where the previous hazard end is equal to the current hazard start
//an example of a consecutive entry:
//hazard1 {Start=0; End=3}, hazard2 {Start=3; End=6} = hazard {Start=0; End=6}
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
