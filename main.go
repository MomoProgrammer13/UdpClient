package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
)

func main() {
	p := make([]byte, 2048)
	conn, err := net.Dial("udp", "127.0.0.1:1234")
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	charset := "abcdefghijklmnopqrstuvwxyz"
	text := make([]byte, 2072)
	for i := range text {
		text[i] = charset[rand.Intn(len(charset))]
	}

	fmt.Fprintf(conn, string(text))
	_, err = bufio.NewReader(conn).Read(p)
	if err == nil {
		fmt.Printf("%s\n", p)
	} else {
		fmt.Printf("Some error %v\n", err)
	}
	conn.Close()
}
