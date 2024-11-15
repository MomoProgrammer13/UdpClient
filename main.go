package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"hash/crc32"
	"log"
	"math"
	"net"
	"os"
	"sort"
	"sync"
	"time"
)

const (
	HOST       = "localhost"
	PORT       = "9001"
	TYPE       = "udp"
	BUFFERSIZE = 2048
)

type ResponseMetaData struct {
	Name     string
	FileSize int64
	Reps     uint32
	Msg      string
}

type RequestMetaData struct {
	Name string
	Reps uint32
	Miss bool
}

type Packet struct {
	Reps     uint32
	Checksum uint32
	Data     []byte
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

func calculateChecksum(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

func verifyChecksum(data []byte, checksum uint32) bool {
	return crc32.ChecksumIEEE(data) == checksum
}

func main() {

	request := make([]byte, BUFFERSIZE)

	fmt.Println("Digite o nome do arquivo que deseja baixar")
	in := bufio.NewScanner(os.Stdin)
	in.Scan()
	request = RequestMetaData{
		Name: in.Text(),
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
	if err != nil {
		log.Fatal(err)
	}
	err = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		log.Fatal(err)
	}

	infoBuffer := make([]byte, BUFFERSIZE)
	_, _, errBuffer := conn.ReadFromUDP(infoBuffer)
	if errBuffer != nil {
		log.Fatal("Servidor não teve resposta ao cliente")
	}

	err = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		return
	}

	infoDecode := gob.NewDecoder(bytes.NewReader(infoBuffer))
	var infoData ResponseMetaData
	errorDecode := infoDecode.Decode(&infoData)
	if errorDecode != nil {
		log.Fatal("Falha ao interpretar a resposta do servidor")
	}

	if infoData.Name == "__ERROR__" {
		log.Fatal(infoData.Msg)
	}

	fmt.Printf("Recebendo arquivo %s\n", infoData.Name)

	dados := mutexMapData{m: make(map[int][]byte)}

	wg := sync.WaitGroup{}
	f, err := os.Create(fmt.Sprintf("Downloads/%s", infoData.Name))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	start := time.Now()

	var bar Bar
	bar.NewOption(0, int64(infoData.FileSize))
	totalEscrito := 0

	for i := uint32(0); i <= infoData.Reps; i++ {
		if i != 0 && i%10 == 0 {
			wg.Wait()
			// Check the bytes received to write to the file
			AppendFile(&dados, f, &totalEscrito, &bar)
			dados.Lock()
			dados.m = make(map[int][]byte)
			dados.Unlock()
			request = RequestMetaData{
				Name: in.Text(),
				Reps: i,
				Miss: false,
			}.RequestMetaDataToBytes()
			_, err = conn.Write(request)
			if err != nil {
				log.Fatal(err)
			}
		}
		received := make([]byte, BUFFERSIZE)
		_, _, err = conn.ReadFromUDP(received)
		if err != nil {
			if i == 0 {
				log.Fatal("Falha ao se conectar com o servidor Linha", err)
			}
			esperado := func() int {
				if i/10 < infoData.Reps/10 {
					return 10
				}
				return int(infoData.Reps%10 + 1)
			}
			a := esperado()
			if len(dados.m) != a {
				partsVerificar := (i - i%10) + uint32(a)
				for j := i - i%10; j < partsVerificar; j++ {
					if _, ok := dados.m[int(j)]; !ok {
						request = RequestMetaData{
							Name: in.Text(),
							Reps: uint32(j),
							Miss: true,
						}.RequestMetaDataToBytes()
						fmt.Println("  - Pedindo Novamente pacote", j)
						_, err = conn.Write(request)
						if err != nil {
							log.Fatal("Falha ao se conectar com o servidor", err)
						}
						conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
						_, _, err = conn.ReadFromUDP(received)
						if err != nil {
							log.Fatal("Falha ao se conectar com o servidor Linha")
						}
						wg.Add(1)
						go handleIncomingResponse(received, &wg, &dados)
						break
					}
				}
			} else {
				log.Fatal("Falha ao se conectar com o servidor Linha", err)
			}
		} else {
			conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			wg.Add(1)
			go handleIncomingResponse(received, &wg, &dados)
		}
	}
	wg.Wait()
	AppendFile(&dados, f, &totalEscrito, &bar)

	bar.Finish()
	elapsed := time.Since(start)
	log.Printf("File Transfer took %.2fs", elapsed.Seconds())
}
func AppendFile(dados *mutexMapData, f *os.File, totalEscrito *int, bar *Bar) {
	dados.Lock()
	keys := make([]int, 0, len(dados.m))
	for k := range dados.m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, i := range keys {
		*totalEscrito += len(dados.m[i])
		_, err := f.Write(dados.m[i])
		if err != nil {
			log.Fatal(err)
		}
	}

	bar.Play(int64(*totalEscrito))
	dados.Unlock()
}

func handleIncomingResponse(buffer []byte, wg *sync.WaitGroup, dados *mutexMapData) {
	defer wg.Done()
	dec := gob.NewDecoder(bytes.NewReader(buffer))
	var q Packet
	errorDec := dec.Decode(&q)
	if errorDec != nil {
		log.Fatal(errorDec)
	}

	if !verifyChecksum(q.Data, q.Checksum) {
		log.Fatal("Checksum failed")
	}

	dados.Lock()
	dados.m[int(q.Reps)] = q.Data
	dados.Unlock()
}

func prettyByteSize(b int64) string {
	bf := float64(b)
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f%sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}
