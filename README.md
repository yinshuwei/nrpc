# nrpc

## 基于consul、TCP、msgpack的简单RPC

### Demo

#### 安装consul

docker-compose.yml

    version: '3'

    services:

    consul-agent-1:
        image: consul:latest
        networks:
        - consul-demo
        command: "agent -retry-join consul-server-bootstrap -client 0.0.0.0"
        ports:
        - "8501:8500"

    consul-agent-2: 
        image: consul:latest
        networks:
        - consul-demo
        command: "agent -retry-join consul-server-bootstrap -client 0.0.0.0"
        ports:
        - "8502:8500"

    consul-agent-3: 
        image: consul:latest
        networks:
        - consul-demo
        command: "agent -retry-join consul-server-bootstrap -client 0.0.0.0"
        ports:
        - "8503:8500"

    consul-server-1: 
        image: consul:latest
        networks:
        - consul-demo
        command: "agent -server -retry-join consul-server-bootstrap -client 0.0.0.0"

    consul-server-2:
        image: consul:latest
        networks:
        - consul-demo
        command: "agent -server -retry-join consul-server-bootstrap -client 0.0.0.0"

    consul-server-bootstrap:
        image: consul:latest
        networks:
        - consul-demo
        ports:
        - "8400:8400"
        - "8500:8500"
        - "8600:8600"
        - "8600:8600/udp"
        command: "agent -server -bootstrap-expect 3 -ui -client 0.0.0.0"

    networks:
    consul-demo:

docker-compose
    docker-compost up

#### go get
    go get -u -v github.com/yinshuwei/nrpc

#### 服务端代码
server.go

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

#### 客户端代码

client.go

    package main

    import (
        "log"
        "github.com/yinshuwei/nrpc"
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
