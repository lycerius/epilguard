package test

import (
	"testing"

	"github.com/lycerius/epilguard/hazards"

	"github.com/lycerius/epilguard/decoder"
	"github.com/lycerius/epilguard/processors"
	"github.com/stretchr/testify/assert"
)

const porygon = "./resources/porygon.mp4"
const h3h3 = "./resources/h3h3.mp4"

func createDecoder(file string) (decoder.Decoder, error) {
	decoder := decoder.NewDecoder(file)

	return decoder, decoder.Start()
}

func createProcessor(file string) (processors.FlashingProcessor, error) {
	var processor processors.FlashingProcessor
	decoder, err := createDecoder(file)

	if err != nil {
		return processor, err
	}

	processor = processors.NewFlashingProcessor(&decoder, "test job")

	return processor, nil
}

func TestProcessorCanInitialize(t *testing.T) {
	assert := assert.New(t)
	_, err := createProcessor(smallVideoFile)

	assert.NoError(err, "Error during initialization")
}

func TestProcessorCanProcessSmallVideo(t *testing.T) {
	assert := assert.New(t)

	proc, err := createProcessor(smallVideoFile)

	assert.NoError(err, "Error during initialization")

	err = proc.Process()

	assert.NoError(err, "Error during processing")
}

func TestProcessorCanProcessLargeVideo(t *testing.T) {
	assert := assert.New(t)

	proc, err := createProcessor(video720pTest)

	assert.NoError(err, "Error during initialization")

	err = proc.Process()
}

func TestProcessorFailsPorygon(t *testing.T) {
	assert := assert.New(t)

	proc, err := createProcessor(porygon)

	assert.NoError(err, "Error during initialization")

	err = proc.Process()
}

func TestProcessorFailsH3H3(t *testing.T) {
	assert := assert.New(t)

	proc, err := createProcessor(h3h3)

	assert.NoError(err, "Error during initialization")

	err = proc.Process()
}

func TestUploadResult(t *testing.T) {
	assert := assert.New(t)

	proc, err := createProcessor(smallVideoFile)

	assert.NoError(err, "Error during initialization")

	err = proc.Process()

	assert.NoError(err, "Error occured during processing")
	report := proc.HazardReport

	hazards.UploadHazardReport(report)

}
