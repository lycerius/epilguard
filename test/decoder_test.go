package test

import (
	"testing"
	"time"

	"github.com/lycerius/epilguard/decoder"
	"github.com/stretchr/testify/assert"
)

const smallVideoFile = "./resources/vid.mp4"
const video720pTest = "./resources/720pTest60fps.mp4"

func createDecoderTestDecoder(file string) decoder.Decoder {
	decoder := decoder.NewDecoder(file)

	return decoder
}
func TestDecoderLoadsFileInfo(t *testing.T) {
	assert := assert.New(t)
	decoder := createDecoderTestDecoder(smallVideoFile)

	err := decoder.Start()
	assert.NoError(err, "Decoder threw error on start")

	assert.Equal(30, decoder.FramesPerSecond)
	assert.Equal(264, decoder.FrameHeight)
	assert.Equal(480, decoder.FrameWidth)
}

func TestDecoderDoesntUpConvert(t *testing.T) {
	assert := assert.New(t)

	decoder := createDecoderTestDecoder(smallVideoFile)

	err := decoder.Start()

	assert.NoError(err, "Could not start decoder")

	assert.Equal(false, decoder.ConvertedTo30FPS)
	assert.Equal(false, decoder.ConvertedTo480p)
}

func TestDecoderDownConverts(t *testing.T) {
	assert := assert.New(t)

	decoder := createDecoderTestDecoder(video720pTest)

	err := decoder.Start()

	assert.NoError(err, "Could not start decoder")

	assert.True(decoder.ConvertedTo30FPS, "Video is not converted to 30fps")
	assert.True(decoder.ConvertedTo480p, "Video is not converted to 480p")
}

func TestDecoderOpens(t *testing.T) {
	assert := assert.New(t)

	decoder := createDecoderTestDecoder(smallVideoFile)

	err := decoder.Start()

	assert.NoError(err, "Could not start decoder")

	assert.True(decoder.IsOpen(), "Decoder reporting closed")
}

func TestDecoderCanGetFrame(t *testing.T) {

	assert := assert.New(t)

	decoder := createDecoderTestDecoder(smallVideoFile)

	err := decoder.Start()

	assert.NoError(err, "Could not start decoder")

	f, err := decoder.NextFrame()

	assert.NotNil(f, "Frame was nil")
}

func TestDecoderCloses(t *testing.T) {
	assert := assert.New(t)

	decoder := createDecoderTestDecoder(smallVideoFile)

	err := decoder.Start()

	assert.NoError(err, "Could not start decoder")

	decoder.Close()
	time.Sleep(1 * time.Second)
	assert.NoError(err, "Could not close decoder")

	assert.False(decoder.IsOpen(), "Decoder reported opened")
}

func TestDecoderClosesInProcess(t *testing.T) {
	assert := assert.New(t)
	decoder := createDecoderTestDecoder(smallVideoFile)
	err := decoder.Start()
	assert.NoError(err, "Could not start decoder")
	decoder.NextFrame()
	decoder.Close()
}

func TestDecoderCanDecodeVideoToEnd(t *testing.T) {

	assert := assert.New(t)

	decoder := createDecoderTestDecoder(smallVideoFile)

	err := decoder.Start()

	assert.NoError(err, "Could not start decoder")

	for {
		frame, err := decoder.NextFrame()

		if err != nil {
			assert.EqualError(err, "EOF", "Error was not EOF")
			return
		}
		frame.GetRGB(0, 0)
	}
}
