package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"time"

	"github.com/raulaguila/go-blemultimeter/internal/reader"
	"github.com/raulaguila/go-blemultimeter/pkg/multimeter/fs9721"
	"github.com/raulaguila/go-blemultimeter/pkg/multimeter/owon"
)

var appReader *reader.Reader

func runFS9721() error {
	appReader = reader.NewReader(&fs9721.FS9721{})
	if err := appReader.ConnectBT(fs9721.DeviceName); err != nil {
		return err
	}

	appReader.ConfigBTNotifier(fs9721.ServiceUUID, fs9721.CharacteristicNotifyUUID)

	// Receiving results and printing externally.
	ch := make(chan [3]any, 32)
	if err := appReader.RunNotifier(false, true, ch); err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case result := <-ch:
				log.Printf("| Received on main: %v %v [%v]", result[0].(float64), result[1].(string), strings.Join(result[2].([]string), ", "))
			case <-ticker.C:
				if !appReader.Connected() {
					return
				}
			}
		}
	}()

	return nil
}

func runOW18E() error {
	appReader = reader.NewReader(&owon.OW18E{})
	if err := appReader.ConnectBT(owon.DeviceName); err != nil {
		return err
	}

	appReader.ConfigBTNotifier(owon.ServiceUUID, owon.CharacteristicNotifyUUID)

	// Receiving results and printing internally.
	return appReader.RunNotifier(true, false, nil)
}

func main() {
	if len(os.Args) == 1 {
		log.Println("Required argument: \"fs9721\" or \"ow18e\"")
		return
	}

	device := strings.TrimSpace(strings.ToLower(os.Args[1]))
	var err error
	if device == "fs9721" {
		err = runFS9721()
	} else if device == "ow18e" {
		err = runOW18E()
	} else {
		log.Println("Invalid argument!")
		log.Println("Valid arguments: \"fs9721\" or \"ow18e\"")
		return
	}

	if err != nil {
		log.Printf("Failed to start reader: %v", err)
		return
	}

	if appReader != nil && appReader.Connected() {
		log.Println("Press <ENTER> to exit")
		bufio.NewScanner(os.Stdin).Scan()
		if err := appReader.Disconnect(); err != nil {
			log.Printf("Failed to disconnect bluetooth: %v", err)
		}
		time.Sleep(2 * time.Second)
	}
}
