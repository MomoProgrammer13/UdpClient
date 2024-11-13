package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"sort"
	"strconv"
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

	_, err = conn.Write(request)

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	//Make a regex to only get numeric values and save to variable
	re := regexp.MustCompile(`\d+`)
	numericValues := re.FindAllString(response, -1)

	// Print the numeric values

	tamanho, err := strconv.Atoi(numericValues[0])
	if err != nil {
		log.Fatal(err)
	}
	var dados = struct {
		sync.RWMutex
		m map[int][]byte
	}{m: make(map[int][]byte)}

	fmt.Println(tamanho)
	wg := sync.WaitGroup{}

	for i := 0; i < tamanho; i++ {
		fmt.Println("Valor de i: ", i)
		received := make([]byte, 1024)
		_, _, err = conn.ReadFromUDP(received)
		if err != nil {
			log.Fatal(err)
		}
		wg.Add(1)
		go handleIncomingResponse(received, &wg, &dados)
	}

	wg.Wait()

	dados.RLock()
	fmt.Println("Printando o map de dados")
	keys := make([]int, 0, len(dados.m))
	for k := range dados.m {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	for _, i := range keys {
		fmt.Print(string(dados.m[i]))
	}
	dados.RUnlock()
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

func handleIncomingResponse(buffer []byte, wg *sync.WaitGroup, dados *struct {
	sync.RWMutex
	m map[int][]byte
}) {
	defer wg.Done()
	dec := gob.NewDecoder(bytes.NewReader(buffer))
	var q MetaData
	errorDec := dec.Decode(&q)
	if errorDec != nil {
		log.Fatal(errorDec)
	}
	fmt.Printf("FileName: %v - Part: %d\n  Data = %v\n", q, q.Name, q.Reps, q.Data)
	fmt.Printf("Data Size: %d\n", len(q.Data))
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
