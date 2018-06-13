package tools

const PercentageFlashArea float32 = 0.25
const FlashFrequencyMax float32 = 3
const FlashDeltaMax float32 = 20

//Optimization, lookup table
func RGBtoLuminance(R, G, B float32) float32 {
	return 0.299*R + 0.587*G + 0.114*B
}

//Optimization, lookup table
func LuminanceToBrightness(Y float32) int {
	return int(413.435 * (0.002745*Y + 0.0189623))
}

func RGBtoBrightness(R, G, B float32) int {
	return LuminanceToBrightness(RGBtoLuminance(R, G, B))
}
