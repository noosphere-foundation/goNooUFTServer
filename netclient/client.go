package netclient

import (
	"bufio"
	"io"
	"net"

	"log"
	"time"
)

type NetworkType int

const (
	UDP NetworkType = iota
	TCP
)

func (n NetworkType) String() string {
	//

	if n == UDP {
		//

		return "udp"
	}

	if n == TCP {
		//

		return "tcp"
	}

	return ""
}

type Client struct {
	address       string
	protocol      NetworkType
	outcomingData chan *[]byte
	incomingData  chan []byte
	writer        *bufio.Writer
	reader        *bufio.Reader
	scanner       *bufio.Scanner
	connection    net.Conn
	connected     bool
}

func New(address string, protocol NetworkType, readChannel chan []byte) *Client {
	//

	c := new(Client)
	c.address = address
	c.protocol = protocol
	c.outcomingData = make(chan *[]byte, 1024)
	c.incomingData = readChannel

	return c
}

func (c *Client) write() {
	//

	for {
		//

		select {

		case datap := <-c.outcomingData:
			data := *datap
			if !c.connected {
				//

				if !c.reconnect() {
					//

					break
				}
			}

			err := c.connection.SetWriteDeadline(time.Now().Add(200 * time.Millisecond))
			if err != nil {
				//

				log.Println("SETTINGS TIMEOUT")
				c.Close()
				break
			}

			if data[len(data)-1] != '\n' {
				//

				data = append(data, '\n')
			}

			_, err = c.writer.Write(data)
			if err != nil {
				//

				log.Println("WRITING")
				c.Close()
				break
			}

			err = c.writer.Flush()
			if err != nil {
				//

				log.Println("fLUSHING: ", err)
				c.Close()
				break
			}

			if c.protocol == TCP {
				//

				tmp, err := c.reader.ReadBytes('\n')
				if err == io.EOF {
					//
					log.Println("CLIENT DISCONNECTED", string(tmp))

					break

				} else if err != nil {
					//
					log.Println("!!", err)

					break
				}

				c.incomingData <- tmp[:len(tmp)-1]
			}
		}
	}
}

func (c Client) CheckChan() int {
	//

	return len(c.outcomingData)
}

func (c Client) Write(data *[]byte) {
	//

	c.outcomingData <- data
}

func (c Client) Start() {
	//

	c.reconnect()

	go c.write()
}

func (c Client) reconnect() bool {
	//

	log.Println("TIME TO RECONNECT: ", c.protocol)

	if c.connected {
		//

		return true
	}

	if c.protocol == UDP {
		//

		addr, err := net.ResolveUDPAddr(c.protocol.String(), c.address)
		if err != nil {
			//

			log.Println("ERROR RESOLVING ADDRESS. ", err)
			return false
		}

		c.connection, err = net.DialUDP(c.protocol.String(), nil, addr)
		if err != nil {
			//

			log.Println(err)
			return false
		}

	} else {
		//

		addr, err := net.ResolveTCPAddr(c.protocol.String(), c.address)
		if err != nil {
			//

			log.Println("ERROR RESOLVING ADDRESS. ", err)
			return false
		}

		c.connection, err = net.DialTCP(c.protocol.String(), nil, addr)
		if err != nil {
			//rea

			log.Println(err)
			return false
		}
	}

	c.writer = bufio.NewWriter(c.connection)
	c.reader = bufio.NewReaderSize(c.connection, 16384)

	c.connected = true

	return true
}

func (c Client) Close() {
	//

	c.connection.Close()
	c.connected = false
}
