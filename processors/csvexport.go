package processors

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//newCSVFile creates a new CSV file and returns a writeable stream
func newCSVFile(path string) (*csv.Writer, error) {

	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	writer := csv.NewWriter(file)
	return writer, nil
}

func GenerateCSVFileName(videoName, csvDir, datasetName string, date time.Time) string {
	videoName = filepath.Base(videoName)
	lastIndexOfDot := strings.LastIndex(videoName, ".")
	if lastIndexOfDot == -1 {
		lastIndexOfDot = len(videoName)
	}
	videoName = videoName[0:lastIndexOfDot]
	clamp := len(videoName)
	if clamp > 20 {
		clamp = 20
	}

	normalName := strings.Replace(videoName[:clamp], " ", "-", -1)
	return filepath.Join(csvDir, strconv.FormatUint(uint64(date.Unix()), 16)+"-"+normalName+"-"+datasetName+".csv")
}

func ExportBrightnessAccumulation(path string, csvDir string, accTab BrightnessAccumulationTable, date time.Time) error {
	path = GenerateCSVFileName(path, csvDir, "Accumulation", date)
	csv, err := newCSVFile(path)

	if err != nil {
		return err
	}
	defer csv.Flush()

	//Write the header
	csv.Write([]string{"Index", "Brightness", "Accumulation"})

	//Write the values
	for tableElement := accTab.Front(); tableElement != nil; tableElement = tableElement.Next() {
		brightnessAccumulation := tableElement.Value.(BrightnessAccumulation)
		index := strconv.FormatUint(uint64(brightnessAccumulation.Index), 10)
		brightness := strconv.Itoa(brightnessAccumulation.Brightness)
		accumulation := strconv.Itoa(brightnessAccumulation.Accumulation)
		csv.Write([]string{index, brightness, accumulation})
	}
	return nil
}

func ExportFlashTable(path string, csvDir string, flashTable FlashTable, date time.Time) error {
	path = GenerateCSVFileName(path, csvDir, "Flashes", date)
	csv, err := newCSVFile(path)

	if err != nil {
		return err
	}
	defer csv.Flush()

	csv.Write([]string{"Brightness", "Frames"})

	for tableElement := flashTable.Front(); tableElement != nil; tableElement = tableElement.Next() {
		flash := tableElement.Value.(Flash)
		brightness := strconv.Itoa(flash.Brightness)
		frames := strconv.Itoa(flash.Frames)
		csv.Write([]string{brightness, frames})
	}

	return nil
}

func ExportFlashTableByFrames(path string, csvDir string, flashTable FlashTable, date time.Time) error {
	path = GenerateCSVFileName(path, csvDir, "FrameFlashes", date)
	csv, err := newCSVFile(path)

	if err != nil {
		return err
	}
	defer csv.Flush()

	csv.Write([]string{"FrameIndex", "Brightness"})

	var frameIndex uint64
	for tableElement := flashTable.Front(); tableElement != nil; tableElement = tableElement.Next() {
		flash := tableElement.Value.(Flash)
		brightness := strconv.Itoa(flash.Brightness)
		frameCount := flash.Frames
		for i := 0; i < frameCount; i++ {
			frameIndexS := strconv.FormatUint(frameIndex, 10)
			csv.Write([]string{frameIndexS, brightness})
			frameIndex++
		}
	}
	return nil
}