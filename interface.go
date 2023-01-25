package main

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

type ItemValueIndex struct {
	ItemIndex  int
	ValueIndex int
}

type NamedValue struct {
	Name  string
	Value Value
}

type CompoundType struct {
	TypeName string
	Members  []NamedValue
}

type DecodedABIWithCompundTypes struct {
	OriginalABI   DecodedABI
	CompoundTypes []CompoundType
	EnrichedABI   DecodedABI
}

type InterfaceSpecification struct {
	Name           string
	ABI            DecodedABI
	CompoundTypes  []CompoundType
	SolfaceVersion string
}

func GenerateName(nameCounter *int) string {
	result := fmt.Sprintf("Attribute%d", *nameCounter)
	(*nameCounter) += 1
	return result
}

func GenerateType(typeCounter *int) string {
	result := fmt.Sprintf("Compound%d", *typeCounter)
	(*typeCounter) += 1
	return result
}

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
		// It is not exactly "bytes" because that was handled above
		return false
	}

	return true
}

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
	compound.TypeName = GenerateType(typeCounter)
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

func GenerateInterface(interfaceName string, abi DecodedABI, writer io.Writer) error {
	const interfaceTemplate = `
// Interface generated by solface: https://github.com/bugout-dev/solface
// solface version: {{.SolfaceVersion}}
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
{{- range .ABI.Functions}}
	function {{.Name}}({{- range $i, $input := .Inputs}}{{if $i}}, {{end}}{{.Type}}{{if (needsMemory .Type)}} memory{{end}} {{.Name}} {{- end}}) external {{if (or (eq .StateMutability "view") (eq .StateMutability "pure"))}}{{.StateMutability}}{{end}}{{if .Outputs}} returns ({{- range $i, $output := .Outputs}}{{if $i}}, {{end}}{{.Type}}{{if (needsMemory .Type)}} memory{{end}}{{if .Name}} {{.Name}}{{end}}{{- end}}){{end}};
{{- end}}

	// errors
{{- range .ABI.Errors}}
	error {{.Name}}({{- range $i, $error := .Inputs}}{{if $i}}, {{end}}{{.Type}}{{.Name}}{{- end}});
{{- end}}
}
`

	resolved := ResolveCompounds(abi)
	spec := InterfaceSpecification{Name: interfaceName, ABI: resolved.EnrichedABI, CompoundTypes: resolved.CompoundTypes, SolfaceVersion: VERSION}

	templateFuncs := map[string]any{
		"needsMemory": SolidityTypeRequiresLocation,
	}

	templ, templateParseErr := template.New("yourface").Funcs(templateFuncs).Parse(interfaceTemplate)
	if templateParseErr != nil {
		return templateParseErr
	}
	templ.Execute(writer, spec)

	return nil
}
