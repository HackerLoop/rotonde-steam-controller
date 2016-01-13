package main

import (
	"time"

	"github.com/GeertJohan/go.hid"
	"github.com/HackerLoop/rotonde-client.go"
	"github.com/HackerLoop/rotonde/shared"
	log "github.com/Sirupsen/logrus"
)

const vendorId = 0x28DE
const productId = 0x1142

func bitNames(b byte, bits []string, data map[string]interface{}) map[string]interface{} {
	for i, bit := range bits {
		data[bit] = (b>>uint(i))&0x1 == 0x1
	}
	return data
}

func main() {
	client := client.NewClient("ws://127.0.0.1:4224")

	controller := rotonde.Definition{"STEAM_CONTROLLER", "event", false, []*rotonde.FieldDefinition{}}
	client.AddLocalDefinition(&controller)

	go func() {
		var packet = make([]byte, 24)
		var data = make(map[string]interface{})
		var cc *hid.Device
		for {
			for {
				devices, err := hid.Enumerate(vendorId, productId)
				if err != nil {
					log.Warning(err)
					time.Sleep(time.Second * 2)
					continue
				}
				if len(devices) == 0 {
					log.Warning("No steam controller found")
					time.Sleep(time.Second * 2)
					continue
				}
				cc, err = hid.Open(vendorId, productId, "")
				if err != nil {
					log.Warning(err)
					time.Sleep(time.Second * 2)
					continue
				}
				break
			}

			log.Info("Start listening controller")

			for {
				if n, err := cc.Read(packet); err != nil {
					log.Warning(err)
					break
				} else if n == 0 {
					log.Warning("HID connection closed")
					break
				}

				// this is taken from https://github.com/virgilvox/node-steam-controller/blob/master/index.js
				// but it is actually wrong, bit masks should be used

				data = bitNames(packet[8], []string{"TRIGGER_RIGHT", "TRIGGER_LEFT", "TOP_RIGHT", "TOP_LEFT", "Y", "B", "X", "A"}, data)
				data = bitNames(packet[9], []string{"PAD_UP", "PAD_RIGHT", "PAD_LEFT", "PAD_DOWN", "CENTER_LEFT", "CENTER_STEAM", "CENTER_RIGHT", "BACK_LEFT"}, data)
				data = bitNames(packet[10], []string{"BACK_RIGHT", "PAD_PRESSED", "MOUSE_PRESSED", "PAD_TOUCHED", "MOUSE_TOUCHED"}, data)
				data["TRIGGER_LEFT_VALUE"] = uint8(packet[11])
				data["TRIGGER_RIGHT_VALUE"] = uint8(packet[12])
				data["PAD_A"] = int8(packet[16])
				data["PAD_B"] = int8(packet[17])
				data["PAD_C"] = int8(packet[18])
				data["PAD_D"] = int8(packet[19])

				data["MOUSE_A"] = int8(packet[20])
				data["MOUSE_B"] = int8(packet[21])
				data["MOUSE_C"] = int8(packet[22])
				data["MOUSE_D"] = int8(packet[23])
				client.SendEvent("STEAM_CONTROLLER", data)

			}
			cc.Close()
		}
	}()
	select {}
}
