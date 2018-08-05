package processors

import (
	"bufio"
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lycerius/epilguard/hazards"
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

//newFile creates a file at path and returns a writer to it
func newFile(path string) (*bufio.Writer, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	writer := bufio.NewWriter(file)
	return writer, err
}

func generateCSVFileName(videoName, csvDir, datasetName string, date time.Time) string {
	return generateExportItemFileName(videoName, csvDir, datasetName, date) + ".csv"
}

func generateJSONFileName(videoName, csvDir, datasetName string, date time.Time) string {
	return generateExportItemFileName(videoName, csvDir, datasetName, date) + ".json"
}

func generateExportItemFileName(videoName, csvDir, datasetName string, date time.Time) string {
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
	return filepath.Join(csvDir, strconv.FormatUint(uint64(date.Unix()), 16)+"-"+normalName+"-"+datasetName)
}

//ExportBrightnessAccumulation creates a csv using accTab at csvDir using the name of the video at path and the current time date
func ExportBrightnessAccumulation(path string, csvDir string, accTab BrightnessAccumulationTable, date time.Time) error {
	path = generateCSVFileName(path, csvDir, "Accumulation", date)
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

//ExportFlashTable creates a csv using flashTable at csvDir using the name of the video at path and the current time date
func ExportFlashTable(path string, csvDir string, flashTable FlashTable, date time.Time) error {
	path = generateCSVFileName(path, csvDir, "Flashes", date)
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

//ExportFlashTableByFrames creates a csv using flastTable at csvDir using the name of the video at path and the current time date
func ExportFlashTableByFrames(path string, csvDir string, flashTable FlashTable, date time.Time) error {
	path = generateCSVFileName(path, csvDir, "FrameFlashes", date)
	csv, err := newCSVFile(path)

	if err != nil {
		return err
	}
	defer csv.Flush()

	csv.Write([]string{"FrameIndex", "Brightness"})

	var frameIndex uint64 = 1
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

//ExportHazardReport creates a json file using report  at csvDir using the name of the video at path and the current time date
func ExportHazardReport(path, csvDir string, report hazards.HazardReport, date time.Time) error {
	path = generateJSONFileName(path, csvDir, "Report", date)
	file, err := newFile(path)
	if err != nil {
		return err
	}

	defer file.Flush()
	json, err := report.MarshalJSON()

	if err != nil {
		return err
	}

	_, err = file.Write(json)

	return err
}
