package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"time"

	"github.com/raulaguila/go-blemultimeter/internal/pkg/reader"
	"github.com/raulaguila/go-blemultimeter/pkg/alert"
	"github.com/raulaguila/go-blemultimeter/pkg/multimeter/fs9721"
	"github.com/raulaguila/go-blemultimeter/pkg/multimeter/owon"
)

var myReader *reader.Reader

func fs9721lp3() {
	myReader = reader.NewReader(&fs9721.Fs9721{})
	myReader.ConnectBT(fs9721.DeviceName)
	myReader.ConfigBTNotifier(fs9721.ServiceUUID, fs9721.CharacteristicNotifyUUID)

	// Receiving results and printing externally.
	ch := make(chan [3]interface{})
	myReader.RunNotifier(false, true, ch)
	go func() {
		for myReader.Connected() {
			result := <-ch
			log.Printf("| Received on main: %v %v [%v]\n", result[0].(float64), result[1].(string), strings.Join(result[2].([]string), ", "))
		}
	}()
}

func ow18e() {
	myReader = reader.NewReader(&owon.OW18E{})
	myReader.ConnectBT(owon.DeviceName)
	myReader.ConfigBTNotifier(owon.ServiceUUID, owon.CharacteristicNotifyUUID)

	// Receiving results and printing internally.
	myReader.RunNotifier(true, false, nil)
}

func main() {
	if len(os.Args) == 1 {
		log.Println("Required argument: \"fs9721\" or \"ow18e\"")
		return
	}

	device := strings.TrimSpace(strings.ToLower(os.Args[1]))
	switch device {
	case "fs9721":
		fs9721lp3()
	case "ow18e":
		ow18e()
	default:
		log.Println("Invalid argument!")
		log.Println("Valid arguments: \"fs9721\" or \"ow18e\"")
	}

	if myReader.Connected() {
		log.Println("Press <ENTER> to exit")
		bufio.NewScanner(os.Stdin).Scan()
		alert.Error(myReader.Disconnect())
		time.Sleep(2 * time.Second)
	}
}
