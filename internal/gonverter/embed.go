// Package gonverter provides the code generation logic for struct conversions.
package gonverter

import _ "embed"

//go:embed templates/converter.go.tmpl
var converterTemplate string
