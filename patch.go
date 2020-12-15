package gopatch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Todo list
// float to in conversion
// support pointer fields
// support test op
// support nested structs
// comments and documentation
// allow user to provide operation logic

const (
	Test    = "test"
	Replace = "replace"
	Add     = "add"
	Remove  = "remove"

	FieldTag = "patch_field"
)

var (
	ErrNotImplemented = errors.New("not implemented yet")
	ErrInvalidPath    = errors.New("invalid path")
	ErrUnsupportedOp  = errors.New("unsupported operation")
	ErrApplyOp        = errors.New("cannot apply patch operation")
)

type Operation interface {
	Op() string
	Path() []string
	Value() interface{}
	ApplyTo(object interface{}) error
}

type Patch []Operation

type operationData struct {
	Path  []string
	Value interface{}
}

type JsonOperation struct {
	Op    string
	Path  string
	Value interface{}
}

func NewPatch(reader io.Reader) (Patch, error) {
	var newPatch []JsonOperation

	err := json.NewDecoder(reader).Decode(&newPatch)
	if err != nil {
		return nil, err
	}

	ops := make([]Operation, len(newPatch))
	for i, jsonOp := range newPatch {
		operation, err := toOperation(jsonOp)
		if err != nil {
			return nil, err
		}
		ops[i] = operation
	}

	return ops, nil
}

func ApplyPatch(patch Patch, object interface{}) error {
	for _, op := range patch {
		err := op.ApplyTo(object)
		if err != nil {
			return err
		}
	}
	return nil
}

func toOperation(jsonOp JsonOperation) (Operation, error) {
	path := strings.Split(jsonOp.Path, "/")
	fields := make([]string, 0)
	for _, f := range path {
		if f != "" {
			fields = append(fields, f)
		}
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPath, jsonOp.Path)
	}
	// temporary check until nested structs are implemented
	if len(fields) > 1 {
		return nil, fmt.Errorf("%w: %s", ErrNotImplemented, jsonOp.Path)
	}

	opData := operationData{
		Path:  fields,
		Value: jsonOp.Value,
	}

	switch jsonOp.Op {
	case Replace:
		return &ReplaceOperation{opData}, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedOp, jsonOp.Op)
}

type ReplaceOperation struct {
	operationData
}

func (r ReplaceOperation) Op() string {
	return Replace
}

func (r ReplaceOperation) Path() []string {
	return r.operationData.Path
}

func (r ReplaceOperation) Value() interface{} {
	return r.operationData.Value
}

func (r ReplaceOperation) String() string {
	return fmt.Sprintf("Op: %s, Path: %q, Value: %v", Replace, r.Path(), r.Value())
}

func (r ReplaceOperation) ApplyTo(object interface{}) error {
	target := reflect.ValueOf(object)
	if target.Kind() == reflect.Ptr && target.Elem().Kind() == reflect.Struct {
		target = findField(target.Elem(), r.Path()[0])
		if target.IsValid() && target.CanSet() {
			source := reflect.ValueOf(r.Value())
			if target.Kind() == source.Kind() {
				target.Set(source)
				return nil
			}
			if !source.IsValid() {
				zeroValue := reflect.Zero(target.Type())
				target.Set(zeroValue)
				return nil
			}
		}
	}

	return fmt.Errorf("%w {%s} to %v", ErrApplyOp, r, object)
}

func findField(structValue reflect.Value, fieldName string) reflect.Value {
	field := structValue.FieldByNameFunc(func(structFieldName string) bool {
		return strings.EqualFold(structFieldName, fieldName)
	})
	if field.IsValid() {
		return field
	}

	structType := structValue.Type()
	for i := 0; i < structValue.NumField(); i++ {
		tag := structType.Field(i).Tag.Get(FieldTag)
		if tag == fieldName {
			return structValue.Field(i)
		}
	}

	return reflect.Value{}
}
