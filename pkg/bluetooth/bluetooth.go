package bluetooth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/raulaguila/go-blemultimeter/pkg/bluetooth/enum"
	"tinygo.org/x/bluetooth"
)

type Bluetooth struct {
	chScan          chan bluetooth.ScanResult
	connected       atomic.Bool
	adapter         *bluetooth.Adapter
	device          *bluetooth.Device
	characteristics [2]bluetooth.DeviceCharacteristic
	mu              sync.RWMutex
}

func (b *Bluetooth) enable() error {
	b.mu.RLock()
	if b.adapter != nil {
		b.mu.RUnlock()
		return nil
	}
	b.mu.RUnlock()

	adapter := bluetooth.DefaultAdapter
	if err := adapter.Enable(); err != nil {
		return err
	}

	b.mu.Lock()
	if b.adapter == nil {
		b.adapter = adapter
	}
	b.mu.Unlock()

	return nil
}

func (b *Bluetooth) find(deviceName string) error {
	b.mu.RLock()
	adapter := b.adapter
	b.mu.RUnlock()
	if adapter == nil {
		return errors.New("bluetooth adapter not initialized")
	}

	chScan := make(chan bluetooth.ScanResult, 1)
	b.mu.Lock()
	b.chScan = chScan
	b.mu.Unlock()

	go func() {
		err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
			if device.LocalName() != deviceName {
				return
			}

			_ = adapter.StopScan()
			select {
			case chScan <- device:
			default:
			}
		})
		if err != nil {
			log.Printf("BLE scan error: %v", err)
		}
	}()

	return nil
}

func (b *Bluetooth) Connected() bool {
	return b.connected.Load()
}

func (b *Bluetooth) Disconnect() error {
	if !b.connected.Swap(false) {
		return nil
	}

	b.mu.RLock()
	adapter := b.adapter
	device := b.device
	b.mu.RUnlock()

	if adapter != nil {
		_ = adapter.StopScan()
	}

	if device != nil {
		return device.Disconnect()
	}

	return nil
}

func (b *Bluetooth) Connect(ctx context.Context, deviceName string) error {
	if b.Connected() {
		return nil
	}

	if err := b.enable(); err != nil {
		return err
	}

	if err := b.find(deviceName); err != nil {
		return err
	}

	b.mu.RLock()
	chScan := b.chScan
	adapter := b.adapter
	b.mu.RUnlock()
	if chScan == nil || adapter == nil {
		return errors.New("scanner not initialized")
	}

	select {
	case <-ctx.Done():
		_ = adapter.StopScan()
		return ctx.Err()
	case device := <-chScan:
		dev, err := adapter.Connect(device.Address, bluetooth.ConnectionParams{})
		if err != nil {
			return err
		}

		b.mu.Lock()
		b.device = dev
		b.mu.Unlock()
		b.connected.Store(true)
		return nil
	}
}

func (b *Bluetooth) ScanDevices() {
	b.mu.RLock()
	adapter := b.adapter
	b.mu.RUnlock()
	if adapter == nil {
		log.Println("adapter is not enabled")
		return
	}

	go func() {
		err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			log.Println(result, result.LocalName())
		})
		if err != nil {
			log.Printf("BLE scan error: %v", err)
		}
	}()
}

func (b *Bluetooth) ListUUIDs() error {
	b.mu.RLock()
	device := b.device
	b.mu.RUnlock()
	if device == nil {
		return errors.New("bluetooth device not connected")
	}

	services, err := device.DiscoverServices(nil)
	if err != nil {
		return err
	}

	for _, service := range services {
		fmt.Printf("Service UUID: %v\n", service.UUID())
		characteristics, err := service.DiscoverCharacteristics(nil)
		if err != nil {
			return err
		}

		for _, characteristic := range characteristics {
			fmt.Printf(" - Characteristic UUID: %v\n", characteristic.UUID())
		}

		fmt.Println("")
	}

	return nil
}

func (b *Bluetooth) getCharacteristic(serviceUUID [16]byte, characteristicUUID [16]byte) (*bluetooth.DeviceCharacteristic, error) {
	if !b.Connected() {
		return nil, errors.New("bluetooth not connected")
	}

	b.mu.RLock()
	device := b.device
	b.mu.RUnlock()
	if device == nil {
		return nil, errors.New("bluetooth device not initialized")
	}

	services, err := device.DiscoverServices([]bluetooth.UUID{bluetooth.NewUUID(serviceUUID)})
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, errors.New("could not find service")
	}

	characteristics, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{bluetooth.NewUUID(characteristicUUID)})
	if err != nil {
		return nil, err
	}

	if len(characteristics) == 0 {
		return nil, errors.New("could not find characteristic")
	}

	return &characteristics[0], nil
}

func (b *Bluetooth) StartNotifier(ctx context.Context, chNotify chan<- []byte, serviceUUID [16]byte, characteristicUUID [16]byte) error {
	characteristic, err := b.getCharacteristic(serviceUUID, characteristicUUID)
	if err != nil {
		return err
	}

	b.characteristics[enum.Reader] = *characteristic
	if err := b.characteristics[enum.Reader].EnableNotifications(func(byteArray []byte) {
		payload := append([]byte(nil), byteArray...)
		select {
		case <-ctx.Done():
		case chNotify <- payload:
		default:
		}
	}); err != nil {
		return err
	}

	return nil
}

func (b *Bluetooth) StartWriter(ctx context.Context, ch <-chan []byte, serviceUUID [16]byte, characteristicUUID [16]byte) error {
	characteristic, err := b.getCharacteristic(serviceUUID, characteristicUUID)
	if err != nil {
		return err
	}

	b.characteristics[enum.Writer] = *characteristic
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case payload, ok := <-ch:
				if !ok {
					return
				}
				if !b.Connected() {
					return
				}
				if _, err := b.characteristics[enum.Writer].WriteWithoutResponse(payload); err != nil {
					log.Printf("BLE write error: %v", err)
				}
			}
		}
	}()

	return nil
}
