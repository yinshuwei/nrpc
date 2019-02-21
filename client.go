// NRPC client 从gob改写成msgpack

package nrpc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/rpc"
	"strings"
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
	seq         uint32

	netClients  []*netClient
	refreshTime time.Time
	mutex       sync.Mutex
}

type netClient struct {
	id string
	c  *rpc.Client

	enable bool
}

func (client *Client) refreshClients() {
	ncMap := map[string]*netClient{}
	for _, nc := range client.netClients {
		ncMap[nc.id] = nc
		nc.enable = false
	}

	services := client.getServices()
	var newNetClients []*netClient
	for _, s := range services {
		nc, ok := ncMap[s.Service.ID]
		if ok && nc != nil {
			newNetClients = append(newNetClients, nc)
			nc.enable = true
		} else {
			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.Service.Address, s.Service.Port))
			if err != nil {
				log.Println("[nrpc client] tcp Dial failure, " + err.Error())
				continue
			}
			nc = &netClient{s.Service.ID, NewClient(conn), true}
			log.Println("[nrpc client] new conn, " + nc.id)
			newNetClients = append(newNetClients, nc)
		}
	}

	for _, nc := range client.netClients {
		if !nc.enable {
			log.Println("[nrpc client] close conn, " + nc.id)
			nc.c.Close()
		}
	}

	client.netClients = newNetClients
}

func (client *Client) getNetClient(oldNc *netClient) (*netClient, error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if oldNc != nil {
		oldNc.enable = false
	}

	now := time.Now()
	if len(client.netClients) == 0 || now.Sub(client.refreshTime) > 5*time.Second {
		client.refreshClients()
		client.refreshTime = now
	}

	clientLen := len(client.netClients)
	var nc *netClient
	tryCount := 0
	for nc == nil || !nc.enable {
		if tryCount >= clientLen {
			return nil, errors.New("[nrpc client] call service failure, no available services")
		}
		nc = client.netClients[client.seq%uint32(clientLen)]
		client.seq++
		tryCount++
	}
	if client.seq > math.MaxUint32-100000 {
		client.seq = 0
	}
	return nc, nil
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client *Client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	nc, err := client.getNetClient(nil)
	if err != nil {
		return err
	}
	err = nc.c.Call(serviceMethod, args, reply)
	if err != nil {
		errorText := err.Error()
		if strings.Contains(errorText, "connection is shut down") ||
			strings.Contains(errorText, "read: connection reset by peer") ||
			strings.Contains(errorText, "unexpected EOF") ||
			strings.Contains(errorText, "write: broken pipe") {
			log.Println("[nrpc client] " + errorText + ", refersh clients and use other one!")
			nc, err = client.getNetClient(nc)
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
