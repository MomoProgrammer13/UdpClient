package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
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

	//file, errorFile := os.Create(`C:\Users\gfanha\Documents\testeUDP\` + "excelPlantonistas.xlsx")
	//if errorFile != nil {
	//	log.Fatal(errorFile)
	//}
	//defer file.Close()

	conn, err := net.DialUDP(TYPE, nil, udpServer)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	_, err = conn.Write(request)
	response, err := bufio.NewReaderSize(conn, 1024).ReadBytes()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(response))
	//data := make(chan byte)
	//done := make(chan bool)
	//wg := sync.WaitGroup{}
	//var finish bool = false

	//for !finish {
	//	received := make([]byte, 1024)
	//	_, _, err = conn.ReadFromUDP(received)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	wg.Add(1)
	//	go handleIncomingResponse(received, data, &wg, &finish)
	//}
	//if len(received) > 0 {
	//	errorDec := dec.Decode(&q)
	//	if errorDec != nil {
	//		log.Fatal(errorDec)
	//	}
	//	_, err2 := file.WriteAt(q.Data, int64(i*900))
	//	if err2 != nil {
	//		return
	//	}
	//	fmt.Println(q)
	//	if q.Reps == 0 {
	//		break
	//	}
	//
	//}
	//go writeToFile(data, done)
	//go func() {
	//	wg.Wait()
	//	close(data)
	//}()
	//d := <-done
	//if d {
	//	fmt.Println("File written successfully")
	//} else {
	//	fmt.Println("File not written")
	//}
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

func handleIncomingResponse(buffer []byte, data chan byte, wg *sync.WaitGroup, finish *bool) {
	dec := gob.NewDecoder(bytes.NewReader(buffer))
	var q MetaData
	errorDec := dec.Decode(&q)
	if errorDec != nil {
		log.Fatal(errorDec)
	}
	fmt.Printf("Received: %v FileName: %v - Part: %d\n  Data = %v\n", q, q.Name, q.Reps, q.Data)
	fmt.Printf("Data Size: %d\n", len(q.Data))
	if len(q.Data) > 0 {
		for _, d := range q.Data {
			data <- d
		}
	}
	wg.Done()
	if q.Reps == 0 {
		fmt.Println("Arquivo 0")
		*finish = true
	}
}

func writeToFile(data chan byte, done chan bool) {
	f, err := os.Create("concurrent")
	if err != nil {
		log.Fatal(err)
	}
	for d := range data {
		_, err = fmt.Fprintln(f, d)
		if err != nil {
			fmt.Println(err)
			f.Close()
			done <- false
			return
		}
	}
	err = f.Close()
	if err != nil {
		log.Fatal(err)
		done <- false
		return
	}
	done <- true
}
