package reader

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/raulaguila/go-blemultimeter/internal/domain"
	"github.com/raulaguila/go-blemultimeter/pkg/alert"
	"github.com/raulaguila/go-blemultimeter/pkg/bluetooth"
)

type Reader struct {
	bt         bluetooth.Bluetooth
	multimeter domain.Multimeter

	serviceUUID              [16]byte
	characteristicNotifyUUID [16]byte
}

func NewReader(multimeter domain.Multimeter) *Reader {
	return &Reader{
		bt:         bluetooth.Bluetooth{},
		multimeter: multimeter,
	}
}

func (r *Reader) ConnectBT(deviceName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	alert.Error(r.bt.Connect(ctx, deviceName))
}

func (r *Reader) ConfigBTNotifier(serviceUUID [16]byte, characteristicNotifyUUID [16]byte) {
	r.serviceUUID = serviceUUID
	r.characteristicNotifyUUID = characteristicNotifyUUID
}

func (r *Reader) startBTNotifier(chNotify chan []byte) {
	alert.Error(r.bt.StartNotifier(chNotify, r.serviceUUID, r.characteristicNotifyUUID))
}

func (r *Reader) StartBTWriter(ch chan []byte, ServiceUUID [16]byte, CharacteristicUUID [16]byte) {
	alert.Error(r.bt.StartWriter(ch, ServiceUUID, CharacteristicUUID))
}

func (r *Reader) Connected() bool {
	return r.bt.Connected()
}

func (r *Reader) Disconnect() error {
	return r.bt.Disconnect()
}

func (r *Reader) RunNotifier(debug, extern bool, externChanel chan [3]interface{}) {
	chNotify := make(chan []byte)
	r.startBTNotifier(chNotify)

	go r.loopNotify(chNotify, debug, extern, externChanel)
}

func (r *Reader) loopNotify(chNotify chan []byte, debug, extern bool, externChanel chan [3]interface{}) {
	for r.bt.Connected() {
		val, unit, flags := r.multimeter.ProccessArray(<-chNotify)
		if unit != "" {
			if debug {
				log.Printf("| Debugging: %v %v [%v]\n", val, unit, strings.Join(flags, ", "))
			}

			if extern {
				externChanel <- [3]interface{}{val, unit, flags}
			}
		}
	}
}
