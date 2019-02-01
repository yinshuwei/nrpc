// NRPC server 从gob改写成msgpack

package nrpc

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/rpc"

	"github.com/vmihailenco/msgpack"
)

type serverCodec struct {
	rwc    io.ReadWriteCloser
	dec    *msgpack.Decoder
	enc    *msgpack.Encoder
	encBuf *bufio.Writer
	closed bool
}

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.dec.Decode(r)
}

func (c *serverCodec) ReadRequestBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *serverCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		if c.encBuf.Flush() == nil {
			// MSGPACK couldn't encode the header. Should not happen, so if it does,
			// shut down the connection to signal that the connection is broken.
			log.Println("rpc: msgpack error encoding response:", err)
			c.Close()
		}
		return
	}
	if err = c.enc.Encode(body); err != nil {
		if c.encBuf.Flush() == nil {
			// Was a msgpack problem encoding the body but the header has been written.
			// Shut down the connection to signal that the connection is broken.
			log.Println("rpc: msgpack error encoding body:", err)
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *serverCodec) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}

// Server Server
type Server struct {
	s *rpc.Server
}

// NewServer NewServer
func NewServer() *Server {
	return &Server{
		s: rpc.NewServer(),
	}
}

// Register Register
func (server *Server) Register(rcvr interface{}) error {
	return server.s.Register(rcvr)
}

// ServeConn runs the MSGPACK-RPC server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	buf := bufio.NewWriter(conn)
	srv := &serverCodec{
		rwc:    conn,
		dec:    msgpack.NewDecoder(conn),
		enc:    msgpack.NewEncoder(buf),
		encBuf: buf,
	}
	server.s.ServeCodec(srv)
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection. Accept blocks until the listener
// returns a non-nil error. The caller typically invokes Accept in a
// go statement.
func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Print("rpc.Serve: accept:", err.Error())
			return
		}
		go server.ServeConn(conn)
	}
}

// Serve Serve
func (server *Server) Serve(url string) {
	// if conn, err := amqp.Dial(url); err != nil {
	// 	failOnError(err, "Failed to connect to MQServer")
	// } else {
	// 	server.ServeConn(conn, queue)
	// }
}
