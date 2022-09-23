package service

import (
	"errors"
	"fmt"
	"sync/atomic"
)

var names = make(map[string]ServiceIdentity)
var id int32

func NewServiceIdentity(name string) (ServiceIdentity, error) {
	if name == "" {
		return ServiceIdentity{}, errors.New("service name is required")
	}

	if _, ok := names[name]; ok {
		return ServiceIdentity{}, fmt.Errorf("service name (%s) already exists", name)
	}

	var id = atomic.AddInt32(&id, 1)
	if id > 500 {
		return ServiceIdentity{}, errors.New("too many services. This is likely a bug in the calling code")
	}

	s := ServiceIdentity{atomic.AddInt32(&id, 1), name}
	names[name] = s

	return s, nil
}

func (s *ServiceIdentity) Id() int32 {
	s.assertValid()
	return s.id
}

func (s *ServiceIdentity) Name() string {
	s.assertValid()
	return s.name
}

func (s *ServiceIdentity) assertValid() {
	if s.id == 0 {
		panic("service identity is not initialized. Use NewServiceIdentity to create a new identity")
	}
}
