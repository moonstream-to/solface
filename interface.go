package main

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

// Represents an ordered pair of array indices, the first index representing a position in the ABI array,
// and the second index representing a position in the parameters array for that ABI item.
type ItemValueIndex struct {
	ItemIndex  int
	ValueIndex int
}

// Represents a named parameter in an ABI item.
type NamedValue struct {
	Name  string
	Value Value
}

// Represents a compound type.
type CompoundType struct {
	TypeName string
	Members  []NamedValue
}

// Represents a decoded ABI along with the compound types that need to be defined in a Solidity interface
// to a contract exposing that ABI.
type DecodedABIWithCompundTypes struct {
	OriginalABI   DecodedABI
	CompoundTypes []CompoundType
	EnrichedABI   DecodedABI
}

// InterfaceSpecification specifies certain details about the Solidity interface that should be generated.
//  1. Name: The name of the Solidity interface.
//  2. ABI: The ABI that the interface is being generated for.
//  3. Annotations: A list of annotations (interface ID, method selectors) for the interface.
//  4. IncludeAnnotations: Whether or not to include the annotations in the generated interface.
//  5. CompoundTypes: The compound types that need to be defined in the interface.
//  6. SolfaceVersion: The version of solface that generated the interface.
//  7. License: The SPDX license identifier to be generated at the top of the output - if empty, this
//     will not be included.
//  8. Pragma: The Solidity pragma to be generated at the top of the output - if empty, this will not
//     be included.
type InterfaceSpecification struct {
	Name               string
	ABI                DecodedABI
	Annotations        Annotations
	IncludeAnnotations bool
	CompoundTypes      []CompoundType
	SolfaceVersion     string
	License            string
	Pragma             string
}

// Generates a fresh name for an anonymous attribute.
func GenerateName(nameCounter *int) string {
	result := fmt.Sprintf("Attribute%d", *nameCounter)
	(*nameCounter) += 1
	return result
}

// Parses the name of an internal type and either returns that name (for structs) or "Compound" (for
// any other type).
// For nested structs (e.g. structs defined in other contracts or interfaces), this only returns the
// final component of the name.
func ParseInternalType(internalType string) string {
	if !strings.HasPrefix(internalType, "struct") {
		return "Compound"
	}

	structQualifiedName := strings.TrimPrefix(internalType, "struct ")
	structNameComponents := strings.Split(structQualifiedName, ".")
	structName := structNameComponents[len(structNameComponents)-1]
	return structName
}

// Generates a fresh name for an anonymous compound type.
func GenerateType(typeCounter *int, internalType string) string {
	typeName := ParseInternalType(internalType)
	result := fmt.Sprintf("%s%d", typeName, *typeCounter)
	(*typeCounter) += 1
	return result
}

// This function returns true if the given Solidity type requires a location modifier ("memory", "storage", "calldata")
// when used as a function parameter or return value.
func SolidityTypeRequiresLocation(solidityType string) bool {
	if strings.HasSuffix(solidityType, "[]") {
		return true
	} else if solidityType == "string" {
		return true
	} else if solidityType == "bytes" {
		return true
	} else if solidityType == "bool" {
		return false
	} else if strings.HasPrefix(solidityType, "uint") {
		return false
	} else if solidityType == "address" {
		return false
	} else if strings.HasPrefix(solidityType, "bytes") {
		// It is not exactly "bytes" because that was handled above. "bytes[]" also handled above.
		// This covers bytes32, etc.
		return false
	}

	return true
}

