package client

import (
	"fmt"
	"net"

	"github.com/zeebo/bencode"
)

type Message struct {
	Inst map[string]interface{}
}

type Response struct {
	Ex     string   `bencode:"ex"`
	Out    string   `bencode:"out"`
	Err    string   `bencode:"err"`
	Value  string   `bencode:"value"`
	Status []string `bencode:"status"`

	NewSession    string `bencode:"new-session"`    // for clone
	FormattedCode string `bencode:"formatted-code"` // for format-code
}

type ResponseHandler func(r Response)

type Client struct {
	conn net.Conn
	enc  *bencode.Encoder
	dec  *bencode.Decoder
}

type EvalErr string

func (e EvalErr) Error() string {
	return fmt.Sprintf("error evaluating `%s`\n", e)
}

// NewClient creates a client that communicates with a nREPL server
// at the given TCP address.
func NewClient(addr string) (Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return Client{}, err
	}
	return Client{
		conn: conn,
		enc: bencode.NewEncoder(conn),
		dec: bencode.NewDecoder(conn),
	}, nil
}

func (c Client) Close() {
	c.conn.Close()
}

func (c Client) Send(m Message, rh ResponseHandler) (err error) {
	if err := c.enc.Encode(m.Inst); err != nil {
		return fmt.Errorf("error writing instruction: %v", err)
	}

	for {
		resp := Response{}
		if err := c.dec.Decode(&resp); err != nil {
			return fmt.Errorf("error decoding response: %v", err)
		}
		rh(resp)
		if len(resp.Status) > 0 {
			if resp.Status[0] == "done" {
				break
			} else if resp.Status[0] == "eval-error" {
				err = EvalErr(m.Inst["code"].(string))
			}
		}
	}
	return
}
