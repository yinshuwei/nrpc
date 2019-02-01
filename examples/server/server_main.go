package main

import (
	"log"
	"net"
	"nrpc"
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
	listen, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	sever.Accept(listen)
}
