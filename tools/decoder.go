package tools

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

//Magic command for executing ffmpeg
const _FFMPEGCommand string = "ffmpeg"
const _FFMPEGArgs string = "-i [filename] -framerate 30 -r 30 -s hd480 -an -pix_fmt rgb24 -c:v rawvideo -map 0:v -f image2pipe -"
const _FrameBufferDefaultSize = 30

//Resolution Finding Regex
var resolutionRegex = regexp.MustCompile(`rgb24, (\d*)x(\d*)`)
var fpsRegex = regexp.MustCompile(`(\d*) fps`)

//FFMPEGDecoder A video stream with an fmpeg process as the server
type FFMPEGDecoder struct {
	FileName                string
	FrameWidth, FrameHeight int           //Size of a frame in X,Y
	cmd                     *exec.Cmd     //FFMPEG process
	stdout                  io.ReadCloser //Stdout for ffmpeg
	Opened                  bool          //Decoding in process
	_OpenedChanel           chan bool
	Fps                     int         //Frames per second
	FrameBuffer             chan *Frame //Frame buffer channel
	rawFrameSize            int         //FrameWidth * FrameHeight * 3
	FrameBufferSize         int
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

//NewDecoder Creates a new video decoder for the given file
func NewDecoder(fileName string) FFMPEGDecoder {
	var fvs FFMPEGDecoder
	fvs.FileName = fileName
	args := strings.Split(_FFMPEGArgs, " ")
	args[1] = fileName
	fvs.cmd = exec.Command(_FFMPEGCommand, args...)
	fvs._OpenedChanel = make(chan bool)
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

	//Link Stdout
	stdout, err := f.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	//Used for getting new resolution
	stderr, err := f.cmd.StderrPipe()
	if err != nil {
		return err
	}
	defer stderr.Close()

	f.cmd.Start()

	//Getting the resolution and from stderr output
	errReader := bufio.NewReader(stderr)
	for {
		str, err := errReader.ReadString('\n')
		if err != nil {
			return err
		}
		//Find resolution
		if resolutionRegex.MatchString(str) {
			matchGroups := resolutionRegex.FindStringSubmatch(str)
			if resx, err := strconv.Atoi(matchGroups[len(matchGroups)-2]); err != nil {
				return err
			} else {
				f.FrameWidth = resx
			}
			if resy, err := strconv.Atoi(matchGroups[len(matchGroups)-1]); err != nil {
				return err
			} else {
				f.FrameHeight = resy
			}
		}
		//Find frames per second
		if fpsRegex.MatchString(str) {

			if fps, err := strconv.Atoi(fpsRegex.FindStringSubmatch(str)[1]); err != nil {
				return err
			} else {
				f.Fps = fps
			}
		}

		//Found all needed parameters?
		if f.Fps != 0 && f.FrameHeight != 0 && f.FrameWidth != 0 {
			break
		}
	}

	f.stdout = stdout
	f.Opened = true
	f.rawFrameSize = f.FrameHeight * f.FrameWidth * 3
	f.FrameBuffer = make(chan *Frame, f.FrameBufferSize)
	//Concurrently fill the framebuffer
	go frameBufferFiller(f)
	return nil

}

func frameBufferFiller(f *FFMPEGDecoder) {
	frameBuffer := f.FrameBuffer
	var fIndex uint
	for {
		select {
		case <-f._OpenedChanel:
			close(f.FrameBuffer)
			f.FrameBuffer = nil
			f.stdout.Close()
			break
		default:
			if frame, err := f.NextFrame(); err == nil {
				frame.Index = fIndex
				fIndex++
				frameBuffer <- frame
			}
		}
	}
}

//Close Closes the video decoder
func (f *FFMPEGDecoder) Close() error {
	if f.Opened {
		if err := f.stdout.Close(); err == nil {
			f._OpenedChanel <- false
			f.Opened = false
			return nil
		} else {
			return err
		}
		//Kill process?
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
