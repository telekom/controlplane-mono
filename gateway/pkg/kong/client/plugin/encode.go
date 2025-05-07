package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// StringMap is a map of strings that encodes to a JSON array of strings
// with the format [key1:value1, key2:value2]
type StringMap struct {
	items map[string]string
}

func New() *StringMap {
	return &StringMap{items: make(map[string]string)}
}

func (m *StringMap) AddKV(key, value string) {
	m.items[key] = value
}

func (m *StringMap) Add(value string) {
	parts := strings.Split(value, ":")
	m.items[parts[0]] = parts[1]
}

func (m *StringMap) RemoveK(key, value string) {
	delete(m.items, key)
}

func (m *StringMap) Remove(value string) {
	parts := strings.Split(value, ":")
	delete(m.items, parts[0])
}

func (m *StringMap) Clear() {
	m.items = make(map[string]string)
}

func (m *StringMap) Contains(key string) bool {
	if _, contains := m.items[key]; !contains {
		return false
	}
	return true
}

func (m *StringMap) Get(key string) string {
	if m.Contains(key) {
		return m.items[key]
	}
	return ""
}

// MarshalJSON encodes the map into a format like ["key1:value1", "key2:value2"]
func (m *StringMap) MarshalJSON() ([]byte, error) {
	if len(m.items) == 0 {
		return []byte("[]"), nil
	}

	var builder bytes.Buffer
	_, err := builder.WriteString("[")
	if err != nil {
		return nil, errors.Wrap(err, "failed to write to buffer")
	}
	for k, v := range m.items {
		_, err = builder.WriteString(fmt.Sprintf("\"%s:%s\",", k, v))
		if err != nil {
			return nil, errors.Wrap(err, "failed to write to buffer")
		}
	}
	result := builder.Bytes()
	// Replace the last comma with a closing bracket
	result[len(result)-1] = ']'
	return result, nil
}

// UnmarshalJSON decodes a string like ["key1:value1","key2:value2"] into a map
func (m *StringMap) UnmarshalJSON(b []byte) error {
	if m.items == nil {
		m.items = make(map[string]string)
	}
	if len(b) < 2 {
		return nil
	}
	var pairs []string
	if err := json.Unmarshal(b, &pairs); err != nil {
		return errors.Wrap(err, "failed to unmarshal JSON")
	}
	for _, pair := range pairs {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid key-value pair: %s", pair)
		}
		m.items[kv[0]] = kv[1]
	}
	return nil
}
