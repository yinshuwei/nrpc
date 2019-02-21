package nrpc

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"

	consulapi "github.com/hashicorp/consul/api"
)

func (server *Server) consulServe(serviceName string, consuls []string, localIP string, portRange []int) error {
	var err error
	localPort := 0
	var listen net.Listener
	for _, port := range portRange {
		tcp := fmt.Sprintf(":%d", port)
		listen, err = net.Listen("tcp", tcp)
		if err != nil {
			log.Printf("try port %d fail", port)
		} else {
			log.Printf("try port %d succ", port)
			localPort = port
			break
		}
	}
	if localPort == 0 {
		return errors.New("server tcp accept fail: no port can use")
	}

	consulsSucc := false
	serviceID := fmt.Sprintf("%s:%s:%d", serviceName, localIP, localPort)
	service := &consulapi.AgentServiceRegistration{
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
	}
	for _, consulAddress := range consuls {
		config := consulapi.DefaultConfig()
		config.Address = consulAddress
		client, err := consulapi.NewClient(config)
		if err != nil {
			log.Printf("try consul address %s fail, consul client error: %s", consulAddress, err)
			continue
		}

		err = client.Agent().ServiceRegister(service)
		if err != nil {
			log.Printf("try consul address %s fail, register server error: %s", consulAddress, err)
			continue
		}
		log.Printf("try consul address %s succ", consulAddress)
		consulsSucc = true
		server.onShutdowns = append(server.onShutdowns, func(server *Server) {
			log.Println("Service Deregister")
			client.Agent().ServiceDeregister(serviceID)
		})
		break
	}
	if !consulsSucc {
		return errors.New("register server fail")
	}
	return server.Accept(listen)
}

func (c *Client) getServices() []*consulapi.ServiceEntry {
	for _, consulAddress := range c.consuls {
		config := consulapi.DefaultConfig()
		config.Address = consulAddress
		client, err := consulapi.NewClient(config)
		if err != nil {
			log.Printf("try consul address %s fail, consul client error: %s", consulAddress, err)
			continue
		}

		addrs, _, err := client.Health().Service(c.serviceName, "", true, nil)
		if err != nil {
			log.Printf("try consul address %s fail, discovery server error: %s", consulAddress, err)
			continue
		}
		log.Printf("try consul address %s succ", consulAddress)
		return addrs
	}
	return nil
}
