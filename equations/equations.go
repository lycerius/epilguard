package equations

const PercentageFlashArea float32 = 0.25
const FlashFrequencyMax float32 = 3
const FlashDeltaMax float32 = 20

var rgbBrightnessLookup = map[int]int{}

//Optimization, lookup table
func RGBtoLuminance(R, G, B float32) float32 {
	return 0.299*R + 0.587*G + 0.114*B
}

//Optimization, lookup table
func LuminanceToBrightness(Y float32) int {
	return int(413.435 * (0.002745*Y + 0.0189623))
}

//RGBtoBrightness coverts RGB values to brightness values
func RGBtoBrightness(R, G, B int) int {
	value := int(413.435 * (0.002745*(0.299*float64(R)+0.587*float64(G)+0.114*float64(B)) + 0.0189623))
	return value
}
