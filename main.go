package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	HOST = "localhost"
	PORT = "9001"
	TYPE = "udp"
)

type MetaData struct {
	Name     string
	FileSize int64
	Reps     int32
	Data     []byte
}

func main() {
	request := make([]byte, 1024)

	request = []byte("Hello from client")

	udpServer, err := net.ResolveUDPAddr(TYPE, HOST+":"+PORT)
	if err != nil {
		log.Fatal(err)
	}

	file, errorFile := os.Create(`C:\Users\gfanha\Documents\testeUDP\` + "Escala-Controle-Julho (11).xls")
	if errorFile != nil {
		log.Fatal(errorFile)
	}
	defer file.Close()

	conn, err := net.DialUDP(TYPE, nil, udpServer)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.Write(request)

	received := make([]byte, 1024)

	dec := gob.NewDecoder(bytes.NewReader(received))
	var q MetaData

	for i := 0; ; i++ {
		_, err = conn.Read(received)
		if err != nil {
			log.Fatal(err)
		}
		if len(received) > 0 {
			errorDec := dec.Decode(&q)
			if errorDec != nil {
				log.Fatal(errorDec)
			}
			_, err2 := file.WriteAt(q.Data, int64(i*900))
			if err2 != nil {
				return
			}
			fmt.Println(q)
			if q.Reps == 0 {
				break
			}

		}
	}
}
func AppendFile() {
	file, err := os.OpenFile(`C:\Users\gfanha\Documents\testeUDP\`+"Escala-Controle-Julho (11).xls", os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()

	len, err := file.WriteString(" The Go language was conceived in September 2007 by Robert Griesemer, Rob Pike, and Ken Thompson at Google.")
	if err != nil {
		log.Fatalf("failed writing to file: %s", err)
	}
	fmt.Printf("\nLength: %d bytes", len)
	fmt.Printf("\nFile Name: %s", file.Name())
}
