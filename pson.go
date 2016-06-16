// Program pson reads a text-format protobuf message from stdin or the named
// file, and converts it to a JSON message on stdout.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"bitbucket.org/creachadair/pson/textpb"
)

var (
	linePrefix = flag.String("prefix", "", "Message prefix (enables indentation)")
	indent     = flag.String("indent", "", "Indentation marker (enables indentation)")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage: pson <file>...

Reads the contents of each named file (or stdin if none are named) as a
text-format [1] protobuf message, converts each message to JSON, and catenates
the resulting JSON values to stdout.

This is intended to bridge between tools that know how to emit text-format
protobuf messages, but not JSON. You can use jq [2] to manipulate JSON messages
with ease, but there is no analogue of this for text-format protobufs.

The translation done by this tool is purely lexical; it does not know the
schema of the underlying protobuf messages.

[1] https://developers.google.com/protocol-buffers/docs/reference/cpp/google.protobuf.text_format
[2] https://stedolan.github.io/jq/

Options:`)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	paths := flag.Args()
	if len(paths) == 0 {
		paths = append(paths, "-")
	}

	for _, path := range paths {
		path, in := mustOpen(path)
		msg, err := textpb.Parse(in)
		if err != nil {
			log.Fatalf("Parsing %q failed: %v", path, err)
		}
		in.Close()
		bits, err := marshal(msg.Combine())
		if err != nil {
			log.Fatalf("Marshaling %q to JSON failed: %v", path, err)
		}
		fmt.Println(string(bits))
	}
}

func mustOpen(path string) (string, io.ReadCloser) {
	if path == "-" {
		return "stdin", os.Stdin
	}
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Open failed: %v", err)
	}
	return path, f
}

func marshal(msg textpb.Message) ([]byte, error) {
	if *linePrefix != "" || *indent != "" {
		return json.MarshalIndent(msg, *linePrefix, *indent)
	}
	return json.Marshal(msg)
}
