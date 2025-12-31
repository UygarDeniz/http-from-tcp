package main

import (
	"fmt"
	"log"
	"net"

	"github.com/uygardeniz/http-from-tcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		fmt.Println("Connection accepted")
		if err != nil {
			log.Fatal(err)
		}

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Println("Error parsing request:", err)
			conn.Close()
			continue
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Println("Body:")
		fmt.Println(string(req.Body))

		conn.Close()
	}

}
