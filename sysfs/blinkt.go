package sysfs

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/alexellis/blinkt_go/sysfs/gpio"
)

const DAT = "23"
const CLK = "24"
const PIXEL_START = 224 // 0b11100000 (224)

// default raw brightness.  Not to be used user-side
const defaultBrightnessInt int = 15

//upper and lower bounds for user specified brightness
const minBrightness float64 = 0.0
const maxBrightness float64 = 1.0

func writeByte(val int) {
	for i := 0; i < 8; i++ {
		// 0b10000000 = 128
		gpio.DigitalWrite(DAT, val&128)
		gpio.DigitalWriteString(CLK, "1")
		val = val << 1
		gpio.DigitalWriteString(CLK, "0")
	}
}

func convertBrightnessToInt(brightness float64) int {

	if !inRangeFloat(minBrightness, brightness, maxBrightness) {
		log.Fatalf("Supplied brightness was %#v - value should be between: %#v and %#v", brightness, minBrightness, maxBrightness)
	}

	return int(brightness * 31.0)

}

func inRangeFloat(minVal float64, testVal float64, maxVal float64) bool {

	return (testVal >= minVal) && (testVal <= maxVal)
}

// SetClearOnExit turns all pixels off on Control + C / os.Interrupt signal.
func (bl *Blinkt) SetClearOnExit(clearOnExit bool) {

	if clearOnExit {

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		fmt.Println("Press Control + C to stop")

		go func() {
			for range signalChan {
				bl.Close()
				os.Exit(1)
			}
		}()
	}
}

// Delay maps to time.Sleep, for ms milliseconds
func Delay(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// Clear sets all the pixels to off, you still have to call Show.
func (bl *Blinkt) Clear() {
	bl.SetAll(0, 0, 0)
}

func (bl *Blinkt) Close() {
	bl.Clear()
	bl.Show()
	gpio.Cleanup()
}

// Show updates the LEDs with the values from SetPixel/Clear.
func (bl *Blinkt) Show() {
	pixelsChanged := false
	for i, pixel := range bl.pixels {
		if bl.previousPixels[i] != pixel {
			pixelsChanged = true
			break
		}
	}
	if !pixelsChanged {
		return
	}

	sof()
	pixels := bl.pixels
	for p, _ := range pixels {
		pixel := bl.pixels[p]
		writeByte(PIXEL_START | pixel.Brightness)
		writeByte(pixel.B)
		writeByte(pixel.G)
		writeByte(pixel.R)
	}
	eof()
	bl.previousPixels = pixels
}

func sof() {
	gpio.DigitalWriteString(DAT, "0")
	//gpio.DigitalWriteString(CLK, "1010101010101010101010101010101010101010101010101010101010101010")
	for i := 0; i < 32; i++ {
		gpio.DigitalWriteString(CLK, "1")
		gpio.DigitalWriteString(CLK, "0")
	}

	//for i := 0; i < 4; i++ {
	//	writeByte(0)
	//}
}

func eof() {
	gpio.DigitalWriteString(DAT, "0")
	//gpio.DigitalWriteString(CLK, "101010101010101010101010101010101010101010101010101010101010101010101010")
	for i := 0; i < 36; i++ {
		gpio.DigitalWriteString(CLK, "1")
		gpio.DigitalWriteString(CLK, "0")
	}

	//writeByte(255)
	// 0xff = 255
}

// SetAll sets all pixels to specified r, g, b colour. Show must be called to update the LEDs.
func (bl *Blinkt) SetAll(r int, g int, b int) *Blinkt {

	for p, _ := range bl.pixels {
		bl.SetPixel(p, r, g, b)
	}

	return bl
}

// SetPixel sets an individual pixel to specified r, g, b colour. Show must be called to update the LEDs.
func (bl *Blinkt) SetPixel(p int, r int, g int, b int) *Blinkt {
	bl.pixels[p].R = r
	bl.pixels[p].G = g
	bl.pixels[p].B = b
	return bl
}

// SetBrightness sets the brightness of all pixels. Brightness supplied should be between: 0.0 to 1.0
func (bl *Blinkt) SetBrightness(brightness float64) *Blinkt {

	brightnessInt := convertBrightnessToInt(brightness)

	for p, _ := range bl.pixels {
		bl.pixels[p].Brightness = brightnessInt
	}

	return bl
}

// SetPixelBrightness sets the brightness of pixel p. Brightness supplied should be between: 0.0 to 1.0
func (bl *Blinkt) SetPixelBrightness(p int, brightness float64) *Blinkt {

	brightnessInt := convertBrightnessToInt(brightness)
	bl.pixels[p].Brightness = brightnessInt
	return bl
}

func initPixels(brightness int) [8]Pixel {
	var pixels [8]Pixel
	for p, _ := range pixels {
		pixels[p] = Pixel{0, 0, 0, brightness}
	}
	return pixels
}

// Setup initializes GPIO via WiringPi base library.
func (bl *Blinkt) Setup() {
	gpio.Setup()
	gpio.PinMode(DAT, gpio.OUTPUT)
	gpio.PinMode(CLK, gpio.OUTPUT)
}

// NewBlinkt creates a Blinkt to interact with. You must call "Setup()" immediately afterwards.
func NewBlinkt(brightness ...float64) Blinkt {

	//brightness is optional so set the default
	brightnessInt := defaultBrightnessInt

	//over-ride the default if the user has supplied a brightness value
	if len(brightness) > 0 {
		brightnessInt = convertBrightnessToInt(brightness[0])
	}
	return Blinkt{
		pixels: initPixels(brightnessInt),
	}
}

// Blinkt use the NewBlinkt function to initialize the pixels property.
type Blinkt struct {
	pixels         [8]Pixel
	previousPixels [8]Pixel
}

type Pixel struct {
	R, G, B, Brightness int
}

func init() {

}
