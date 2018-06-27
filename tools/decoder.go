package tools

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

//Magic command for executing ffmpeg
const _FFProbeCommnand = "ffprobe"
const _FFProbeArgs = "[filename] -v quiet -print_format json -show_format -show_streams"
const _FFMPEGCommand string = "ffmpeg"
const _FFMPEGArgs string = "-i [filename] -an -pix_fmt rgb24 -c:v rawvideo -map 0:v -f image2pipe -"
const _FFMPEGArgs_480p = "-s hd480"
const _FFMPEGArgs_30fps = "-r 30 -framerate 30"
const _FrameBufferDefaultSize = 30

//Resolution Finding Regex
var resolutionRegex = regexp.MustCompile(`rgb24, (\d*)x(\d*)`)
var fpsRegex = regexp.MustCompile(`(\d*) fps`)

//FFMPEGDecoder A video stream with an fmpeg process as the server
type FFMPEGDecoder struct {
	FileName                string
	FrameWidth, FrameHeight int //Size of a frame in X,Y
	cmdString               string
	cmd                     *exec.Cmd     //FFMPEG process
	stdout                  io.ReadCloser //Stdout for ffmpeg
	Opened                  bool          //Decoding in process
	CloseDecoder            chan interface{}
	Fps                     int         //Frames per second
	FrameBuffer             chan *Frame //Frame buffer channel
	rawFrameSize            int         //FrameWidth * FrameHeight * 3
	FrameBufferSize         int
	ConvertTo30FPS          bool
	ConvertTo480p           bool
}

//Frame 2D Image Frame with colors in z axis
type Frame struct {
	raw           []byte //Frame container
	Height, Width int    //Height and Width for the current frame
	Index         uint
}

//Pixel Reperesents colored element a within a Frame
type Pixel struct {
	Red, Green, Blue int
}

type videoInfo struct {
	Height, Width int
	Fps           int
}

//NewDecoder Creates a new video decoder for the given file
func NewDecoder(fileName string) FFMPEGDecoder {
	var fvs FFMPEGDecoder
	fvs.FileName = fileName
	fvs.cmdString = _FFMPEGArgs
	fvs.FrameBufferSize = _FrameBufferDefaultSize
	return fvs
}

//Start begins decoding the video
func (f *FFMPEGDecoder) Start() error {

	//Already in process
	if f.Opened {
		return errors.New("Decoder has already been started")
	}

	//File does not exist
	if _, err := os.Stat(f.FileName); err != nil {
		return err
	}

	info, err := getVideoInformation(f.FileName)

	if err != nil {
		return err
	}

	arguments := createArguments(f.FileName, info.Fps > 30, info.Width > 480)

	f.cmd = exec.Command(_FFMPEGCommand, arguments...)
	//Link Stdout
	stdout, err := f.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	f.cmd.Start()

	f.ConvertTo30FPS = info.Fps > 30
	f.ConvertTo480p = info.Width > 480
	f.stdout = stdout
	f.Opened = true
	f.rawFrameSize = f.FrameHeight * f.FrameWidth * 3
	f.FrameBuffer = make(chan *Frame, f.FrameBufferSize)
	f.CloseDecoder = make(chan interface{}, 1)
	f.Fps = info.Fps
	f.FrameHeight = info.Height
	f.FrameWidth = info.Width
	//Concurrently fill the framebuffer
	go frameBufferFiller(f)
	return nil

}

func frameBufferFiller(f *FFMPEGDecoder) {
	frameBuffer := f.FrameBuffer
	var fIndex uint
	for f.Opened {
		if frame, err := f.NextFrame(); err == nil {
			frame.Index = fIndex
			fIndex++

			select {
			case <-f.CloseDecoder:
				f.Opened = false
			case frameBuffer <- frame:
			}

		} else {
			f.Opened = false
			break
		}

	}
	f.stdout.Close()
	close(f.FrameBuffer)
	f.FrameBuffer = nil
}

func createArguments(fileName string, fps30, conv480p bool) []string {
	args := strings.Split(_FFMPEGArgs, " ")

	args[1] = fileName

	magic := make([]string, 0)
	if fps30 {
		magic = append(magic, strings.Split(_FFMPEGArgs_30fps, " ")...)
	}

	if conv480p {
		magic = append(magic, strings.Split(_FFMPEGArgs_480p, " ")...)
	}

	fullargs := make([]string, 0)
	fullargs = append(fullargs, args[:2]...)
	fullargs = append(fullargs, magic...)
	fullargs = append(fullargs, args[2:]...)

	return fullargs
}

//Close Closes the video decoder
func (f *FFMPEGDecoder) Close() error {
	if f.Opened {
		f.CloseDecoder <- nil
	}
	return nil
}

//NextFrame Gets the next frame in a stream
func (f *FFMPEGDecoder) NextFrame() (*Frame, error) {
	amountToGrab := f.rawFrameSize

	buffer := make([]byte, amountToGrab, amountToGrab)
	amount, err := io.ReadFull(f.stdout, buffer)

	if err != nil || amount != amountToGrab {
		return nil, err
	}

	return &Frame{buffer, f.FrameHeight, f.FrameWidth, 0}, nil
}

func (f *FFMPEGDecoder) Next() *Frame {
	if !f.Opened {
		return nil
	}

	return <-f.FrameBuffer
}

//GetRGB Get RGB as a point within a frame
func (f *Frame) GetRGB(x, y int) Pixel {
	//Every pixel is reperesented by 3 bytes, each in the RGB spectrum
	position := y*f.Width + x*3
	return Pixel{int(f.raw[position]), int(f.raw[position+1]), int(f.raw[position+2])}
}

func getVideoInformation(fileLocation string) (videoInfo, error) {
	args := strings.Split(_FFProbeArgs, " ")
	args[0] = fileLocation
	probe := exec.Command(_FFProbeCommnand, args...)

	reader, err := probe.StdoutPipe()

	if err != nil {
		return videoInfo{}, err
	}

	var info videoInfo
	probe.Start()

	jsonDecoder := json.NewDecoder(reader)

	type Streams struct {
		Width      int    `json:"width"`
		Height     int    `json:"height"`
		RFrameRate string `json:"r_frame_rate"`
	}

	type Info struct {
		Streams []Streams `json:"streams"`
	}

	var vidInfo Info
	jsonDecoder.Decode(&vidInfo)

	info.Fps = int(getFrameRateFromRatio(vidInfo.Streams[0].RFrameRate))
	info.Height = vidInfo.Streams[0].Height
	info.Width = vidInfo.Streams[0].Width

	return info, nil
}

func getFrameRateFromRatio(ratio string) float64 {
	operands := strings.Split(ratio, "/")
	numerator, _ := strconv.Atoi(operands[0])
	denominator, _ := strconv.Atoi(operands[1])

	if denominator == 0 {
		denominator = 1 //Can't divide by 0
	}

	return float64(numerator) / float64(denominator)
}