// Finds all the compound types that need to be defined in order to interface with a contract with the
// given decoded ABI.
//
// Return values specify which items in the following arrays are compound types (by index):
// 1. Event inputs
// 2. Function inputs
// 3. Function outputs
// 4. Error inputs
func FindCompoundTypes(abi DecodedABI) ([]ItemValueIndex, []ItemValueIndex, []ItemValueIndex, []ItemValueIndex) {
	var eventInputs, functionInputs, functionOutputs, errorInputs []ItemValueIndex

	eventInputs = make([]ItemValueIndex, 0)
	for i, eventItem := range abi.Events {
		for j, input := range eventItem.Inputs {
			if input.IsCompoundType() {
				eventInputs = append(eventInputs, ItemValueIndex{i, j})
			}
		}
	}

	functionInputs = make([]ItemValueIndex, 0)
	functionOutputs = make([]ItemValueIndex, 0)
	for i, functionItem := range abi.Functions {
		for j, input := range functionItem.Inputs {
			if input.IsCompoundType() {
				functionInputs = append(functionInputs, ItemValueIndex{i, j})
			}
		}

		for k, output := range functionItem.Outputs {
			if output.IsCompoundType() {
				functionOutputs = append(functionOutputs, ItemValueIndex{i, k})
			}
		}
	}

	errorInputs = make([]ItemValueIndex, 0)
	for i, errorItem := range abi.Errors {
		for j, input := range errorItem.Inputs {
			if input.IsCompoundType() {
				errorInputs = append(errorInputs, ItemValueIndex{i, j})
			}
		}
	}

	return eventInputs, functionInputs, functionOutputs, errorInputs
}

// Recursively creates the compound types required to represent the given value.
// If the top-level compound type has members which are also compound types themselves, they will be
// included in the second return value.
// The first return value is a transformation of the original value represented using the new
// compound types.
func CompoundSingleValue(val Value, typeCounter, nameCounter *int) (Value, []CompoundType) {
	newTypes := make([]CompoundType, 0)

	// base case of recursion
	if !val.IsCompoundType() {
		return val, newTypes
	}

	var result Value
	result.Name = val.Name

	updatedComponents := make([]Value, 0)
	for _, component := range val.Components {
		subvalue, subTypes := CompoundSingleValue(component, typeCounter, nameCounter)
		updatedComponents = append(updatedComponents, subvalue)
		if len(subTypes) > 0 {
			newTypes = append(newTypes, subTypes...)
		}
	}

	var compound CompoundType
	compound.TypeName = GenerateType(typeCounter, val.InternalType)
	compound.Members = make([]NamedValue, len(updatedComponents))
	for i, component := range updatedComponents {
		memberName := component.Name
		if memberName == "" && nameCounter != nil {
			memberName = GenerateName(nameCounter)
		}
		compound.Members[i] = NamedValue{memberName, component}
	}
	newTypes = append(newTypes, compound)

	result.Type = compound.TypeName
	if strings.HasSuffix(val.Type, "[]") {
		result.Type = fmt.Sprintf("%s[]", compound.TypeName)
	}

	return result, newTypes
}

// Transitively resolves all compound types comprising the parameters and return values of all items
// in the given decoded ABI.
func ResolveCompounds(abi DecodedABI) DecodedABIWithCompundTypes {
	var typeCounter, nameCounter int

	var result DecodedABIWithCompundTypes
	result.OriginalABI = abi
	result.EnrichedABI.Events = make([]EventItem, len(abi.Events))
	result.EnrichedABI.Functions = make([]FunctionItem, len(abi.Functions))
	result.EnrichedABI.Errors = make([]ErrorItem, len(abi.Errors))
	result.CompoundTypes = make([]CompoundType, 0)

	for j, eventItem := range abi.Events {
		newEventItem := EventItem{Type: eventItem.Type, Name: eventItem.Name, Anonymous: eventItem.Anonymous}
		newEventItem.Inputs = make([]EventArgument, len(eventItem.Inputs))
		for i, inputEventArgument := range eventItem.Inputs {
			newInputValue, newTypes := CompoundSingleValue(inputEventArgument.Value, &typeCounter, &nameCounter)
			newEventArgument := EventArgument{Indexed: inputEventArgument.Indexed, Value: newInputValue}
			newEventItem.Inputs[i] = newEventArgument
			result.CompoundTypes = append(result.CompoundTypes, newTypes...)
		}

		result.EnrichedABI.Events[j] = newEventItem
	}

	for j, functionItem := range abi.Functions {
		newFunctionItem := FunctionItem{Type: functionItem.Type, Name: functionItem.Name, StateMutability: functionItem.StateMutability}
		newFunctionItem.Inputs = make([]Value, len(functionItem.Inputs))
		newFunctionItem.Outputs = make([]Value, len(functionItem.Outputs))

		for i, value := range functionItem.Inputs {
			newValue, newTypes := CompoundSingleValue(value, &typeCounter, &nameCounter)
			newFunctionItem.Inputs[i] = newValue
			result.CompoundTypes = append(result.CompoundTypes, newTypes...)
		}

		for i, value := range functionItem.Outputs {
			newValue, newTypes := CompoundSingleValue(value, &typeCounter, nil)
			newFunctionItem.Outputs[i] = newValue
			result.CompoundTypes = append(result.CompoundTypes, newTypes...)
		}

		result.EnrichedABI.Functions[j] = newFunctionItem
	}

	for j, errorItem := range abi.Errors {
		newErrorItem := ErrorItem{Type: errorItem.Type, Name: errorItem.Name}
		newErrorItem.Inputs = make([]Value, len(errorItem.Inputs))
		for i, value := range errorItem.Inputs {
			newValue, newTypes := CompoundSingleValue(value, &typeCounter, &nameCounter)
			newErrorItem.Inputs[i] = newValue
			result.CompoundTypes = append(result.CompoundTypes, newTypes...)
		}

		result.EnrichedABI.Errors[j] = newErrorItem
	}

	return result
}

