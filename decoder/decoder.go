package decoder

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

//Command line magic for ffmpeg and ffprobe
const _FFProbeCommnand = "ffprobe"
const _FFProbeArgs = "[filename] -v quiet -print_format json -show_format -show_streams"
const _FFMPEGCommand string = "ffmpeg"
const _FFMPEGArgs string = "-i [filename] -an -pix_fmt rgb24 -c:v rawvideo -map 0:v -f image2pipe -"
const _FFMPEGArgs480p = "-s hd480"
const _FFMPEGArgs30fps = "-r 30 -framerate 30"

const _FrameBufferDefaultSize = 30

//Resolution and FPS Finding Regex
var resolutionRegex = regexp.MustCompile(`rgb24, (\d*)x(\d*)`)
var fpsRegex = regexp.MustCompile(`(\d*) fps`)

//Decoder Video decoder with ffmpeg as the frame source
type Decoder struct {
	FileName                string
	FrameWidth, FrameHeight int
	FramesPerSecond         int
	FrameBufferCacheSize    int
	ConvertedTo30FPS        bool
	ConvertedTo480p         bool
	opened                  bool
	decoderOpened           bool
	caching                 bool
	cmdString               string
	ffmpegProcess           *exec.Cmd
	frameSource             io.ReadCloser
	frameBuffer             chan Frame
	signalUserCloseDecoder  chan interface{}
	signalDecoderClosed     chan interface{}
	rawFrameSize            int
}

//NewDecoder Creates a new video decoder for the given file
func NewDecoder(fileName string) Decoder {
	var decoder Decoder
	decoder.FileName = fileName
	decoder.cmdString = _FFMPEGArgs
	decoder.FrameBufferCacheSize = _FrameBufferDefaultSize
	return decoder
}

//Start opens the stream and begins decoding the video
func (f *Decoder) Start() error {

	//Already in process
	if f.IsOpen() {
		return errors.New("Decoder has already been started")
	}

	//Check if file exists
	if _, err := os.Stat(f.FileName); err != nil {
		return err
	}

	fileHeight, fileFps, err := probeFileInformation(f.FileName)

	if err != nil {
		return err
	}

	f.ConvertedTo30FPS = fileFps > 30
	f.ConvertedTo480p = fileHeight > 480

	arguments := createFFMPegArguments(f.FileName, f.ConvertedTo30FPS, f.ConvertedTo480p)

	ffmpegProcess := exec.Command(_FFMPEGCommand, arguments...)
	f.ffmpegProcess = ffmpegProcess

	stdout, err := ffmpegProcess.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := ffmpegProcess.StderrPipe()
	if err != nil {
		return err
	}

	err = ffmpegProcess.Start()
	if err != nil {
		return err
	}

	height, width, fps, err := probeStreamInfo(stderr)
	stderr.Close()
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	f.FrameHeight = height
	f.FrameWidth = width
	f.FramesPerSecond = fps

	f.frameSource = stdout
	f.rawFrameSize = f.FrameHeight * f.FrameWidth * 3
	f.frameBuffer = make(chan Frame, f.FrameBufferCacheSize)
	f.signalDecoderClosed = make(chan interface{}, 1)
	f.signalUserCloseDecoder = make(chan interface{}, 1)

	//Concurrently fill the framebuffer
	go cacheFrameBuffer(f)
	f.opened = true
	return nil
}

//IsOpen returns if the decoder stream is currently open
func (f *Decoder) IsOpen() bool {
	return f.opened || len(f.frameBuffer) > 0
}

//Close Closes the video decoder
func (f *Decoder) Close() {
	if f.IsOpen() {
		f.signalUserCloseDecoder <- nil
	}
}

//NextFrame gets the next frame of the video
func (f *Decoder) NextFrame() (Frame, error) {

	//We may be done decoding frames
	if len(f.frameBuffer) == 0 {
		select {
		//Decoder may be producing frames
		case fr := <-f.frameBuffer:
			return fr, nil
		//Signaled by decoder that no more frames will be produced
		case <-f.signalDecoderClosed:
			//So no more frames left
			return Frame{}, errors.New("EOF")
		}
	}

	//We arnt empty, so who cares if ffmpeg is still running
	return <-f.frameBuffer, nil
}

