package opaflags

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/open-policy-agent/opa/v1/rego"
)

type RawRego struct {
	text string
	name string
	err  error
}

func readFile(filepath string) *RawRego {
	content, err := os.ReadFile(filepath)
	raw_rego := RawRego{
		text: string(content),
		name: filepath,
		err:  err,
	}
	return &raw_rego
}

func readFlagFiles(file_path string) []*RawRego {
	files, _ := filepath.Glob(file_path)
	flag_files := allGo(files, readFile)
	return flag_files
}

func FromFilePath(glob_path string, namespace string) (OPAFlags, error) {
	modules := readFlagFiles(glob_path)
	return FromStructArray(modules, namespace)
}

type OPAFlags struct {
	raw_regos      []*RawRego
	rego_modules   []func(r *rego.Rego)
	rego_functions []func(r *rego.Rego)
	query_text     string
	namespace      string
	parent_context context.Context
	timed_context  context.Context
	prepped_query  rego.PreparedEvalQuery
}

func FromMap(raw_regos map[string]string, namespace string) (OPAFlags, error) {
	var regos []*RawRego
	for name, text := range raw_regos {
		regos = append(regos, &RawRego{name: name, text: text})
	}
	return FromStructArray(regos, namespace)
}

func FromStructArray(raw_regos []*RawRego, namespace string) (OPAFlags, error) {
	var f OPAFlags
	for _, rego := range raw_regos {
		if rego.err != nil {
			return f, rego.err
		}
	}
	modules := compileStructs(raw_regos)
	query_text := "output = data." + namespace
	ask_for := rego.Query(query_text)
	rego_functions := append([]func(r *rego.Rego){ask_for}, modules...)
	parent_context := context.Background()
	timed_context, _ := context.WithTimeout(parent_context, time.Millisecond*100)
	prepped_query, err := rego.New(
		rego_functions...,
	).PrepareForEval(timed_context)

	if err != nil {
		return f, err
	}
	f = OPAFlags{
		rego_modules:   modules,
		rego_functions: rego_functions,
		query_text:     query_text,
		namespace:      namespace,
		raw_regos:      raw_regos,
		parent_context: parent_context,
		timed_context:  timed_context,
		prepped_query:  prepped_query,
	}
	return f, nil
}

func mapArray[T any, U any](items []T, f func(t T, i int) U) []U {
	var u_arr []U
	for i, t := range items {
		u_arr = append(u_arr, f(t, i))
	}
	return u_arr
}

func allGo[T any, U any](items []T, f func(t T) U) []U {
	var wg sync.WaitGroup
	channel := make(chan U, len(items))
	run_f := func(t T) {
		defer wg.Done()
		channel <- f(t)
	}
	for _, t := range items {
		wg.Add(1)
		go run_f(t)
	}
	wg.Wait()
	close(channel)
	var u_arr []U
	for u := range channel {
		u_arr = append(u_arr, u)
	}
	return u_arr
}

func compileStructs(modules []*RawRego) []func(r *rego.Rego) {

	compile := func(r *RawRego) func(r *rego.Rego) {
		return rego.Module(r.name, r.text)
	}
	return allGo(modules, compile)
}

func (f *OPAFlags) EvaluateFlags(input map[string]any) (map[string]any, error) {
	timed_context, _ := context.WithTimeout(f.parent_context, time.Millisecond*100)
	results, err := f.prepped_query.Eval(timed_context, rego.EvalInput(input))
	output := results[0].Bindings["output"]
	return output.(map[string]any), err
}
