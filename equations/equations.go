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
	y := int(0.2126*float64(R) + 0.7152*float64(G) + 0.0722*float64(B))
	if _, ok := rgbBrightnessLookup[y]; !ok {
		rgbBrightnessLookup[y] = int(413.435 * math.Pow((0.002745*float64(y)+0.0189623), 2.2))
	}
	return rgbBrightnessLookup[y]
}
