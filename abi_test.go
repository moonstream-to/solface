package main

import (
	"os"
	"testing"
)

func TestDecodeOwnableERC20(t *testing.T) {
	contents, readErr := os.ReadFile("fixtures/abis/OwnableERC20.json")
	if readErr != nil {
		t.Fatal("Could not read file containing ABI")
	}

	decodedABI, decodeErr := Decode(contents)
	if decodeErr != nil {
		t.Fatalf("Could not decode ABI: %s", decodeErr.Error())
	}

	expectedNumEvents := 3
	actualNumEvents := len(decodedABI.Events)
	if actualNumEvents != expectedNumEvents {
		t.Fatalf("Failure decoding events from ABI. Expected number of events: %d, actual number of events: %d", expectedNumEvents, actualNumEvents)
	}

	expectedNumFunctions := 15
	actualNumFunctions := len(decodedABI.Functions)
	if actualNumFunctions != expectedNumFunctions {
		t.Fatalf("Failure decoding functions from ABI. Expected number of functions: %d, actual number of functions: %d", expectedNumFunctions, actualNumFunctions)
	}

	expectedNumErrors := 0
	actualNumErrors := len(decodedABI.Errors)
	if actualNumErrors != expectedNumErrors {
		t.Fatalf("Failure decoding errors from ABI. Expected number of errors: %d, actual number of errors: %d", expectedNumErrors, actualNumErrors)
	}
}

func TestSingleEvent(t *testing.T) {
	var events = []byte(`[{
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "owner",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "spender",
        "type": "address"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "value",
        "type": "uint256"
      }
    ],
    "name": "Approval",
    "type": "event"
  }]`)

	decodedABI, err := Decode(events)
	if err != nil {
		t.Fatalf("Could not decode ABI: %s", err.Error())
	}

	expectedNumEvents := 1
	actualNumEvents := len(decodedABI.Events)
	if actualNumEvents != expectedNumEvents {
		t.Fatalf("Failure decoding events from ABI. Expected number of events: %d, actual number of events: %d", expectedNumEvents, actualNumEvents)
	}

	expectedNumFunctions := 0
	actualNumFunctions := len(decodedABI.Functions)
	if actualNumFunctions != expectedNumFunctions {
		t.Fatalf("Failure decoding functions from ABI. Expected number of functions: %d, actual number of functions: %d", expectedNumFunctions, actualNumFunctions)
	}

	expectedNumErrors := 0
	actualNumErrors := len(decodedABI.Errors)
	if actualNumErrors != expectedNumErrors {
		t.Fatalf("Failure decoding errors from ABI. Expected number of errors: %d, actual number of errors: %d", expectedNumErrors, actualNumErrors)
	}

	actualEvent := decodedABI.Events[0]
	if actualEvent.Anonymous != false {
		t.Fatal("Expected event *not* to be anonymous. It was.")
	}
	expectedName := "Approval"
	if actualEvent.Name != expectedName {
		t.Fatalf("Expected event name: %s. Actual name: %s", expectedName, actualEvent.Name)
	}
	expectedType := "event"
	if actualEvent.Type != expectedType {
		t.Fatalf("Expected event type: %s. Actual type: %s", expectedType, actualEvent.Type)
	}

	expectedInputNames := []string{"owner", "spender", "value"}
	expectedInputTypes := []string{"address", "address", "uint256"}
	expectedInputIndexed := []bool{true, true, false}

	expectedInputsLength := 3
	actualInputsLength := len(actualEvent.Inputs)
	if actualInputsLength != expectedInputsLength {
		t.Fatalf("Expected length of inputs: %d. Actual length of inputs: %d", expectedInputsLength, actualInputsLength)
	}

	for i, inputItem := range actualEvent.Inputs {
		if inputItem.Name != expectedInputNames[i] {
			t.Fatalf("Input item %d: Expected name: %s. Actual name: %s", i, expectedInputNames[i], inputItem.Name)
		}
		if inputItem.Type != expectedInputTypes[i] {
			t.Fatalf("Input item %d: Expected type : %s. Actual type: %s", i, expectedInputTypes[i], inputItem.Type)
		}
		if inputItem.Indexed != expectedInputIndexed[i] {
			t.Fatalf("Input item %d: Expected indexed: %t. Actual indexed: %t", i, expectedInputIndexed[i], inputItem.Indexed)
		}
	}
}

