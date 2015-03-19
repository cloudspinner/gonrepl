// An nREPL client that reads Clojure code from standard input and
// writes the result to standard output.  Connects to
// localhost:$LEIN_REPL_PORT by default. Pass the -a flag to override
// the default address.
//
// Exceptions and captured stderr go to standard error.
// Values and captured stdout go to standard output.
// Return with non-zero exit code if there was an evaluation error.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/zeebo/bencode"
)

type Response struct {
	Ex     string
	Out    string
	Err    string
	Value  string
	Status []string
}

var addr = flag.String("a", "localhost:"+os.Getenv("LEIN_REPL_PORT"), "nREPL port")

func main() {
	flag.Parse()

	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("error reading standard input: ", err)
	}
	code := string(bytes)
	inst := map[string]interface{}{
		"op":   "eval",
		"code": code,
	}

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatal("error connecting to "+*addr+": ", err)
	}
	defer conn.Close()

	enc := bencode.NewEncoder(conn)
	if err := enc.Encode(inst); err != nil {
		conn.Close()
		log.Fatal("error writing instruction: ", err)
	}

	status := 0
	dec := bencode.NewDecoder(conn)
	for {
		resp := Response{}
		if err := dec.Decode(&resp); err != nil {
			conn.Close()
			log.Fatal("error decoding response: ", err)
		}
		if resp.Ex != "" {
			fmt.Fprint(os.Stderr, resp.Ex)
		}
		if resp.Err != "" {
			fmt.Fprint(os.Stderr, resp.Err)
		}
		if resp.Out != "" {
			fmt.Print(resp.Out)
		}
		if resp.Value != "" {
			fmt.Print(resp.Value)
		}
		if len(resp.Status) > 0 {
			if resp.Status[0] == "done" {
				break
			} else if resp.Status[0] == "eval-error" {
				status = 1
			}
		}
	}
	os.Exit(status)
}