//cacheFrameBuffer decodes video frames from ffmpeg and places them in the buffer
func cacheFrameBuffer(f *Decoder) {
	var fIndex uint
	frameBuffer := f.frameBuffer
	f.caching = true
	for f.IsOpen() {
		if frame, err := f.nextSourceFrame(); err == nil {
			frame.Index = fIndex
			fIndex++
			select {
			case <-f.signalUserCloseDecoder:
				f.finalize()
				break
			case frameBuffer <- frame:
			}

		} else {
			f.opened = false
			break
		}
	}
	f.caching = false
	f.signalDecoderClosed <- nil
}

//nextFrame Gets the next frame in a stream
func (f *Decoder) nextSourceFrame() (Frame, error) {
	var frame Frame
	amountToGrab := f.rawFrameSize

	buffer := make([]byte, amountToGrab, amountToGrab)
	amount, err := io.ReadFull(f.frameSource, buffer)

	if err != nil || amount != amountToGrab {
		return frame, err
	}

	frame.pixels = buffer
	frame.Width = f.FrameWidth
	frame.Height = f.FrameHeight
	frame.Index = 0
	return frame, nil
}

//finalize Clears the frame buffer and closes the decoding stream
func (f *Decoder) finalize() {
	close(f.frameBuffer)
	f.opened = false
	f.frameSource.Close()
	f.decoderOpened = false
}

//createFFMPegArguments creates command line magic with the given options for the video fileName
func createFFMPegArguments(fileName string, fps30, conv480p bool) []string {
	args := strings.Split(_FFMPEGArgs, " ")

	args[1] = fileName

	magic := make([]string, 0)
	if fps30 {
		magic = append(magic, strings.Split(_FFMPEGArgs30fps, " ")...)
	}

	if conv480p {
		magic = append(magic, strings.Split(_FFMPEGArgs480p, " ")...)
	}

	fullargs := make([]string, 0)
	fullargs = append(fullargs, args[:2]...)
	fullargs = append(fullargs, magic...)
	fullargs = append(fullargs, args[2:]...)

	return fullargs
}

//probeFileInformation retrieves the hieght, width, and fps from a video file using ffprobe
func probeFileInformation(fileLocation string) (int, int, error) {
	var height, fps int
	args := strings.Split(_FFProbeArgs, " ")
	args[0] = fileLocation
	probe := exec.Command(_FFProbeCommnand, args...)

	reader, err := probe.StdoutPipe()

	if err != nil {
		return height, fps, err
	}

	err = probe.Start()
	if err != nil {
		return height, fps, err
	}

	jsonDecoder := json.NewDecoder(reader)

	type Streams struct {
		Height     int    `json:"height"`
		RFrameRate string `json:"r_frame_rate"`
	}

	type Info struct {
		Streams []Streams `json:"streams"`
	}

	var info Info
	jsonDecoder.Decode(&info)

	stream0 := info.Streams[0]
	height = stream0.Height
	fps = int(calculateFpsFromRatio(stream0.RFrameRate))
	return height, fps, nil
}

//probeStreamInfo retrieves the video hieght, width, and fps from an ffmpeg stderr stream
func probeStreamInfo(stderr io.ReadCloser) (int, int, int, error) {
	height, width, fps := -1, -1, -1

	reader := bufio.NewReader(stderr)

	//Read until we have all variables
	for fps == -1 || width == -1 || height == -1 {
		str, err := reader.ReadString('\n')

		if err != nil {
			return height, width, fps, err
		}

		//find resolution
		if resolutionRegex.MatchString(str) {
			matchGroups := resolutionRegex.FindStringSubmatch(str)
			if resx, err := strconv.Atoi(matchGroups[len(matchGroups)-2]); err != nil {
				return height, width, fps, err
			} else {
				width = resx
			}
			if resy, err := strconv.Atoi(matchGroups[len(matchGroups)-1]); err != nil {
				return height, width, fps, err
			} else {
				height = resy
			}
		}

		//Find frames per second
		if fpsRegex.MatchString(str) {
			if parsedFps, err := strconv.Atoi(fpsRegex.FindStringSubmatch(str)[1]); err != nil {
				return height, width, fps, err
			} else {
				fps = parsedFps
			}
		}
	}

	return height, width, fps, nil
}

//calculateFpsFromRatio takes a ratio string from FFMpeg (ex: Frames/Seconds) and converts it to fps
func calculateFpsFromRatio(ratio string) float64 {
	operands := strings.Split(ratio, "/")
	numerator, _ := strconv.Atoi(operands[0])
	denominator, _ := strconv.Atoi(operands[1])

	if denominator == 0 {
		denominator = 1 //Can't divide by 0, it is possible for a ratio to have a 0 denominator
	}

	return float64(numerator) / float64(denominator)
}