// This is the Go template used to generate Solidity interfaces to contracts with a given ABI.
// The template is meant to be applied to InterfaceSpecification structs.
const InterfaceTemplate string = `
{{ if .License }}
// SPDX-License-Identifier: {{.License}}
{{ end }}
{{- if .Pragma }}
pragma solidity {{.Pragma}};
{{ end }}
// Interface generated by solface: https://github.com/moonstream-to/solface
// solface version: {{.SolfaceVersion}}
{{- $includeAnnotations := .IncludeAnnotations}}
{{- $annotations := .Annotations}}
{{ if $includeAnnotations -}}
// Interface ID: {{printf "%x" .Annotations.InterfaceID}}
{{ end -}}
interface {{.Name}} {
	// structs
{{- range .CompoundTypes}}
	struct {{.TypeName}} {
	{{- range .Members}}
		{{.Value.Type}} {{.Name}};
	{{- end}}
	}
{{- end}}

	// events
{{- range .ABI.Events}}
	event {{.Name}}({{- range $i, $input := .Inputs}}{{if $i}}, {{end}}{{.Type}} {{.Name}}{{- end}});
{{- end}}

	// functions
{{- range $i, $function := .ABI.Functions}}
	{{if $includeAnnotations -}}
	// Selector: {{printf "%x" (index $annotations.FunctionSelectors $i)}}
	{{end -}}
	function {{.Name}}({{- range $i, $input := .Inputs}}{{if $i}}, {{end}}{{.Type}}{{if (needsMemory .Type)}} memory{{end}} {{.Name}} {{- end}}) external {{if (or (eq .StateMutability "view") (eq .StateMutability "pure"))}}{{.StateMutability}}{{end}}{{if .Outputs}} returns ({{- range $i, $output := .Outputs}}{{if $i}}, {{end}}{{.Type}}{{if (needsMemory .Type)}} memory{{end}}{{if .Name}} {{.Name}}{{end}}{{- end}}){{end}};
{{- end}}

	// errors
{{- range .ABI.Errors}}
	error {{.Name}}({{- range $i, $error := .Inputs}}{{if $i}}, {{end}}{{.Type}} {{.Name}}{{- end}});
{{- end}}
}
`

// Generates a Solidity interface for the given ABI (with the given parameters).
// The specification is generated by applying the specification to a Go template.
func GenerateInterface(interfaceName, license, pragma string, abi DecodedABI, annotations Annotations, includeAnnotations bool, writer io.Writer) error {
	resolved := ResolveCompounds(abi)
	spec := InterfaceSpecification{Name: interfaceName, ABI: resolved.EnrichedABI, Annotations: annotations, IncludeAnnotations: includeAnnotations, CompoundTypes: resolved.CompoundTypes, SolfaceVersion: VERSION, License: license, Pragma: pragma}

	templateFuncs := map[string]any{
		"needsMemory": SolidityTypeRequiresLocation,
	}

	templ, templateParseErr := template.New("solface").Funcs(templateFuncs).Parse(InterfaceTemplate)
	if templateParseErr != nil {
		return templateParseErr
	}
	templateExecutionErr := templ.Execute(writer, spec)

	return templateExecutionErr
}
