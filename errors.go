package gocfgmodule

import "errors"

var (
	ErrEmptyModuleName = errors.New("module name is empty")
	ErrDuplicateModule = errors.New("duplicate module")
	ErrCircularDepends = errors.New("circular dependency detected")
	ErrModuleNotFound  = errors.New("module not found")
)
