# opaflags

A drop-dead simple, unopinionated file-based feature flag system using rego/opa

## Usage

Create a directory and make a .rego file in it

```
mkdir -p flags && touch ./flags/my-example.rego
```

Add some rego in it

```rego
package example.flags.exampleFlag
import rego.v1

description := "My example flag"

value := input.name == "Nick Deis"
```

`exampleFlag` is the flag name.

In your code, create an instance with the files and your namespace

```go
flags := opaflags.FromFilePath("*/*.rego","example.flags")
```

## Can I read from an S3 bucket/YAML/JSON/Over HTTP?

While it's not directly supported, you can always read the files into memory and use the the `FromMap` constructor.

```go
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
f := FromMap(TEST_MAP, TEST_NAMESPACE)
output := f.EvaluateFlags(TEST_INPUT)
ff := output["exampleFlag3"].(map[string]any)
```

## Why does the API return all flag values? What if I only want one?

By returning all flags, I can precompile the rego code and query. This means it's actually a little bit faster to return all flag values. If I'm doing something incorrectly here, let me know how to do it.

## Segments?

Segments are super simple with rego, create a new rego file with your segment logic:

```rego
package company.segments.example_segment


value := input.customer == "Acme"
```

And just import it into the flags you want to use

```rego
package company.flags.flagWithSegment
import rego.v1
import data.company.segments.example_segment


description := "My example flag using a segment"

value := example_segment.value
```

## Why?

There are other file based feature flag solutions, and I love them. This was meant to be super simple and work with existing OPA based deployments. But I highly suggest file based feature flagging, because:

### You can use git

With git, you have:

- Version control
- Change control and auditing

If you have Github or Gitlab, you have:

- Roles and permissions
- Triggers for when flag rules change
- APIs

### You can use existing tooling

`opa`:

- Formatting and linting
- Unit testing

`regal`:

- Syntax highlighting

All of these features typically cost thousands of dollars with "Enterprise" feature flagging solutions and are almost always half-baked in comparison. I will admit that some of these feature flagging solutions have certain niceties like analytics, but I almost never use them.
