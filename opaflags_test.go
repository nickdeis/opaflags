package opaflags

import (
	"testing"
)

const TEST_NAMESPACE = "company.flags"
const TEST_FILEPATH = "*/*.rego"

var TEST_INPUT = map[string]any{
	"name":     "Alice",
	"customer": "Acme",
}

var f = FromFilePath(TEST_FILEPATH, TEST_NAMESPACE)

func TestBasicInput(t *testing.T) {
	for i := 1; i <= 100; i++ {
		output := f.EvaluateFlags(TEST_INPUT)
		flagWithSegment := output["flagWithSegment"].(map[string]any)
		if !flagWithSegment["value"].(bool) {
			t.Errorf("Failed match")
		}
	}
}

var TEST_MAP = map[string]string{
	//The name doesn't matter, but it does have to be unique
	"exampleFlag": `
	package company.flags.exampleFlag
	import rego.v1

	description := "My example flag"

	value := true
	`,
	"exampleFlag3": `
	package company.flags.exampleFlag3
	import rego.v1
	description := "My example flag 3"

	value := input.name == "Nick"
	`,
}

func TestFromMap(t *testing.T) {
	f := FromMap(TEST_MAP, "company.flags")
	output := f.EvaluateFlags(map[string]any{
		"name":     "Alice",
		"customer": "Acme",
	})
	ff := output["exampleFlag3"].(map[string]any)
	if ff["value"].(bool) {
		t.Errorf("Failed match")
	}
}
