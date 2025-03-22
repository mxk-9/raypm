package pkglua

import (
	"errors"
	"fmt"
)

var (
	FieldIsNotArray   = errors.New("FieldIsNotArray")
	FieldIsNil        = errors.New("FieldIsNil")
	UnsupportedSystem = errors.New("UnsupportedSystem")
	UnknownSystem     = errors.New("UnknownSystem")
	ParseStringFailed = errors.New("ParseStringFailed")
)

type LuaTableError struct {
	Err       error
	Value     any
	FieldName string
}

func (e *LuaTableError) Error() (err string) {
	switch e.Err {
	case FieldIsNotArray:
		err = fmt.Sprintf("%v: (%s) got = %v", e.Err, e.FieldName, e.Value)
	case ParseStringFailed:
		err = fmt.Sprintf("%v: got = %v", e.Err, e.Value)
	case FieldIsNil:
		err = fmt.Sprintf("%v: %s", e.Err, e.FieldName)
	}
	return
}

type ParseStringItemError struct {
	Err   error
	Value any
}

func (e *ParseStringItemError) Error() string {
	return fmt.Sprintf("%v: got = %v", e.Err, e.Value)
}

type SystemError struct {
	Err   error
	Value string
}

func (e *SystemError) Error() (err string) {
	switch e.Err {
	case UnknownSystem:
		err = fmt.Sprintf(
			"%s: got = %s\n",
			UnknownSystem, e.Value,
		)

	case UnsupportedSystem:
		err = fmt.Sprintf(
			"%s: supporting of '%s' is not implemented yet",
			UnsupportedSystem, e.Value,
		)

	}

	return
}
