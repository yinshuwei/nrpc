// NRPC client 从gob改写成msgpack

package nrpc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack"
)

type clientCodec struct {
	rwc    io.ReadWriteCloser
	dec    *msgpack.Decoder
	enc    *msgpack.Encoder
	encBuf *bufio.Writer
}

func (c *clientCodec) WriteRequest(r *rpc.Request, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		return
	}
	if err = c.enc.Encode(body); err != nil {
		return
	}
	return c.encBuf.Flush()
}

func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	return c.dec.Decode(r)
}

func (c *clientCodec) ReadResponseBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *clientCodec) Close() error {
	return c.rwc.Close()
}

// NewClient returns a new rpc.Client to handle requests to the
// set of services at the other end of the connection.
func NewClient(conn io.ReadWriteCloser) *rpc.Client {
	encBuf := bufio.NewWriter(conn)
	client := &clientCodec{conn, msgpack.NewDecoder(conn), msgpack.NewEncoder(encBuf), encBuf}
	return rpc.NewClientWithCodec(client)
}

// Client Client
type Client struct {
	serviceName string
	consuls     []string
	seq         int

	netClients  []*netClient
	refreshTime time.Time
	mutex       sync.Mutex
}

type netClient struct {
	id string
	c  *rpc.Client
}

func (client *Client) refreshClients() {
	ncMap := map[string]*netClient{}
	for _, nc := range client.netClients {
		ncMap[nc.id] = nc
	}

	services := client.getServices()
	var newNetClients []*netClient
	for _, s := range services {
		nc, ok := ncMap[s.Service.ID]
		if ok && nc != nil {
			newNetClients = append(newNetClients, nc)
		} else {
			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.Service.Address, s.Service.Port))
			if err != nil {
				log.Println(err)
				continue
			}
			nc = &netClient{s.Service.ID, NewClient(conn)}
			newNetClients = append(newNetClients, nc)
		}
	}
	client.netClients = newNetClients
}

func (client *Client) getNetClient(needRefresh bool) (*netClient, error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	now := time.Now()
	if len(client.netClients) == 0 || now.Sub(client.refreshTime) > 5*time.Second || needRefresh {
		client.refreshClients()
		client.refreshTime = now
	}

	clientLen := len(client.netClients)
	if clientLen == 0 {
		return nil, errors.New("no can use service")
	}
	nc := client.netClients[client.seq%clientLen]
	client.seq++
	return nc, nil
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client *Client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	nc, err := client.getNetClient(false)
	if err != nil {
		return nil
	}
	err = nc.c.Call(serviceMethod, args, reply)
	if err != nil {
		if err.Error() == "connection is shut down" {
			log.Println("connection is shut down, refersh client and get new one!")
			nc, err = client.getNetClient(true)
			if err != nil {
				return nil
			}
			err = nc.c.Call(serviceMethod, args, reply)
		}
	}
	return err
}

// Dial Dial
func Dial(serviceName string, consuls []string) (*Client, error) {
	return &Client{
		serviceName: serviceName,
		consuls:     consuls,
	}, nil

}
