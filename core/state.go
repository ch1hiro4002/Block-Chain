package core

import "fmt"

type State struct {
	data map[string][]byte
}

func NewState() *State {
	return &State{
		data: make(map[string][]byte),
	}
}

func (s *State) Put(k, v []byte) {
	s.data[string(k)] = v
}

func (s *State) Get(k []byte) ([]byte, error) {
	key := string(k)

	value, ok := s.data[key]
	if !ok {
		return nil, fmt.Errorf("key (%s) not found value", key)
	}

	return value, nil
}
