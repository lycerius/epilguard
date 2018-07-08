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
	JobID         string               //The job assosiated with this request
	HazardReport  hazards.HazardReport //Generated hazard report
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

type luminanceEvolutionTable = *list.List

type luminanceEvolution struct {
	index                         uint
	lumMagnitude, lumAccumulation int
}

type luminanceExtremeTable = *list.List
type luminanceExtreme struct {
	magnitude, frameCount int
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
	evolution, err := createLuminanceEvolutionTable(proc.decoder)

	if err != nil {
		return err
	}

	extremes := createLocalExtremesTable(evolution)

	report := createHazardReport(extremes, proc.decoder.FramesPerSecond)

	report.JobID = proc.JobID
	report.CreatedOn = time.Now()

	return nil
}

func createLuminanceEvolutionTable(decoder *decoder.Decoder) (luminanceEvolutionTable, error) {
	luminanceEvolutionTable := list.New()
	frame, err := decoder.NextFrame()

	if err != nil {
		return nil, err
	}

	var accLuminance int
	firstFrame := rGBFrameToLuminance(frame)
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

		lumFrame := rGBFrameToLuminance(frame)
		difference := calculateFrameDifference(*lastFrame, lumFrame)
		averageLuminance := findAverageLuminance(difference)

		//Check if signs are different
		if (accLuminance < 0) == (averageLuminance < 0) {
			accLuminance += averageLuminance
		} else {
			accLuminance = averageLuminance
		}

		lastFrame = &lumFrame

		var evolution luminanceEvolution
		evolution.index = frame.Index
		evolution.lumAccumulation = accLuminance
		evolution.lumMagnitude = averageLuminance

		luminanceEvolutionTable.PushBack(evolution)
	}

	return luminanceEvolutionTable, nil
}

//In the future, we may want to parallize this
func rGBFrameToLuminance(frame decoder.Frame) brightnessFrame {

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
	averageDifference = numerator / elementsRequired

	return averageDifference
}

func createLocalExtremesTable(lumAcc luminanceEvolutionTable) luminanceExtremeTable {
	luminanceExtremes := list.New()

	localMaxima := (lumAcc.Front().Value.(luminanceEvolution)).lumAccumulation
	var amountOfFrames int

	lum := lumAcc.Front()

	for {
		if lum == nil {
			var extreme luminanceExtreme
			extreme.frameCount = amountOfFrames
			extreme.magnitude = localMaxima
			luminanceExtremes.PushBack(extreme)
			break
		}

		evolution := lum.Value.(luminanceEvolution)

		lumVal := evolution.lumAccumulation

		if (lumVal < 0) == (localMaxima < 0) {
			amountOfFrames++
			if math.Abs(float64(localMaxima)) < math.Abs(float64(lumVal)) {
				localMaxima = lumVal
			}
		} else {
			var extreme luminanceExtreme
			extreme.frameCount = amountOfFrames
			extreme.magnitude = localMaxima
			luminanceExtremes.PushBack(extreme)
			amountOfFrames = 1
			localMaxima = lumVal
		}

		lum = lum.Next()
	}

	return luminanceExtremes
}

func createHazardReport(lumExtTab luminanceExtremeTable, fps int) hazards.HazardReport {
	var hazardReport hazards.HazardReport

	flashesPerSecondThreshold := 3
	frameCounter := 0
	countedFlashes := 0
	currentFrameIndex := 0
	flashStartIndex := -1
	previousLuminance := (lumExtTab.Front().Value.(luminanceExtreme)).magnitude

	for lumExtremeElement := lumExtTab.Front(); lumExtremeElement != nil; lumExtremeElement = lumExtremeElement.Next() {

		lumExtreme := lumExtremeElement.Value.(luminanceExtreme)

		currentFrameIndex += lumExtreme.frameCount
		previousLuminanceAbs := int(math.Abs(float64(previousLuminance)))
		currentLuminance := lumExtreme.magnitude
		currentLuminanceAbs := int(math.Abs(float64(currentLuminance)))

		var darkerLuminance int
		if previousLuminance < 0 {
			darkerLuminance = previousLuminanceAbs
		} else {
			darkerLuminance = currentLuminanceAbs
		}

		//Has to be a difference of 20 or more candellas, and darker frame must be below 160
		if (currentLuminanceAbs-previousLuminanceAbs) > 20 && darkerLuminance < 160 {
			if flashStartIndex == -1 {
				//Start detecting flashes
				flashStartIndex = currentFrameIndex
			}
			countedFlashes++
		}

		//We are currently checking for flashes
		if flashStartIndex != -1 {
			frameCounter += lumExtreme.frameCount
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
