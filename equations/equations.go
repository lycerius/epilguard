package equations

import (
	"math"
)

const PercentageFlashArea float32 = 0.25
const FlashFrequencyMax float32 = 3
const FlashDeltaMax float32 = 20

var rgbBrightnessLookup = map[int]int{}

//RGBtoBrightness coverts RGB values to brightness values
func RGBtoBrightness(R, G, B int) int {
	y := 0.299*float64(R) + 0.587*float64(G) + 0.114*float64(B)
	brightness := 413.435 * math.Pow(0.002745*y+0.0189623, 2.2)
	return int(brightness)
}
