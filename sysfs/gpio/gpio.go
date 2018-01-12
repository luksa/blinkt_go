package gpio

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

const OUTPUT = 1

type gpioPin struct {
	valueFd     *os.File
	directionFd *os.File
}

var gpioPins map[string]gpioPin

func Setup() {
	gpioPins = make(map[string]gpioPin)
}

func Cleanup() {
	for k, v := range gpioPins {
		fmt.Println("Cleaning up " + k)
		v.directionFd.Close()
		v.valueFd.Close()

		unexport(k)
	}
}

func export(pin string) {
	path := "/sys/class/gpio/export"
	bytesToWrite := []byte(pin)
	writeErr := ioutil.WriteFile(path, bytesToWrite, 0644)
	if writeErr != nil {
		log.Panic(writeErr)
	}
}

func unexport(pin string) {
	path := "/sys/class/gpio/unexport"
	bytesToWrite := []byte(pin)
	writeErr := ioutil.WriteFile(path, bytesToWrite, 0644)
	if writeErr != nil {
		log.Panic(writeErr)
	}
}

func pinExported(pin string) bool {
	pinPath := fmt.Sprintf("/sys/class/gpio/gpio%s", pin)
	if file, err := os.Stat(pinPath); err == nil && len(file.Name()) > 0 {
		return true
	}
	return false
}

func PinMode(pin string, val int) {

	exported := pinExported(pin)
	if val == OUTPUT {
		if exported == false {
			export(pin)
		}
	} else {
		if exported == true {
			unexport(pin)
		}
	}

	_, exists := gpioPins[pin]
	if exists == false {
		pinPath := fmt.Sprintf("/sys/class/gpio/gpio%s", pin)
		valueFd, openErr := os.OpenFile(pinPath+"/value", os.O_WRONLY, 0640)
		if openErr != nil {
			log.Panic(openErr, pinPath)
		}
		directionFd, openErr := os.OpenFile(pinPath+"/direction", os.O_WRONLY, 0640)
		if openErr != nil {
			log.Panic(openErr, pinPath)
		}
		gpioPins[pin] = gpioPin{
			valueFd:     valueFd,
			directionFd: directionFd,
		}
		if val == OUTPUT {
			_, err := gpioPins[pin].directionFd.Write([]byte("out"))
			if err != nil {
				log.Panic(err, fmt.Sprintf("Pin: %s Mode: %s Value: %s ", pin, "direction", val))
			}

		}
	}
}

func DigitalWrite(pin string, val int) {
	DigitalWriteString(pin, strconv.Itoa(val))
}

func DigitalWriteString(pin string, val string) {
	_, err := gpioPins[pin].valueFd.Write([]byte(val))
	if err != nil {
		log.Panic(err, fmt.Sprintf("Pin: %s Mode: %s Value: %s ", pin, "value", val))
	}
}
