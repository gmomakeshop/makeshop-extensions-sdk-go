// Package contentblocks builds and reads Content Blocks v1 JSON.
//
// Use Marshal or Blocks to encode blocks with their type discriminator. Individual
// structs represent payloads: for example, Card.Image contains an image payload
// without a type field. Optional pointers distinguish omission from a zero value.
//
// Parsing checks JSON structure and enum values. Container hierarchy, text and URL
// restrictions, table row widths, and output limits are validated by the runtime.
package contentblocks

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Block is a supported block payload. Encode it with Marshal or inside Blocks.
type Block interface {
	blockType() string
}

// Blocks is a sequence of blocks encoded with their type discriminators.
// A nil sequence encodes as an empty JSON array.
type Blocks []Block

// Marshal encodes blocks as the blocks-json string of a content fragment.
// It rejects nil blocks. It does not validate content or container hierarchy.
func Marshal(blocks ...Block) (string, error) {
	data, err := json.Marshal(Blocks(blocks))
	return string(data), err
}

// Parse decodes a blocks-json string. It rejects unknown types, malformed payloads,
// and missing or unknown fields. It does not validate content or container hierarchy.
func Parse(data string) (Blocks, error) {
	var blocks Blocks
	err := json.Unmarshal([]byte(data), &blocks)
	return blocks, err
}

// MarshalJSON encodes the array with a type discriminator on each block.
func (blocks Blocks) MarshalJSON() ([]byte, error) {
	encoded := make([]json.RawMessage, 0, len(blocks))
	for i, block := range blocks {
		if block == nil {
			return nil, fmt.Errorf("block %d is nil", i)
		}
		payload, err := json.Marshal(block)
		if err != nil {
			return nil, fmt.Errorf("block %d: %w", i, err)
		}
		if bytes.Equal(payload, []byte("null")) {
			return nil, fmt.Errorf("block %d is nil", i)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(payload, &fields); err != nil {
			return nil, err
		}
		fields["type"], err = json.Marshal(block.blockType())
		if err != nil {
			return nil, err
		}
		value, err := json.Marshal(fields)
		if err != nil {
			return nil, err
		}
		encoded = append(encoded, value)
	}
	return json.Marshal(encoded)
}

// UnmarshalJSON decodes an array of block payloads selected by their type fields.
func (blocks *Blocks) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("blocks must be an array")
	}
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	result := make(Blocks, 0, len(items))
	for i, item := range items {
		fields, err := objectFields(item)
		if err != nil {
			return fmt.Errorf("block %d: %w", i, err)
		}
		var kind string
		if err := json.Unmarshal(fields["type"], &kind); err != nil {
			return fmt.Errorf("block %d type: %w", i, err)
		}
		delete(fields, "type")
		payload, err := json.Marshal(fields)
		if err != nil {
			return err
		}
		block, err := decodeBlock(kind, payload)
		if err != nil {
			return fmt.Errorf("block %d: %w", i, err)
		}
		result = append(result, block)
	}
	*blocks = result
	return nil
}

func objectFields(data []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, fmt.Errorf("expected a JSON object")
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		name, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("expected a field name")
		}
		if _, exists := fields[name]; exists {
			return nil, fmt.Errorf("duplicate field %q", name)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		fields[name] = value
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	return fields, nil
}

type fieldShape struct {
	required   bool
	arrayDepth int
}

func checkArray(data []byte, depth int) error {
	if depth == 0 {
		return nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	for _, item := range items {
		if bytes.Equal(bytes.TrimSpace(item), []byte("null")) {
			return fmt.Errorf("array item must not be null")
		}
		if err := checkArray(item, depth-1); err != nil {
			return err
		}
	}
	return nil
}

func decodeObject(data []byte, target any, shape map[string]fieldShape) error {
	fields, err := objectFields(data)
	if err != nil {
		return err
	}
	for key := range fields {
		if _, ok := shape[key]; !ok {
			return fmt.Errorf("unknown field %q", key)
		}
	}
	for key, field := range shape {
		value, ok := fields[key]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			if field.required {
				return fmt.Errorf("missing or null field %q", key)
			}
			continue
		}
		if err := checkArray(value, field.arrayDepth); err != nil {
			return fmt.Errorf("field %q: %w", key, err)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func nonNil[S ~[]E, E any](value S) S {
	if value == nil {
		return make(S, 0)
	}
	return value
}
