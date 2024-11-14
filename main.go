package main

import (
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

type ResponseMetaData struct {
	Name     string
	FileSize int64
	Reps     uint32
	Data     []byte
}

type RequestMetaData struct {
	Name string
	Reps uint32
	Miss bool
}

type Packet struct {
	Reps uint32
	Data []byte
}

type mutexMapData struct {
	sync.RWMutex
	m map[int][]byte
}

func (meta RequestMetaData) RequestMetaDataToBytes() []byte {
	var metaBytes bytes.Buffer
	enc := gob.NewEncoder(&metaBytes)

	err := enc.Encode(meta)
	if err != nil {
		log.Fatal(err)
	}
	return metaBytes.Bytes()
}

func main() {
	request := make([]byte, 1024)

	request = RequestMetaData{
		Name: "img_banco.jpg",
		Reps: 0,
		Miss: false,
	}.RequestMetaDataToBytes()

	udpServer, err := net.ResolveUDPAddr(TYPE, HOST+":"+PORT)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP(TYPE, nil, udpServer)
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Write(request)

	infoBuffer := make([]byte, 1024)
	_, _, errBuffer := conn.ReadFromUDP(infoBuffer)
	if errBuffer != nil {
		log.Fatal(err)
	}

	infoDecode := gob.NewDecoder(bytes.NewReader(infoBuffer))
	var infoData ResponseMetaData
	errorDecode := infoDecode.Decode(&infoData)
	if errorDecode != nil {
		log.Fatal(errorDecode)
	}

	dados := mutexMapData{m: make(map[int][]byte)}

	fmt.Println(infoData.Reps)
	wg := sync.WaitGroup{}

	for i := uint32(0); i <= infoData.Reps; i++ {

		if i != 0 && i%10 == 0 {
			request = RequestMetaData{
				Name: "img_banco.jpg",
				Reps: i * 10,
				Miss: false,
			}.RequestMetaDataToBytes()
			_, err = conn.Write(request)
			if err != nil {
				log.Fatal(err)
			}
		}
		received := make([]byte, 1024)
		_, _, err = conn.ReadFromUDP(received)
		if err != nil {
			log.Fatal(err)
		}
		wg.Add(1)
		go handleIncomingResponse(received, &wg, &dados)
	}
	wg.Wait()
	//f, err := os.Create(`ordenacao.txt`)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer f.Close()
	//dados.RLock()
	//fmt.Println("Printando o map de dados")
	//keys := make([]int, 0, len(dados.m))
	//for k := range dados.m {
	//	keys = append(keys, k)
	//}
	//fmt.Println("Tamanho da Keys ", len(keys))
	//sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	//for _, i := range keys {
	//	_, err = f.Write(dados.m[i])
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//}
	//dados.RUnlock()
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

func handleIncomingResponse(buffer []byte, wg *sync.WaitGroup, dados *mutexMapData) {
	defer wg.Done()
	dec := gob.NewDecoder(bytes.NewReader(buffer))
	var q Packet
	errorDec := dec.Decode(&q)
	if errorDec != nil {
		log.Fatal(errorDec)
	}

	fmt.Printf("Part %d -  Data Size: %d\n", q.Reps, len(q.Data))
	dados.Lock()
	dados.m[int(q.Reps)] = q.Data
	dados.Unlock()
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
