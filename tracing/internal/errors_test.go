package internal

import (
	"testing"

	"github.com/pkg/errors"
)

func TestUnwrap(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.Wrap(err1, "wrap1")
	err3 := errors.Wrap(err2, "wrap2")

	unwrapped := UnwrapError(err3)
	if unwrapped != err1 {
		t.Errorf("expected unwrapped error to be %v, got %v", err1, unwrapped)
	}
}

func TestUnwrap_NoWrap(t *testing.T) {
	err := errors.New("simple error")

	unwrapped := UnwrapError(err)
	if unwrapped != err {
		t.Errorf("expected unwrapped error to be %v, got %v", err, unwrapped)
	}
}

func TestUnwrap_Nil(t *testing.T) {
	var err error = nil

	unwrapped := UnwrapError(err)
	if unwrapped != nil {
		t.Errorf("expected unwrapped error to be nil, got %v", unwrapped)
	}
}
