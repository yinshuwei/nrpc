package main

import (
	"log"

	"github.com/yinshuwei/nrpc"
)

// Args Args
type Args struct {
	A, B int
}

// Reply Reply
type Reply struct {
	C int
}

// Arith Arith
type Arith struct{}

// Add Add
func (t *Arith) Add(args *Args, reply *Reply) error {
	reply.C = args.A + args.B
	log.Println(*args)
	return nil
}

func main() {
	sever := nrpc.NewServer()
	sever.Register(&Arith{})
	err := sever.Serve(
		"carts_v4", // server_name
		[]string{"127.0.0.1:8109", "127.0.0.1:8501", "127.0.0.1:8502", "127.0.0.1:8503"}, // consol address
		"172.17.0.1",                        // local ip
		[]int{8001, 8002, 8003, 8004, 8005}, // local ports
	)
	if err != nil {
		log.Println(err)
	}
}
