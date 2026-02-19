package reader

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/raulaguila/go-blemultimeter/internal/domain"
	"github.com/raulaguila/go-blemultimeter/pkg/bluetooth"
)

const (
	maxConnectAttempts = 3
	connectTimeout     = 15 * time.Second
	baseRetryBackoff   = 1 * time.Second
)

type Reader struct {
	bt         bluetooth.Bluetooth
	multimeter domain.Multimeter

	serviceUUID              [16]byte
	characteristicNotifyUUID [16]byte

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewReader(multimeter domain.Multimeter) *Reader {
	ctx, cancel := context.WithCancel(context.Background())

	return &Reader{
		bt:         bluetooth.Bluetooth{},
		multimeter: multimeter,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (r *Reader) ConnectBT(deviceName string) error {
	var lastErr error

	for attempt := 1; attempt <= maxConnectAttempts; attempt++ {
		if r.ctx.Err() != nil {
			return r.ctx.Err()
		}

		ctx, cancel := context.WithTimeout(r.ctx, connectTimeout)
		err := r.bt.Connect(ctx, deviceName)
		cancel()
		if err == nil {
			return nil
		}

		lastErr = err
		log.Printf("BLE connect attempt %d/%d failed: %v", attempt, maxConnectAttempts, err)
		if attempt < maxConnectAttempts {
			select {
			case <-r.ctx.Done():
				return r.ctx.Err()
			case <-time.After(time.Duration(attempt) * baseRetryBackoff):
			}
		}
	}

	if lastErr == nil {
		return fmt.Errorf("connect to %q failed: unknown error", deviceName)
	}

	return fmt.Errorf("connect to %q failed after %d attempts: %w", deviceName, maxConnectAttempts, lastErr)
}

func (r *Reader) ConfigBTNotifier(serviceUUID [16]byte, characteristicNotifyUUID [16]byte) {
	r.serviceUUID = serviceUUID
	r.characteristicNotifyUUID = characteristicNotifyUUID
}

func (r *Reader) startBTNotifier(chNotify chan []byte) error {
	return r.bt.StartNotifier(r.ctx, chNotify, r.serviceUUID, r.characteristicNotifyUUID)
}

func (r *Reader) StartBTWriter(ch chan []byte, serviceUUID [16]byte, characteristicUUID [16]byte) error {
	return r.bt.StartWriter(r.ctx, ch, serviceUUID, characteristicUUID)
}

func (r *Reader) Connected() bool {
	return r.bt.Connected()
}

func (r *Reader) Disconnect() error {
	if r.cancel != nil {
		r.cancel()
	}

	r.wg.Wait()
	err := r.bt.Disconnect()

	r.ctx, r.cancel = context.WithCancel(context.Background())
	return err
}

func (r *Reader) RunNotifier(debug, extern bool, externChannel chan [3]any) error {
	if extern && externChannel == nil {
		return errors.New("extern channel is required when extern=true")
	}

	chNotify := make(chan []byte, 32)
	if err := r.startBTNotifier(chNotify); err != nil {
		return err
	}

	r.wg.Add(1)
	go r.loopNotify(chNotify, debug, extern, externChannel)
	return nil
}

func (r *Reader) loopNotify(chNotify chan []byte, debug, extern bool, externChannel chan [3]any) {
	defer r.wg.Done()

	for {
		select {
		case <-r.ctx.Done():
			return
		case payload := <-chNotify:
			if !r.bt.Connected() {
				continue
			}

			val, unit, flags := r.multimeter.ProcessArray(payload)
			if unit == "" {
				continue
			}

			if debug {
				log.Printf("| Debugging: %v %v [%v]", val, unit, strings.Join(flags, ", "))
			}

			if extern {
				select {
				case externChannel <- [3]any{val, unit, flags}:
				case <-r.ctx.Done():
					return
				}
			}
		}
	}
}