func TestSingleFunction(t *testing.T) {
	var functions = []byte(`[{
    "inputs": [
      {
        "internalType": "address",
        "name": "spender",
        "type": "address"
      },
      {
        "internalType": "uint256",
        "name": "amount",
        "type": "uint256"
      }
    ],
    "name": "approve",
    "outputs": [
      {
        "internalType": "bool",
        "name": "",
        "type": "bool"
      }
    ],
    "stateMutability": "nonpayable",
    "type": "function"
  }]`)

	decodedABI, err := Decode(functions)
	if err != nil {
		t.Fatalf("Could not decode ABI: %s", err.Error())
	}

	expectedNumEvents := 0
	actualNumEvents := len(decodedABI.Events)
	if actualNumEvents != expectedNumEvents {
		t.Fatalf("Failure decoding events from ABI. Expected number of events: %d, actual number of events: %d", expectedNumEvents, actualNumEvents)
	}

	expectedNumFunctions := 1
	actualNumFunctions := len(decodedABI.Functions)
	if actualNumFunctions != expectedNumFunctions {
		t.Fatalf("Failure decoding functions from ABI. Expected number of functions: %d, actual number of functions: %d", expectedNumFunctions, actualNumFunctions)
	}

	expectedNumErrors := 0
	actualNumErrors := len(decodedABI.Errors)
	if actualNumErrors != expectedNumErrors {
		t.Fatalf("Failure decoding errors from ABI. Expected number of errors: %d, actual number of errors: %d", expectedNumErrors, actualNumErrors)
	}

	actualFunction := decodedABI.Functions[0]
	expectedName := "approve"
	if actualFunction.Name != expectedName {
		t.Fatalf("Expected function name: %s. Actual name: %s", expectedName, actualFunction.Name)
	}
	expectedType := "function"
	if actualFunction.Type != expectedType {
		t.Fatalf("Expected function type: %s. Actual type: %s", expectedType, actualFunction.Type)
	}

	expectedInputNames := []string{"spender", "amount"}
	expectedInputTypes := []string{"address", "uint256"}
	expectedInputsLength := 2
	actualInputsLength := len(actualFunction.Inputs)
	if actualInputsLength != expectedInputsLength {
		t.Fatalf("Expected length of inputs: %d. Actual length of inputs: %d", expectedInputsLength, actualInputsLength)
	}

	for i, inputItem := range actualFunction.Inputs {
		if inputItem.Name != expectedInputNames[i] {
			t.Fatalf("Input item %d: Expected name: %s. Actual name: %s", i, expectedInputNames[i], inputItem.Name)
		}
		if inputItem.Type != expectedInputTypes[i] {
			t.Fatalf("Input item %d: Expected type : %s. Actual type: %s", i, expectedInputTypes[i], inputItem.Type)
		}
	}

	expectedOutputNames := []string{""}
	expectedOutputTypes := []string{"bool"}
	expectedOutputsLength := 1
	actualOutputsLength := len(actualFunction.Outputs)
	if actualOutputsLength != expectedOutputsLength {
		t.Fatalf("Expected length of outputs: %d. Actual length of outputs: %d", expectedOutputsLength, actualOutputsLength)
	}

	for i, outputItem := range actualFunction.Outputs {
		if outputItem.Name != expectedOutputNames[i] {
			t.Fatalf("Output item %d: Expected name: %s. Actual name: %s", i, expectedOutputNames[i], outputItem.Name)
		}
		if outputItem.Type != expectedOutputTypes[i] {
			t.Fatalf("Output item %d: Expected type : %s. Actual type: %s", i, expectedOutputTypes[i], outputItem.Type)
		}
	}
}
