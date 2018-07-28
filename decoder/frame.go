package decoder

//Frame 2D Image Frame
type Frame struct {
	pixels        []byte //Frame container
	Height, Width int    //Height and Width for the current frame
	Index         uint   //The frame number
}

//Pixel Reperesents colored element a within a Frame
type Pixel struct {
	Red, Green, Blue int
}

//GetRGB Returns the Pixel at x,y in a frame
func (f *Frame) GetRGB(x, y int) Pixel {
	//Every pixel is reperesented by 3 bytes, each in the RGB spectrum
	position := y*f.Width + x*3
	return Pixel{int(f.pixels[position]), int(f.pixels[position+1]), int(f.pixels[position+2])}
}
