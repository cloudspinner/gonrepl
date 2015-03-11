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
	Value  string
	Status []string
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: gonrepl [-p port]\n")
	os.Exit(2)
}

var addr = flag.String("a", "localhost:"+os.Getenv("LEIN_REPL_PORT"), "nREPL port")

func main() {
	flag.Usage = usage
	flag.Parse()

	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("error reading standard input", err)
	}
	code := string(bytes)
	inst := map[string]interface{}{
		"op":   "eval",
		"code": code,
	}

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatal("error connecting to "+*addr, err)
	}
	defer conn.Close()

	enc := bencode.NewEncoder(conn)
	if err := enc.Encode(inst); err != nil {
		conn.Close()
		log.Fatal("error writing instruction", err)
	}

	dec := bencode.NewDecoder(conn)
	for {
		resp := Response{}
		if err := dec.Decode(&resp); err != nil {
			conn.Close()
			log.Fatal("error decoding response", err)
		}
		if resp.Ex != "" {
			fmt.Println(resp.Ex)
		}
		if resp.Value != "" {
			fmt.Println(resp.Value)
		}
		if len(resp.Status) > 0 && resp.Status[0] == "done" {
			break
		}
	}
}
