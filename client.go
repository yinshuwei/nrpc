// NRPC client 从gob改写成msgpack

package nrpc

import (
	"bufio"
	"io"
	"net"
	"net/rpc"

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

// Dial connects to a MSGPACK-RPC server at the specified network address.
func Dial(network, address string) (*rpc.Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), err
}
