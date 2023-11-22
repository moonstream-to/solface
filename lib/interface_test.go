package lib

import (
	"io"
	"os"
	"reflect"
	"testing"
)

func TestFindCompoundTypesOnDiamondCutFacetABI(t *testing.T) {
	contents, readErr := os.ReadFile("../fixtures/abis/DiamondCutFacet.json")
	if readErr != nil {
		t.Fatal("Could not read file containing ABI")
	}

	abi, decodeErr := Decode(contents)
	if decodeErr != nil {
		t.Fatalf("Error decoding ABI: %s", decodeErr.Error())
	}

	eventInputs, functionInputs, functionOutputs, errorInputs := FindCompoundTypes(abi)

	expectedEventInputs := []ItemValueIndex{{0, 0}}
	expectedFunctionInputs := []ItemValueIndex{{0, 0}}
	expectedFunctionOutputs := []ItemValueIndex{}
	expectedErrorInputs := []ItemValueIndex{}

	if !reflect.DeepEqual(eventInputs, expectedEventInputs) {
		t.Fatal("Actual indices of compound event inputs did not match expectation")
	}
	if !reflect.DeepEqual(functionInputs, expectedFunctionInputs) {
		t.Fatal("Actual indices of compound function inputs did not match expectation")
	}
	if !reflect.DeepEqual(functionOutputs, expectedFunctionOutputs) {
		t.Fatal("Actual indices of compound function outputs did not match expectation")
	}
	if !reflect.DeepEqual(errorInputs, expectedErrorInputs) {
		t.Fatal("Actual indices of compound error inputs did not match expectation")
	}
}

func TestCompoundSingleValue(t *testing.T) {
	originalValue := Value{Name: "lol", Type: "tuple()", Components: []Value{
		{Name: "rofl", Type: "uint256", Components: []Value{}},
		{Name: "omg", Type: "address", Components: []Value{}},
	}}

	var typeCounter, nameCounter int

	compound, newTypes := CompoundSingleValue(originalValue, &typeCounter, &nameCounter)

	if compound.Name != originalValue.Name {
		t.Fatalf("Incorrect name for compound type. Expected: %s, actual: %s", originalValue.Name, compound.Name)
	}

	if len(newTypes) != 1 {
		t.Fatalf("Expected 1 new types. Got: %d", len(newTypes))
	}

	actualNewType := newTypes[0]
	if len(actualNewType.Members) != 2 {
		t.Fatalf("Expected 2 members in new type. Got: %d", len(actualNewType.Members))
	}
}

func TestCompoundSingleValueDeep(t *testing.T) {
	originalValue := Value{Name: "lol", Type: "tuple()", Components: []Value{
		{Name: "rofl", Type: "uint256", Components: []Value{}},
		{Name: "omg", Type: "tuple()", Components: []Value{
			{Name: "wtf", Type: "address", Components: []Value{}},
			{Name: "bbq", Type: "uint256", Components: []Value{}},
		}},
	}}

	var typeCounter, nameCounter int

	compound, newTypes := CompoundSingleValue(originalValue, &typeCounter, &nameCounter)

	if compound.Name != originalValue.Name {
		t.Fatalf("Incorrect name for compound type. Expected: %s, actual: %s", originalValue.Name, compound.Name)
	}

	if len(newTypes) != 2 {
		t.Fatalf("Expected 2 new types. Got: %d", len(newTypes))
	}

	typeIndex := map[string]*CompoundType{}
	for _, newType := range newTypes {
		typeIndex[newType.TypeName] = &newType
	}
}

func TestResolveCompoundsDiamondCutFacet(t *testing.T) {
	contents, readErr := os.ReadFile("../fixtures/abis/DiamondCutFacet.json")
	if readErr != nil {
		t.Fatal("Could not read file containing ABI")
	}

	abi, decodeErr := Decode(contents)
	if decodeErr != nil {
		t.Fatalf("Error decoding ABI: %s", decodeErr.Error())
	}

	oldEventInputs, oldFunctionInputs, oldFunctionOutputs, oldErrorInputs := FindCompoundTypes(abi)
	if len(oldEventInputs) != 1 {
		t.Fatalf("Expected 1 compound event inputs. Actual: %d", len(oldEventInputs))
	}
	if len(oldFunctionInputs) != 1 {
		t.Fatalf("Expected 1 compound oldFunction inputs. Actual: %d", len(oldFunctionInputs))
	}
	if len(oldFunctionOutputs) != 0 {
		t.Fatalf("Expected 0 compound oldFunction outputs. Actual: %d", len(oldFunctionOutputs))
	}
	if len(oldErrorInputs) != 0 {
		t.Fatalf("Expected 0 compound oldError inputs. Actual: %d", len(oldErrorInputs))
	}

	enrichedABI := ResolveCompounds(abi)

	if len(enrichedABI.CompoundTypes) != 2 {
		t.Fatalf("Expected 2 compound types. Actual: %d", len(enrichedABI.CompoundTypes))
	}

	eventInputs, functionInputs, functionOutputs, errorInputs := FindCompoundTypes(enrichedABI.EnrichedABI)
	if len(eventInputs) != 0 {
		t.Fatalf("Expected 0 compound event inputs. Actual: %d", len(eventInputs))
	}
	if len(functionInputs) != 0 {
		t.Fatalf("Expected 0 compound function inputs. Actual: %d", len(functionInputs))
	}
	if len(functionOutputs) != 0 {
		t.Fatalf("Expected 0 compound function outputs. Actual: %d", len(functionOutputs))
	}
	if len(errorInputs) != 0 {
		t.Fatalf("Expected 0 compound error inputs. Actual: %d", len(errorInputs))
	}
}

func TestGenerateInterfaceDiamondCutFacet(t *testing.T) {
	contents, readErr := os.ReadFile("../fixtures/abis/DiamondCutFacet.json")
	if readErr != nil {
		t.Fatal("Could not read file containing ABI")
	}

	abi, decodeErr := Decode(contents)
	if decodeErr != nil {
		t.Fatalf("Error decoding ABI: %s", decodeErr.Error())
	}

	var annotations Annotations
	includeAnnotations := false

	// Replace io.Discard with os.Stdout to inspect output:
	// err := GenerateInterface("IDiamondCutFacet", "", "", abi, annotations, includeAnnotations, os.Stdout)
	err := GenerateInterface("IDiamondCutFacet", "", "", abi, annotations, includeAnnotations, io.Discard)

	if err != nil {
		t.Fatalf("Error generating interface: %s", err.Error())
	}
}

func TestGenerateInterfaceOwnableERC20(t *testing.T) {
	contents, readErr := os.ReadFile("../fixtures/abis/OwnableERC20.json")
	if readErr != nil {
		t.Fatal("Could not read file containing ABI")
	}

	abi, decodeErr := Decode(contents)
	if decodeErr != nil {
		t.Fatalf("Error decoding ABI: %s", decodeErr.Error())
	}

	var annotations Annotations
	includeAnnotations := false

	// Replace io.Discard with os.Stdout to inspect output:
	// err := GenerateInterface("IOwnableERC20", "Apache-2.0", "^8.20.0", abi, annotations, includeAnnotations, os.Stdout)
	err := GenerateInterface("IOwnableERC20", "Apache-2.0", "^8.20.0", abi, annotations, includeAnnotations, io.Discard)

	if err != nil {
		t.Fatalf("Error generating interface: %s", err.Error())
	}
}
