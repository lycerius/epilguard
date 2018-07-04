package decoder

//Frame 2D Image Frame with colors in z axis
type Frame struct {
	raw           []byte //Frame container
	Height, Width int    //Height and Width for the current frame
	Index         uint   //The frame number
}

//Pixel Reperesents colored element a within a Frame
type Pixel struct {
	Red, Green, Blue int
}

func (f *Frame) GetRGB(x, y int) Pixel {
	//Every pixel is reperesented by 3 bytes, each in the RGB spectrum
	position := y*f.Width + x*3
	return Pixel{int(f.raw[position]), int(f.raw[position+1]), int(f.raw[position+2])}
}
