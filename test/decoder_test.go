package test

import (
	"testing"

	"github.com/lycerius/epilguard/decoder"
	"github.com/stretchr/testify/assert"
)

func createDecoderTestDecoder(file string, t *assert.Assertions) decoder.Decoder {
	decoder := decoder.NewDecoder(file)
	err := decoder.Start()
	t.NoError(err, err)
	return decoder
}

func TestDecoderLoadsFileInfo(t *testing.T) {
	assert := assert.New(t)
	decoder := createDecoderTestDecoder(Test_Video_White, assert)

	assert.Equal(24, decoder.FramesPerSecond)
	assert.Equal(854, decoder.FrameWidth)
	assert.Equal(480, decoder.FrameHeight)
}

func TestDecoderCanGetFrame(t *testing.T) {
	assert := assert.New(t)
	decoder := createDecoderTestDecoder(Test_Video_White, assert)

	decoder.Start()
	frame, err := decoder.NextFrame()

	assert.NoError(err, "Unable to retrieve frame", err)

	assert.Equal(frame.Height, decoder.FrameHeight)
	assert.Equal(frame.Width, decoder.FrameWidth)
}

func TestDecoderCanDecodeFullVideo(t *testing.T) {
	assert := assert.New(t)
	decoder := createDecoderTestDecoder(Test_Video_White, assert)

	decoder.Start()

	var err error
	for _, err = decoder.NextFrame(); err == nil; _, err = decoder.NextFrame() {

	}

	if !assert.Error(err, "Error expected here") {
		assert.FailNow("Error is nil")
	}

	assert.Equal(err.Error(), "EOF", "Unexpected error occured", err)
}
