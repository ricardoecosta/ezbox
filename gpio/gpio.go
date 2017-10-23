package gpio

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"log"

	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

const PinRootDirectory = "/sys/class/gpio"
const PinExportFileName = PinRootDirectory + "/export"

const (
	In  uint8 = 0
	Out uint8 = 1
)

var (
	stream        chan Pin
	simulatedGPIO gpioMap
	watchedPins   = make(map[string]Pin)
)

type gpioMap struct {
	pins  map[uint8]Pin
	mutex sync.Mutex
}

type Pin struct {
	Number        uint8 `json:"number"`
	Value         uint8 `json:"value"`
	valueFileName string
}

func SetSimulatedGpio(pin Pin) {
	simulatedGPIO.mutex.Lock()
	defer simulatedGPIO.mutex.Unlock()

	simulatedGPIO.pins[pin.Number] = pin
	stream <- pin
}

func SubscribeToGpioStream(simulatedGPIOEnabled bool) <-chan Pin {
	stream = make(chan Pin, 1)
	if simulatedGPIOEnabled {
		simulatedGPIO = gpioMap{pins: make(map[uint8]Pin)}
		// todo: setup SimulatePinUpdate message handler
	} else {
		// todo: read from config file
		go watchPins(2)
	}
	return stream
}

func watchPins(pinNumbers ...uint8) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.WithMessage(err, "unable to create fs watcher")
	}
	defer watcher.Close()

	for _, n := range pinNumbers {
		pin := Pin{Number: n, valueFileName: pinValueFileName(n)}
		watchedPins[pin.valueFileName] = pin
		pin.watch(watcher)
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					fileName := event.Name
					if strings.HasSuffix(fileName, "/value") {
						pin := watchedPins[fileName]

						file, err := openFileInReadMode(fileName)
						if err != nil {
							log.Printf("unable to open pin value file in read mode, pin=%v err=%v", pin.Number, err)
						}

						buf := make([]byte, 1)
						if _, err := file.Read(buf); err != nil {
							log.Printf("unable to read pin from value file, pin=%v err=%v", pin.Number, err)
						}

						strVal := string(buf[:])
						val, err := strconv.Atoi(strVal)
						if err != nil {
							log.Printf("unable to parse pin value to integer, pin=%v value=%v", pin.Number, strVal)
						}

						stream <- Pin{Number: pin.Number, Value: uint8(val)}
					}
				}
			case err := <-watcher.Errors:
				// todo: this can produce lots of errors
				log.Println("error:", err)
			}
		}
	}()

	<-make(chan bool)
	return nil
}

func (p Pin) watch(fileWatcher *fsnotify.Watcher) error {
	err := export(p)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("error exporting pin, pin=%d", p.Number))
	}

	err = setToReadMode(p)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("error setting pin to read mode, pin=%d", p.Number))
	}

	pinValueFileName := pinValueFileName(p.Number)
	_, err = canOpenFileInReadMode(pinValueFileName)
	if err != nil {
		message := fmt.Sprintf("unable to open file in reading mode, file=%v", pinValueFileName)
		return errors.WithMessage(err, message)
	}

	err = setEdgeToBoth(p)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("error setting pin edge to both, pin=%d", p.Number))
	}

	err = setActiveLowToOne(p)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("error setting active low to one, pin=%d", p.Number))
	}

	pinDirectory := pinDirectory(p.Number)
	err = fileWatcher.Add(pinDirectory)
	if err != nil {
		message := fmt.Sprintf("unable to watch directory for changes, dir=%s", pinDirectory)
		return errors.WithMessage(err, message)
	}

	log.Printf("watching pin directory, number=%d dir=%s", p.Number, pinDirectory)

	return nil
}

// todo: unexport?
func export(pin Pin) error {
	return writePinNumberToFile(PinExportFileName, pin)
}

func setEdgeToBoth(pin Pin) error {
	return writeStringToFile(pinEdgeFileName(pin.Number), "both")
}

func setToReadMode(pin Pin) error {
	return writeStringToFile(pinDirectionFileName(pin.Number), "in")
}

func setActiveLowToOne(pin Pin) error {
	return writeStringToFile(pinActiveLowFileName(pin.Number), "1")
}

func writePinNumberToFile(fileName string, pin Pin) error {
	return writeStringToFile(fileName, strconv.Itoa(int(pin.Number)))
}

func writeStringToFile(fileName string, str string) error {
	file, err := openFileInWriteMode(fileName)
	if err != nil {
		message := fmt.Sprintf("unable to write string to file, file=%v string=%v", fileName, str)
		errors.WithMessage(err, message)
	}
	defer file.Close()

	file.Write([]byte(str))
	file.Sync()

	return nil
}

func canOpenFileInReadMode(fileName string) (bool, error) {
	return canOpenFileWithFlagAndPerm(fileName, os.O_RDONLY, 0400)
}

func canOpenFileWithFlagAndPerm(fileName string, flag int, perm os.FileMode) (bool, error) {
	file, err := os.OpenFile(fileName, flag, perm)
	if err != nil {
		return false, err
	}
	defer file.Close()
	return true, nil
}

func openFileInReadMode(fileName string) (*os.File, error) {
	return os.OpenFile(fileName, os.O_RDONLY, 0400)
}

func openFileInWriteMode(fileName string) (*os.File, error) {
	return os.OpenFile(fileName, os.O_WRONLY, 0600)
}

func pinDirectory(pinNumber uint8) string {
	return fmt.Sprintf("%v/gpio%v", PinRootDirectory, pinNumber)
}

func pinValueFileName(pinNumber uint8) string {
	return fmt.Sprintf("%v/value", pinDirectory(pinNumber))
}

func pinDirectionFileName(pinNumber uint8) string {
	return fmt.Sprintf("%v/direction", pinDirectory(pinNumber))
}

func pinActiveLowFileName(pinNumber uint8) string {
	return fmt.Sprintf("%v/active_low", pinDirectory(pinNumber))
}

func pinEdgeFileName(pinNumber uint8) string {
	return fmt.Sprintf("%v/gpio%v/edge", PinRootDirectory, pinNumber)
}
