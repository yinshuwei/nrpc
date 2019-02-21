package main

import (
	"log"
	"nrpc"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	cli, err := nrpc.Dial(
		"carts_v4", // server_name
		[]string{"127.0.0.1:8109", "127.0.0.1:8501", "127.0.0.1:8502", "127.0.0.1:8503"}, // consol address
	)
	if err != nil {
		log.Println(err)
	}

	reply := map[string]int{}
	err = cli.Call("Arith.Add", map[string]int{"A": 1, "B": 5}, &reply)
	if err != nil {
		log.Println(err)
	}

	log.Println(reply["C"])

}
