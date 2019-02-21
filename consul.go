package nrpc

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
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
			log.Printf("[nrpc server] try port %d failure", port)
		} else {
			log.Printf("[nrpc server] try port %d success", port)
			localPort = port
			break
		}
	}
	if localPort == 0 {
		return errors.New("[nrpc server] server tcp accept failure: no port can use")
	}

	consulsSuccess := false
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
			log.Printf("[consul] try consul address %s failure, consul client error: %s", consulAddress, err)
			continue
		}

		err = client.Agent().ServiceRegister(service)
		if err != nil {
			log.Printf("[consul] try consul address %s failure, register server error: %s", consulAddress, err)
			continue
		}
		log.Printf("[consul] Service Registed, try consul address %s success", consulAddress)
		consulsSuccess = true
		server.onShutdowns = append(server.onShutdowns, func(server *Server) {
			log.Println("[consul] Service Deregisted")
			client.Agent().ServiceDeregister(serviceID)
		})
		break
	}
	if !consulsSuccess {
		return errors.New("[consul] register server failure")
	}
	return server.Accept(listen)
}

func (c *Client) getServices() []*consulapi.ServiceEntry {
	for _, consulAddress := range c.consuls {
		config := consulapi.DefaultConfig()
		config.Address = consulAddress
		config.HttpClient = http.DefaultClient
		client, err := consulapi.NewClient(config)
		if err != nil {
			log.Println("[consul] try consul address " + consulAddress + " failure, consul client error: " + err.Error())
			continue
		}

		addrs, _, err := client.Health().Service(c.serviceName, "", true, nil)
		if err != nil {
			log.Println("[consul] try consul address " + consulAddress + " failure, discovery server error: " + err.Error())
			continue
		}
		log.Println("[consul] try consul address " + consulAddress + " success")
		return addrs
	}
	return nil
}
