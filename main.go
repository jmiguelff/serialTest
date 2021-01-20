package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/tarm/serial"
	"gopkg.in/yaml.v2"
)

type serialConfigT struct {
	SerialMode struct {
		Name     string `yaml:"name"`
		Device   string `yaml:"device"`
		DataSize int    `yaml:"dataSize"`
		Baud     int    `yaml:"baud"`
		Stopbits int    `yaml:"stopbits"`
		Parity   string `yaml:"parity"`
		Timeout  int    `yaml:"timeout"`
	} `yaml:"serial"`
}

func setSerialMode(c *serialConfigT) *serial.Config {
	s := new(serial.Config)
	s.Name = c.SerialMode.Device
	s.Baud = c.SerialMode.Baud
	s.Size = byte(c.SerialMode.DataSize)
	s.StopBits = serial.StopBits(c.SerialMode.Stopbits)
	s.Parity = serial.Parity(c.SerialMode.Parity[0])
	s.ReadTimeout = time.Millisecond * time.Duration(c.SerialMode.Timeout)

	return s
}

func main() {
	// Open settings file
	fd, err := ioutil.ReadFile("settings.yml")
	if err != nil {
		log.Fatalln(err)
	}

	// Parse settings file (YAML)
	opts := new(serialConfigT)
	err = yaml.Unmarshal(fd, opts)
	if err != nil {
		log.Fatalln(err)
	}

	// Serial port configuration
	mode := setSerialMode(opts)

	log.Println("Open serial port")
	log.Println(*mode)

	sfd, err := serial.OpenPort(mode)
	if err != nil {
		log.Fatalln(err)
	}
	defer sfd.Close()

	// Sibas16 commands
	cmd := []byte{'S', 'F', 'S', 'P'}

	// Send 'S'
	_, err = sfd.Write(cmd[:1])
	if err != nil {
		log.Fatalln(err)
	}

	reader := bufio.NewReader(sfd)
	res, err := reader.ReadByte()
	if err != nil {
		log.Fatalln(err)
	}

	if res != cmd[0] {
		log.Fatalln("Command 'S' does not match")
	}

	// Send 'P'
	_, err = sfd.Write(cmd[1:2])
	if err != nil {
		log.Fatalln(err)
	}

	reply, err := reader.ReadBytes('P')
	if err != nil {
		log.Fatalln(err)
	}

	if bytes.Equal(reply, cmd[1:]) != true {
		log.Fatalln("Command 'P' does not get 'FSP' echo")
	}

	// Send Enter
	enterCmd := []byte{0x0D}
	_, err = sfd.Write(enterCmd)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Receive serial data")

	// Receive all bytes
	var buf []byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			log.Println("Error or end of data, hopefuly we have some data")
			break
		}
		buf = append(buf, b)
	}

	fp, err := os.Create("output.bin")
	if err != nil {
		log.Fatalln(err)
	}
	defer fp.Close()

	_, err = fp.WriteString(hex.Dump(buf))
	if err != nil {
		log.Fatalln(err)
	} else {
		fp.Sync()
	}

	// fmt.Println(hex.Dump(buf))
}
