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
	"bitbucket.org/creachadair/pson/textpb/format"
)

var (
	linePrefix = flag.String("prefix", "", "Message prefix (enables indentation)")
	indent     = flag.String("indent", "", "Indentation marker (enables indentation)")
	doSplit    = flag.Bool("split", false, "Split into single-valued messages")
	doRecur    = flag.Bool("rsplit", false, "Split recursively (implies -split)")
	doCamel    = flag.Bool("camel", false, "Convert names to camel-case")
	doProto    = flag.Bool("proto", false, "Render output as text-format protobuf")
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

		// If requested, split the message into single-valued messages;
		// otherwise combine the (single) top-level message.
		write := writeMessages
		if *doProto {
			write = writeProtos
		}
		if *doRecur {
			err = write(os.Stdout, msg.Split()...)
		} else if *doSplit {
			err = write(os.Stdout, msg.Split1()...)
		} else {
			err = write(os.Stdout, msg.Combine())
		}
		if err != nil {
			log.Fatalf("Error writing JSON output: %v", err)
		}
	}
}

func writeMessages(w io.Writer, msgs ...textpb.Message) error {
	enc := json.NewEncoder(w)
	enc.SetIndent(*linePrefix, *indent)
	for _, out := range msgs {
		if *doCamel {
			out.ToCamel()
		}
		if err := enc.Encode(out); err != nil {
			return err
		}
	}
	return nil
}

func writeProtos(w io.Writer, msgs ...textpb.Message) error {
	cfg := format.Config{
		Compact: *indent == "",
		Indent:  *indent,
	}
	for _, out := range msgs {
		if err := cfg.Text(w, out); err != nil {
			return err
		}
		fmt.Fprintln(w)
	}
	return nil
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
