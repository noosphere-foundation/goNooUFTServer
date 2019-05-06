package uftserver

import (
	"encoding/json"
	"goNooUFTServer/tools"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/evio"
)

var readChannelUDP chan []byte
var readChannelTCP chan []byte
var response []int
var dataMas map[int]bool
var buffer []byte
var mutex sync.RWMutex

func handleUDPConnection(c evio.Conn, in []byte) (out []byte, action evio.Action) {
	//

	readChannelUDP <- in

	return
}

func handleTCPConnection(c evio.Conn, in []byte) (out []byte, action evio.Action) {
	//

	buffer = append(buffer, in...)
	if buffer[len(buffer)-1] != byte('\n') {
		//

		return
	}

	data := string(buffer)
	dec := json.NewDecoder(strings.NewReader(data))
	for {
		//

		var sr tools.CommandStructure
		if err := dec.Decode(&sr); err == io.EOF {
			//

			buffer = nil
			break

		} else if err != nil {
			//

			buffer = nil
			log.Fatal("ERROR: ", err)
			return
		}

		if sr.COMMAND == "ACK" {
			//

			js, _ := json.Marshal(&tools.CommandStructure{COMMAND: "ACK"})
			js = append(js, '\n')
			out = js
			log.Println("ACK SENDED")

		} else if sr.COMMAND == "SYN" {
			//

			js, _ := json.Marshal(&tools.CommandStructure{COMMAND: "SYN-ACK"})
			js = append(js, '\n')
			out = js

		} else if sr.COMMAND == "START" {
			//

			js, _ := json.Marshal(&tools.CommandStructure{COMMAND: "ACK"})
			js = append(js, '\n')
			out = js

		} else if sr.COMMAND == "EOF-SHUFFLE" {
			//

			var response []int
			mutex.RLock()
			for _, v := range sr.DATA {
				//

				if _, ok := dataMas[v]; !ok {
					//

					response = append(response, v)
				}
			}
			mutex.RUnlock()

			sr.COMMAND = "ACK"
			sr.DATA = response
			js, _ := json.Marshal(sr)
			js = append(js, '\n')
			out = js

		} else if sr.COMMAND == "EOF-INTERVAL" {
			//

			var response []int
			mutex.RLock()
			for k := sr.DATA[0]; k < sr.DATA[1]; k++ {
				//

				if _, ok := dataMas[k]; !ok {
					//

					response = append(response, k)
				}
			}
			mutex.RUnlock()

			sr.COMMAND = "ACK"
			sr.DATA = response
			js, _ := json.Marshal(sr)
			js = append(js, '\n')
			out = js
		}
	}

	return
}

func startUDPServer() (err error) {
	//

	var events evio.Events
	events.Data = handleUDPConnection
	return evio.Serve(events, "udp://0.0.0.0:4444")
}

func startTCPServer() (err error) {
	//

	var events evio.Events
	events.Data = handleTCPConnection

	return evio.Serve(events, "tcp://0.0.0.0:3333")
}

func readWorkerUDP() {
	//

	for {
		//

		select {
		case <-time.After(2 * time.Second):
		//

		case data := <-readChannelUDP:
			dec := json.NewDecoder(strings.NewReader(string(data)))
			for {
				//

				var ds tools.DataStructure
				if err := dec.Decode(&ds); err == io.EOF {
					//

					break

				} else if err != nil {
					//

					break
				}

				mutex.RLock()
				_, ok := dataMas[ds.ID]
				mutex.RUnlock()
				if ok {
					//

				} else {
					//

					mutex.Lock()
					dataMas[ds.ID] = true
					mutex.Unlock()
				}
			}
		}
	}
}

func Serve() {
	//

	readChannelUDP = make(chan []byte, 1024*1024)
	readChannelTCP = make(chan []byte, 1024*1024)
	dataMas = make(map[int]bool)

	go readWorkerUDP()
	go startUDPServer()
	startTCPServer()
}
