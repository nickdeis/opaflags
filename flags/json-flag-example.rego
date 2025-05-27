package company.flags.jsonFlag

import rego.v1

description := "A json flag"

value := {"a":0} if input.name == "Alice" else := false
