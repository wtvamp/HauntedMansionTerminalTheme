// Package palettedata carries the palette YAML into the binaries.
//
// It exists only because go:embed cannot reach outside its own directory, and
// palette/haunted-mansion.yaml must stay the single copy. Do not put logic here;
// loading and validation live in internal/palette.
package palettedata

import _ "embed"

//go:embed haunted-mansion.yaml
var YAML []byte
