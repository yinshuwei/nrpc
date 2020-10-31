package main

import (
	"log"
	"net"

	"github.com/yinshuwei/nrpc"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Println(err)
	}

	cli := nrpc.NewClient(conn)

	for i := 0; i < 100; i++ {
		reply := map[string]int{}
		err = cli.Call("Arith.Add", map[string]int{"A": 1, "B": i}, &reply)
		if err != nil {
			log.Println(err)
		}

		log.Println(reply)
	}
}
