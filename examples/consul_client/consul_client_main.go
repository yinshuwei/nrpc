package main

import (
	"log"
	"nrpc"
	"time"
)

func main() {
	cli, err := nrpc.Dial("haha", []string{"127.0.0.1:8502", "127.0.0.1:8501"})
	if err != nil {
		log.Println(err)
	}

	// initNum := 10000000

	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// go run(cli, initNum)

	// initNum += 10000000
	// run(cli, 10000000)

	// initNum += 10000000
	run(cli, 10000000)

}

func run(cli *nrpc.Client, initNum int) {
	var err error
	for i := 0; i < 1000000; i++ {
		reply := map[string]int{}
		err = cli.Call("Arith.Add", map[string]int{"A": 1, "B": i + initNum}, &reply)
		if err != nil {
			log.Println(err)
		}
		if reply["C"] != 1+i+initNum {
			log.Println("error")
		}
		time.Sleep(time.Second)
	}
}
