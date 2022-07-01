// An nREPL client that reads Clojure code from standard input and
// writes the result to standard output.  Connects to
// localhost:$LEIN_REPL_PORT by default. Pass the -a flag to override
// the default address. If LEIN_REPL_PORT is undefined and the default
// address not overriden with the -a flag, falls back to .nrepl-port in
// the user home directory.
//
// Exceptions and captured stderr + stdout go to standard error.
// Value of the evaluated expression go to standard output.
// Return with exit code 1 if there was an evaluation error.
// Return with exit code 2 if an error prevented the sending of
// a message or processing of the respons.
//
// TODO:
// - automatically enclose s-expr with parenthesis if it isn't already
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudspinner/gonrepl/client"
)
const (
	portEnv     = "LEIN_REPL_PORT"
	defaultHost = "localhost:"
	portFile    = ".nrepl-port"
)

type options struct {
	addr  *string
	sid   *string
	op    *string
	clone *bool
	close *bool
}

func readPortFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error reading home directory: %v", err)
	}
	bytes, err := os.ReadFile(home + "/" + portFile)
	if err != nil {
		return "", fmt.Errorf("error reading %s: %v", portFile, err)
	}
	return string(bytes), nil
}

func addr(opts options) (string, error) {
	addr := *opts.addr
	if addr == defaultHost {
		port, err := readPortFile()
		if err != nil {
			return "", fmt.Errorf("cannot read port from %s in home directory: %v", portFile, err)
		}
		return addr + port, nil
	}
	return addr, nil
}

func newMessage(opts options) (client.Message, error) {
	operation := *opts.op
	switch operation {
	case "clone", "close", "eval", "format-code":
		// do nothing
	default:
		return client.Message{}, fmt.Errorf("unsupported operation: %v\n", operation)
	}

	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return client.Message{}, fmt.Errorf("error reading standard input: %v", err)
	}
	code := string(bytes)
	msg := client.Message{}
	msg.Inst = map[string]interface{}{
		"op":   operation,
		"code": code,
	}

	if *opts.clone {
		msg.Inst = map[string]interface{}{
			"op": "clone",
		}
	} else if *opts.close {
		msg.Inst = map[string]interface{}{
			"op": "close",
		}
	}

	if *opts.sid != "" {
		msg.Inst["session"] = *opts.sid
	}
	return msg, nil
}

func handleResp(resp client.Response) {
	if resp.Ex != "" {
		fmt.Fprint(os.Stderr, resp.Ex)
	}
	if resp.Err != "" {
		fmt.Fprint(os.Stderr, resp.Err)
	}
	if resp.Out != "" {
		fmt.Fprint(os.Stderr, resp.Out)
	}
	if resp.Value != "" {
		fmt.Println(resp.Value)
	}
	if resp.NewSession != "" {
		fmt.Println(resp.NewSession)
	}
	if resp.FormattedCode != "" {
		fmt.Println(resp.FormattedCode)
	}
}

func process(opts options) error {
	flag.Parse()
	addr, err := addr(opts)
	if err != nil {
		return fmt.Errorf("cannot find server address: %v", err)
	}
	c, err := client.NewClient(addr)
	if err != nil {
		return fmt.Errorf("cannot connect to server at %s: %v", addr, err)
	}
	defer c.Close()
	msg, err := newMessage(opts)
	if err != nil {
		return fmt.Errorf("error creating message: ", err)
	}
	err = c.Send(msg, handleResp)
	if err == nil {
		return nil
	}
	if e, ok := err.(client.EvalErr); ok {
		return e
	}
	return fmt.Errorf("error sending message: ", err)
}

func main() {
	err := process(options{
		addr: flag.String("a", defaultHost+os.Getenv(portEnv), "nREPL port"),
		sid:  flag.String("s", "", "session id"),
		op: flag.String("o", "eval", `operation, possible values are:
	clone
	close
	eval
	format-code`),
		clone: flag.Bool("clone", false, "clone session"),
		close: flag.Bool("close", false, "close session"),
	})
	if err == nil {
		os.Exit(0)
	}
	if _, ok := err.(client.EvalErr); ok {
		os.Exit(1)
	}
		fmt.Fprint(os.Stderr, err)
		os.Exit(2)
	}
}
