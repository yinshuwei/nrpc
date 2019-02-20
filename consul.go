package nrpc

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"

	consulapi "github.com/hashicorp/consul/api"
)

func (server *Server) consulServe(serviceName string, consulAddress string, localIP string, portRange []int) error {
	var err error
	localPort := 0
	var listen net.Listener
	for _, port := range portRange {
		tcp := fmt.Sprintf(":%d", port)
		listen, err = net.Listen("tcp", tcp)
		if err != nil {
			log.Printf("try port %d fail", port)
		} else {
			log.Printf("try port %d success", port)
			localPort = port
			break
		}
	}
	if localPort == 0 {
		return errors.New("server tcp accept fail: no port can use")
	}

	config := consulapi.DefaultConfig()
	config.Address = consulAddress
	client, err := consulapi.NewClient(config)
	if err != nil {
		return errors.New("consul client error: " + err.Error())
	}
	serviceID := fmt.Sprintf("%s:%s:%d", serviceName, localIP, localPort)

	err = client.Agent().ServiceRegister(&consulapi.AgentServiceRegistration{
		Port:    localPort,
		Name:    serviceName,
		ID:      serviceID,
		Tags:    []string{serviceName, localIP, strconv.Itoa(localPort), serviceID},
		Address: localIP,
		Check: &consulapi.AgentServiceCheck{
			TCP:                            fmt.Sprintf("%s:%d", localIP, localPort),
			Timeout:                        "2s",
			Interval:                       "2s",
			DeregisterCriticalServiceAfter: "2s",
		},
	})
	if err != nil {
		return errors.New("register server error: " + err.Error())
	}
	return server.Accept(listen)
}
