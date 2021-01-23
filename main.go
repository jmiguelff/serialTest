package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
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
	fmt.Println("Select mode")
	fmt.Println("1. Run SFSP commmad as client")
	fmt.Println("2. Run SFSP command as server")

	var option int
	fmt.Scanf("%d", &option)
	fmt.Println("Option ", option, "selected")

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

	if option == 1 {
		useSFSP(sfd)
	} else if option == 2 {
		simSFSP(sfd)
	} else {
		log.Println("Unknown option")
	}
}

func useSFSP(sfd *serial.Port) {
	// Sibas16 commands
	cmd := []byte{'S', 'F', 'S', 'P'}

	// Send 'S'
	_, err := sfd.Write(cmd[:1])
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

	// Send 'F'
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

// TODO: Rename struct to use it to pass also the file name
func simSFSP(sfd *serial.Port) {
	// open binary file
	fd, err := os.Open("test.bin")
	if err != nil {
		log.Fatalln(err)
		return
	}

	s := bufio.NewScanner(fd)
	var buf []byte
	for s.Scan() {
		arr := strings.Split(s.Text(), " ")
		for _, v := range arr {
			b, err := hex.DecodeString(v)
			if err != nil {
				log.Fatalln(err)
				return
			}
			buf = append(buf, b...)
		}
	}
	// fmt.Println(hex.Dump(buf))
	// Sibas16 commands
	cmd := []byte{'S', 'F', 'S', 'P'}

	// Receive 'S'
	reader := bufio.NewReader(sfd)
	res, err := reader.ReadByte()
	if err != nil {
		log.Fatalln(err)
	}

	if res != cmd[0] {
		log.Fatalln("Unknown command from client", reader)
	}

	// Send 'S'
	_, err = sfd.Write(cmd[:1])
	if err != nil {
		log.Fatalln(err)
	}

	// Read 'F'
	reader = bufio.NewReader(sfd)
	res, err = reader.ReadByte()
	if err != nil {
		log.Fatalln(err)
	}

	if res != cmd[1] {
		log.Fatalln("Unknown command from client", reader)
	}

	// Send 'F'
	_, err = sfd.Write(cmd[1:])
	if err != nil {
		log.Fatalln(err)
	}

	// Read '0x0d'
	reader = bufio.NewReader(sfd)
	res, err = reader.ReadByte()
	if err != nil {
		log.Fatalln(err)
	}

	if res != 0x0D {
		log.Fatalln("Unknown command from client", reader)
	}

	// Send 'F'
	_, err = sfd.Write(buf)
	if err != nil {
		log.Fatalln(err)
	}
}
