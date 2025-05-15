package utils

import "encoding/json"

// JSONMarshal returns the JSON encoding of v.
type JSONMarshal func(v any) ([]byte, error)

// JSONUnmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v. If v is nil or not a pointer,
// Unmarshal returns an InvalidUnmarshalError.
type JSONUnmarshal func(data []byte, v any) error

// Use the stdlib JSON package for JSON marshaling and unmarshaling.
// if this becomes a performance bottleneck, we can consider using "github.com/bytedance/sonic"
var Marshal = json.Marshal
var Unmarshal = json.Unmarshal
