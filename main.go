package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

var VERSION string = "0.0.4"

func main() {
	var interfaceName string
	flag.StringVar(&interfaceName, "name", "", "Name for Solidity interface you would like to generate")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s -name <interface name> {<path to ABI file> | stdin}\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nsolface version %s\n", VERSION)
	}

	flag.Parse()

	if interfaceName == "" {
		flag.Usage()
		os.Exit(1)
	}

	var contents []byte
	var readErr error

	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(1)
	} else if flag.NArg() == 1 {
		infile := flag.Arg(0)
		contents, readErr = os.ReadFile(infile)
	} else {
		contents, readErr = io.ReadAll(os.Stdin)
	}
	if readErr != nil {
		log.Fatalf("Error reading ABI: %s", readErr.Error())
	}

	abi, decodeErr := Decode(contents)
	if decodeErr != nil {
		log.Fatalf("Error decoding ABI: %s", decodeErr.Error())
	}

	generateErr := GenerateInterface(interfaceName, abi, os.Stdout)
	if generateErr != nil {
		log.Fatalf("Error generating interface (%s): %s", interfaceName, generateErr.Error())
	}
}
