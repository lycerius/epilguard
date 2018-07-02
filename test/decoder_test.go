package main

import (
	"testing"
	"time"

	"github.com/lycerius/epilguard/tools"
	"github.com/stretchr/testify/assert"
)

const smallVideoFile = "./resources/vid.mp4"
const video720pTest = "./resources/720pTest60fps.mp4"

func TestDecoderLoadsFileInfo(t *testing.T) {
	assert := assert.New(t)
	decoder := tools.NewDecoder(smallVideoFile)

	err := decoder.Start()

	assert.NoError(err, "Decoder threw error on start")

	assert.Equal(30, decoder.Fps)
	assert.Equal(264, decoder.FrameHeight)
	assert.Equal(480, decoder.FrameWidth)
}

func TestDecoderDoesntUpConvert(t *testing.T) {
	assert := assert.New(t)

	decoder := tools.NewDecoder(smallVideoFile)

	err := decoder.Start()

	assert.NoError(err, "Could not start decoder")

	assert.Equal(false, decoder.ConvertTo30FPS)
	assert.Equal(false, decoder.ConvertTo480p)
}

func TestDecoderDownConverts(t *testing.T) {
	assert := assert.New(t)

	decoder := tools.NewDecoder(video720pTest)

	err := decoder.Start()

	assert.NoError(err, "Could not start decoder")

	assert.True(decoder.ConvertTo30FPS, "Video is not converted to 30fps")
	assert.True(decoder.ConvertTo480p, "Video is not converted to 480p")
}

func TestDecoderOpens(t *testing.T) {
	assert := assert.New(t)

	decoder := tools.NewDecoder(smallVideoFile)

	err := decoder.Start()

	assert.NoError(err, "Could not start decoder")

	assert.True(decoder.Opened, "Decoder reporting closed")
}

func TestDecoderCanGetFrame(t *testing.T) {

	assert := assert.New(t)

	decoder := tools.NewDecoder(smallVideoFile)

	err := decoder.Start()

	assert.NoError(err, "Could not start decoder")

	f := decoder.Next()

	assert.NotNil(f, "Frame was nil")
}

func TestDecoderCloses(t *testing.T) {
	assert := assert.New(t)

	decoder := tools.NewDecoder(smallVideoFile)

	err := decoder.Start()

	assert.NoError(err, "Could not start decoder")

	err = decoder.Close()
	time.Sleep(1 * time.Second)
	assert.NoError(err, "Could not close decoder")

	assert.False(decoder.Opened, "Decoder reported opened")
	assert.Nil(decoder.FrameBuffer, "Frame buffer not nil")
}

func TestDecoderFrameBuffer(t *testing.T) {
	assert := assert.New(t)

	decoder := tools.NewDecoder(smallVideoFile)

	err := decoder.Start()

	assert.NoError(err, "Could not start decoder")
	time.Sleep(1 * time.Second)
	assert.Equal(decoder.FrameBufferSize, len(decoder.FrameBuffer))
}
