package main

import (
	"os"
	"reflect"
	"testing"
)

func TestFindCompoundTypesOnDiamondCutFacetABI(t *testing.T) {
	contents, readErr := os.ReadFile("fixtures/abis/DiamondCutFacet.json")
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
	contents, readErr := os.ReadFile("fixtures/abis/DiamondCutFacet.json")
	if readErr != nil {
		t.Fatal("Could not read file containing ABI")
	}

	abi, decodeErr := Decode(contents)
	if decodeErr != nil {
		t.Fatalf("Error decoding ABI: %s", decodeErr.Error())
	}

	enrichedABI := ResolveCompounds(abi)

	if len(enrichedABI.CompoundTypes) != 2 {
		t.Fatalf("Expected 2 compound types. Actual: %d", len(enrichedABI.CompoundTypes))
	}
}
