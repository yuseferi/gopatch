package gopatch

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestNewPatch(t *testing.T) {
	tests := []struct {
		name    string
		arg     io.Reader
		want    Patch
		wantErr error
	}{
		{
			"Creates a replace patch with string value",
			bytes.NewReader([]byte(`
                [
                	{
                		"op": "replace",
                		"path": "/field1",
                		"value": "new value"
                	}
                ]`)),
			Patch{&ReplaceOperation{operationData{
				[]string{"field1"},
				"new value",
			}}},
			nil,
		},
		{
			"Creates a replace patch with bool value",
			bytes.NewReader([]byte(`
                [
                	{
                		"op": "replace",
                		"path": "/field1",
                		"value": true
                	}
                ]`)),
			Patch{&ReplaceOperation{operationData{
				[]string{"field1"},
				true,
			}}},
			nil,
		},
		{
			"Creates a replace patch with number value",
			bytes.NewReader([]byte(`
				[
					{
						"op": "replace",
						"path": "/field1",
						"value": 1
					}
				]`)),
			Patch{&ReplaceOperation{operationData{
				[]string{"field1"},
				1.0,
			}}},
			nil,
		},
		{
			"Creates replace patch with null value",
			bytes.NewReader([]byte(`
				[
					{
						"op": "replace",
						"path": "/field1",
						"value": null
					}
				]`)),
			Patch{&ReplaceOperation{operationData{
				[]string{"field1"},
				nil,
			}}},
			nil,
		},
		{
			"Returns error if operation is not supported",
			bytes.NewReader([]byte(`
				[
					{
						"op": "notSupported",
						"path": "/field1",
						"value": null
					}
				]`)),
			nil,
			ErrUnsupportedOp,
		},
		{
			"Returns error if path is not valid",
			bytes.NewReader([]byte(`
				[
					{
						"op": "replace",
						"path": "//",
						"value": null
					}
				]`)),
			nil,
			ErrInvalidPath,
		},
		{
			"Returns error if path is longer that 1 field",
			bytes.NewReader([]byte(`
				[
					{
						"op": "replace",
						"path": "/field1/field2",
						"value": 1
					}
				]`)),
			nil,
			ErrNotImplemented,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPatch(tt.arg)
			if (err != nil) && !errors.Is(err, tt.wantErr) {
				t.Errorf("NewPatch() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPatch() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceOperation_ApplyTo(t *testing.T) {
	tests := []struct {
		name        string
		opData      operationData
		arg         interface{}
		modifiedArg interface{}
		wantErr     error
	}{
		{
			"replaces string values",
			operationData{
				Path:  []string{"field1"},
				Value: "new value",
			},
			&struct {
				Field1 string
			}{
				Field1: "old value",
			},
			&struct {
				Field1 string
			}{
				Field1: "new value",
			},
			nil,
		},
		{
			"replaces bool values",
			operationData{
				Path:  []string{"field1"},
				Value: true,
			},
			&struct {
				Field1 bool
			}{
				Field1: false,
			},
			&struct {
				Field1 bool
			}{
				Field1: true,
			},
			nil,
		},
		{
			"finds field by tag and replaces the value",
			operationData{
				Path:  []string{"not_standard_field_name"},
				Value: "new value",
			},
			&struct {
				Field1 string `patch_field:"not_standard_field_name"`
			}{
				Field1: "old value",
			},
			&struct {
				Field1 string `patch_field:"not_standard_field_name"`
			}{
				Field1: "new value",
			},
			nil,
		},
		{
			"replaces with a zero value if new value is null",
			operationData{
				Path:  []string{"field1"},
				Value: nil,
			},
			&struct {
				Field1 string
			}{
				Field1: "old value",
			},
			&struct {
				Field1 string
			}{
				Field1: "",
			},
			nil,
		},
		{
			"returns error if target object is not a pointer to a struct",
			operationData{
				Path:  []string{"field1"},
				Value: "new value",
			},
			struct {
				Field1 string
			}{
				Field1: "old value",
			},
			nil,
			ErrApplyOp,
		},
		{
			"returns error if field is not found in the target object",
			operationData{
				Path:  []string{"not_standard_field_name"},
				Value: "new value",
			},
			&struct {
				Field1 string
			}{
				Field1: "old value",
			},
			nil,
			ErrApplyOp,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ReplaceOperation{
				operationData: tt.opData,
			}
			err := r.ApplyTo(tt.arg)
			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ApplyTo() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if !reflect.DeepEqual(tt.arg, tt.modifiedArg) {
				t.Errorf("ApplyTo() got = %v, want %v", tt.arg, tt.modifiedArg)
			}
		})
	}
}
