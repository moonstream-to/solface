package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// Represents a type declaration in an ABI.
type TypeDeclaration struct {
	Type string
}

// Represents a value in an ABI.
type Value struct {
	Name         string
	Type         string
	InternalType string `json:"internalType,omitempty"`
	Components   []Value
}

// Represents a parameter for an event in an ABI.
type EventArgument struct {
	Value
	Indexed bool
}

// Represents a smart contract method in an ABI.
type FunctionItem struct {
	Type            string
	Name            string  `json:"name,omitempty"`
	Inputs          []Value `json:"inputs,omitempty"`
	Outputs         []Value `json:"outputs,omitempty"`
	StateMutability string  `json:"stateMutability,omitempty"`
}

// Represents a log event in an ABI.
type EventItem struct {
	Type      string
	Name      string `json:"name"`
	Inputs    []EventArgument
	Anonymous bool
}

// Represents an exception/error in an ABI.
type ErrorItem struct {
	Type   string
	Name   string
	Inputs []Value
}

// Represents a parsed ABI, usable in the rest of solface.
type DecodedABI struct {
	Events    []EventItem
	Functions []FunctionItem
	Errors    []ErrorItem
}

// Represents annotations for an ABI.
type Annotations struct {
	InterfaceID       []byte
	FunctionSelectors [][]byte
}

// Decodes an ABI from its JSON representation (presented as a byte array).
//
// ABIs are decoded according to the Solidity Contract ABI specification:
// https://docs.soliditylang.org/en/v0.8.17/abi-spec.html
//
// This decoder uses the specification as of Solidity v0.8.17.

func Decode(rawJSON []byte) (DecodedABI, error) {
	var typeDeclarations []TypeDeclaration
	var rawMessages []json.RawMessage
	var decodedABI DecodedABI

	typesDecodeErr := json.Unmarshal(rawJSON, &typeDeclarations)
	if typesDecodeErr != nil {
		return decodedABI, typesDecodeErr
	}

	rawMessagesErr := json.Unmarshal(rawJSON, &rawMessages)
	if rawMessagesErr != nil {
		return decodedABI, rawMessagesErr
	}

	var numEvents, numFunctions, numErrors int
	for _, item := range typeDeclarations {
		if item.Type == "event" {
			numEvents++
		} else if item.Type == "function" {
			numFunctions++
		} else if item.Type == "error" {
			numErrors++
		}
	}
	if numEvents > 0 {
		decodedABI.Events = make([]EventItem, numEvents)
	}
	if numFunctions > 0 {
		decodedABI.Functions = make([]FunctionItem, numFunctions)
	}
	if numErrors > 0 {
		decodedABI.Errors = make([]ErrorItem, numErrors)
	}

	var currentEvent, currentFunction, currentError int
	for i, declaration := range typeDeclarations {
		if declaration.Type == "event" {
			var eventItem EventItem
			decodeEventErr := json.Unmarshal(rawMessages[i], &eventItem)
			if decodeEventErr != nil {
				return decodedABI, decodeEventErr
			}
			decodedABI.Events[currentEvent] = eventItem
			currentEvent++
		} else if declaration.Type == "function" {
			var functionItem FunctionItem
			decodeFunctionErr := json.Unmarshal(rawMessages[i], &functionItem)
			if decodeFunctionErr != nil {
				return decodedABI, decodeFunctionErr
			}
			decodedABI.Functions[currentFunction] = functionItem
			currentFunction++
		} else if declaration.Type == "error" {
			var errorItem ErrorItem
			decodeErrorErr := json.Unmarshal(rawMessages[i], &errorItem)
			if decodeErrorErr != nil {
				return decodedABI, decodeErrorErr
			}
			decodedABI.Errors[currentError] = errorItem
			currentError++
		}
	}

	return decodedABI, nil
}

// Calculates the 4-byte method selector for a given ABI function.
func MethodSelector(function FunctionItem) []byte {
	argumentTypes := make([]string, len(function.Inputs))
	for i, input := range function.Inputs {
		argumentTypes[i] = input.Type
	}
	argumentTypesString := strings.Join(argumentTypes, ",")
	signature := fmt.Sprintf("%s(%s)", function.Name, argumentTypesString)
	return crypto.Keccak256([]byte(signature))[:4]
}

// Generates annotations for a decoded ABI.
func Annotate(decodedABI DecodedABI) (Annotations, error) {
	var annotations Annotations
	annotations.InterfaceID = []byte{0x0, 0x0, 0x0, 0x0}
	annotations.FunctionSelectors = make([][]byte, len(decodedABI.Functions))
	for i, functionItem := range decodedABI.Functions {
		selector := MethodSelector(functionItem)
		annotations.FunctionSelectors[i] = selector

		// XOR into InterfaceID byte by byte
		annotations.InterfaceID[0] ^= selector[0]
		annotations.InterfaceID[1] ^= selector[1]
		annotations.InterfaceID[2] ^= selector[2]
		annotations.InterfaceID[3] ^= selector[3]
	}
	return annotations, nil
}

// Returns true if the given value is a compound type (i.e. composed of other types like a struct or array)
// and false otherwise.
func (v Value) IsCompoundType() bool {
	return len(v.Components) > 0
}
