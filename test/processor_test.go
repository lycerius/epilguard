package test

import (
	"testing"

	"github.com/lycerius/epilguard/processors"
	"github.com/lycerius/epilguard/tools"
	"github.com/stretchr/testify/assert"
)

func createDecoder(file string) (tools.FFMPEGDecoder, error) {
	decoder := tools.NewDecoder(file)

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
