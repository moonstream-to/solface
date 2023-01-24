package main

import (
	"encoding/json"
)

/**
 * ABIs are decoded according to the Solidity Contract ABI specification:
 * https://docs.soliditylang.org/en/v0.8.17/abi-spec.html
 *
 * This decoder uses the specification as of Solidity v0.8.17.
 */

type TypeDeclaration struct {
	Type string
}

type Value struct {
	Name       string
	Type       string
	Components []Value
}

type EventArgument struct {
	Value
	Indexed bool
}

type FunctionItem struct {
	Type            string
	Name            string  `json:"name,omitempty"`
	Inputs          []Value `json:"inputs,omitempty"`
	Outputs         []Value `json:"outputs,omitempty"`
	StateMutability string  `json:"stateMutability,omitempty"`
}

type EventItem struct {
	Type      string
	Name      string `json:"name"`
	Inputs    []EventArgument
	Anonymous bool
}

type ErrorItem struct {
	Type   string
	Name   string
	Inputs []Value
}

type DecodedABI struct {
	Events    []EventItem
	Functions []FunctionItem
	Errors    []ErrorItem
}

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

func (v Value) IsCompoundType() bool {
	return len(v.Components) > 0
}
