package parser

import (
	"errors"
)

type Module interface {
	Name() string
	TryGetScope() (*Scope, error)
	Is(Module) bool
}

type FileModule struct {
	name  string
	scope *Scope
}

func (m *FileModule) Name() string {
	return m.name
}

func (m *FileModule) TryGetScope() (*Scope, error) {
	return m.scope, nil
}

func (m *FileModule) Is(other Module) bool {
	if om, e := other.(*FileModule); e {
		return m.scope == om.scope
	}
	return false
}

type ReprModule struct {
	name string
}

func (m *ReprModule) Name() string {
	return m.name
}

func (m *ReprModule) TryGetScope() (*Scope, error) {
	return nil, errors.New("Trying to get scope from a module representation")
}

func (m *ReprModule) Is(other Module) bool {
	if om, e := other.(*ReprModule); e {
		return m.name == om.name
	}
	return false
}

var core Module = &ReprModule{"Core"}
