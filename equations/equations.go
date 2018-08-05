package equations

import (
	"math"
)

//PercentageFlashArea ITU-R: minimum flash area = 25%
const PercentageFlashArea float32 = 0.25

//FlashFrequencyMax ITU-R: maximum flash frequency = 3Hz
const FlashFrequencyMax int = 3

//FlashDeltaMax ITU-R: Delta Candellas must be >= 20cd/m^2
const FlashDeltaMax float32 = 20

//rgbBrightnessLookup optimization table for calculating brightness from luminance
var rgbBrightnessLookup = map[int]int{}

//RGBtoBrightness coverts RGB values to brightness values
func RGBtoBrightness(R, G, B int) int {
	//First convert to luminance
	y := int(0.2126*float64(R) + 0.7152*float64(G) + 0.0722*float64(B))

	//Did we already calculate gamma before hand?
	if gamma, ok := rgbBrightnessLookup[y]; ok {
		return gamma
	} else { //Calculate and store
		gamma = int(413.435 * math.Pow((0.002745*float64(y)+0.0189623), 2.2))
		rgbBrightnessLookup[y] = gamma
		return gamma
	}
}
