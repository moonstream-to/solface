package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

// The current version of solface.
var VERSION string = "0.1.2"

// Implements the solface CLI.
func main() {
	var interfaceName, license, pragma string
	var addAnnotations bool
	flag.StringVar(&interfaceName, "name", "", "Name for Solidity interface you would like to generate")
	flag.BoolVar(&addAnnotations, "annotations", false, "If present, adds annotations to generated interface. Annotations include: interface ID, method selectors, event signatures.")
	flag.StringVar(&license, "license", "", "License to include in generated interface - adds a comment at the top of the output with this as the SPDX identifier")
	flag.StringVar(&pragma, "pragma", "", "Solidity pragma to include in generated interface - adds this parameter as the pragma constraint at the top of the output")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s -name <interface name> [-annotations] {<path to ABI file> | stdin}\n\n", os.Args[0])
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

	annotations, annotationErr := Annotate(abi)
	if annotationErr != nil && addAnnotations {
		log.Fatalf("Error generating annotations: %s", annotationErr.Error())
	}

	generateErr := GenerateInterface(interfaceName, license, pragma, abi, annotations, addAnnotations, os.Stdout)
	if generateErr != nil {
		log.Fatalf("Error generating interface (%s): %s", interfaceName, generateErr.Error())
	}
}
