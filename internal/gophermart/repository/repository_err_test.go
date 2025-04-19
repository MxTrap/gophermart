package storage

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewRepositoryError(t *testing.T) {
	type args struct {
		operation string
		err       error
	}
	tests := []struct {
		name string
		args args
		want *RepositoryError
	}{
		{
			name: "test repository error creation",
			args: args{
				operation: "create",
				err:       errors.New("test error"),
			},
			want: &RepositoryError{
				Operation: "create",
				Err:       errors.New("test error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRepositoryError(tt.args.operation, tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRepositoryError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepositoryError_Error(t *testing.T) {
	type fields struct {
		Operation string
		Err       error
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "test repository error creation",
			fields: fields{Operation: "create", Err: errors.New("test error")},
			want:   "RepositoryError: operation: create, error: test error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &RepositoryError{
				Operation: tt.fields.Operation,
				Err:       tt.fields.Err,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepositoryError_Unwrap(t *testing.T) {
	type fields struct {
		Operation string
		Err       error
	}
	tests := []struct {
		name   string
		fields fields
		err    string
	}{
		{
			name:   "test repository error unwrap",
			fields: fields{Operation: "create", Err: errors.New("test error")},
			err:    "test error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &RepositoryError{
				Operation: tt.fields.Operation,
				Err:       tt.fields.Err,
			}
			if err := e.Unwrap(); err.Error() != tt.err {
				t.Errorf("Unwrap() error = %v, wantErr %v", err, tt.err)
			}
		})
	}
}
