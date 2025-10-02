package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")

	if err != nil {
		log.Printf("NO connection found: %v\n", err)
	}

	udpConn, err := net.DialUDP(udpAddr.Network(), nil, udpAddr)

	if err != nil {
		log.Fatal(err)
	}

	defer udpConn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		value, err := reader.ReadString(byte('\n'))

		if err != nil {
			fmt.Println("\nFailed to read: ", err)
		}

		if _, err := udpConn.Write([]byte(value)); err != nil {
			fmt.Printf("Failed to send data: %v\n", err)
		}
	}
}
