package main

import (
	"encoding/json"
	"goNooUFTServer/netclient"
	"goNooUFTServer/tools"
	"goNooUFTServer/uftserver"
	"log"
	"sync"
	"time"
)

var readChannel chan []byte
var tcpChannel chan interface{}
var dataMas map[int]string
var dataids []int
var dataForResend []int
var mutex sync.RWMutex
var udptimeout time.Duration
var loststat = 0
var counter = 0
var statanallizer = 0

func readWorker() {
	//

	for {
		//

		select {
		case data := <-readChannel:
			//

			if len(data) == 0 {
				//

				continue
			}

			var cs tools.CommandStructure
			err := json.Unmarshal([]byte(data), &cs)
			if err != nil {
				//

				panic("ERROR JSON UNMARSHAL")
				break
			}

			if cs.COMMAND == "ACK" {
				//

				for _, v := range cs.DATA {
					//

					dataForResend = append(dataForResend, v)
					loststat++
				}

			} else if cs.COMMAND == "START" {
				//

				log.Println("START!")
			}
		}
	}
}

func startClientUdpWriter(c chan []byte) {
	//

	udpclient := netclient.New("192.168.0.8:4444", netclient.UDP, c)
	udpclient.Start()

	tcpclient := netclient.New("192.168.0.8:3333", netclient.TCP, c)
	tcpclient.Start()

	// js, _ := json.Marshal(&tools.CommandStructure{COMMAND: "ACK"})
	// tcpclient.Write(js)

	dataToSend := string(`{
    "TT": "ST",
    "SENDER": "03b35bcb28be70695f467239bba98c5d97a3b41e9880dc0c187d07c1ffad769ba6",
    "RECEIVER": "03fb0cae43230effe29b1d0655ae83decb7e9824f2a099bb7a82d0cd44dffea08b",
    "TTOKEN": "NZT",
    "CTOKEN": "12.0",
    "TST": "1536844742",
    "SIGNATURE": "02bd2d7a72933bcf84517f0ef6802343a7ff3b86947b6c3146fba02d0119030458b54de1773f8b17d5ee7a59cea11b40b2c730f4e9079f7bc4bb38e71c49c9bb00"
}`)
	bulkSize := 10

	for {
		//

		select {
		case <-tcpChannel:
			//

			if len(dataids) >= bulkSize {
				//

				response := []int{dataids[0], dataids[bulkSize-1]}
				dataids = dataids[bulkSize:]
				js, err := json.Marshal(&tools.CommandStructure{"EOF-INTERVAL", response})
				if err != nil {
					//

					panic("WRONG JSON MARSHAL")
					break
				}

				tcpclient.Write(&js)
			}

		case <-time.After(udptimeout):

			if len(dataForResend) > 0 {
				//

				for {
					//

					l := len(dataForResend)
					if l == 0 {
						//

						break
					}
					var i = 0
					for ; i < bulkSize && i < l; i++ {
						//

						dataindex := dataForResend[i]
						if v, ok := dataMas[dataindex]; ok {
							//

							js, err := json.Marshal(&tools.DataStructure{dataindex, v})
							if err != nil {
								//

								panic("WRONG JSON MARSHAL")
								break
							}

							udpclient.Write(&js)
						}

					}

					js, err := json.Marshal(&tools.CommandStructure{"EOF-SHUFFLE", dataForResend[0:i]})
					if err != nil {
						//

						panic("WRONG JSON MARSHAL")
						break
					}

					dataForResend = dataForResend[i:]
					tcpclient.Write(&js)
					continue
				}
			}

			statanallizer++
			for i := 0; i < bulkSize; i++ {
				//

				counter++
				id := counter
				mutex.Lock()
				dataMas[id] = dataToSend
				dataids = append(dataids, id)
				mutex.Unlock()

				js, err := json.Marshal(&tools.DataStructure{id, dataToSend})

				if err != nil {
					//

					panic("WRONG JSON MARSHAL")
					break
				}

				udpclient.Write(&js)
			}

			if len(dataids) >= bulkSize {
				//

				response := []int{dataids[0], dataids[bulkSize-1]}
				dataids = dataids[bulkSize:]
				js, err := json.Marshal(&tools.CommandStructure{"EOF-INTERVAL", response})
				if err != nil {
					//

					panic("WRONG JSON MARSHAL")
					break
				}

				tcpclient.Write(&js)
			}

			if statanallizer == 500 {
				//

				if loststat == 0 {
					//

					bulkSize += 10

				} else {
					//

					bulkSize = bulkSize - bulkSize/2
				}

				if bulkSize <= 0 {
					//

					bulkSize = 10

				} else if bulkSize > 400 {
					//

					bulkSize = 400
				}

				statanallizer = 0
				loststat = 0
			}
		}
	}
}

func main() {
	//

	readChannel = make(chan []byte, 1024)
	tcpChannel = make(chan interface{}, 1024)
	dataMas = make(map[int]string)
	udptimeout = 1 * time.Microsecond
	go readWorker()
	go startClientUdpWriter(readChannel)
	uftserver.Serve()
}
