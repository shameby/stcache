package cache

import (
	"encoding/json"
	"io"
	"sync"
)

var slot *Slot

type Slot struct {
	data map[string]string
	sync.RWMutex
}

func init() {
	slot = newSlot()
}

func newSlot() *Slot {
	return &Slot{
		data: map[string]string{},
	}
}

func GetSlot() *Slot {
	if slot == nil {
		slot = newSlot()
	}
	return slot
}

func (s *Slot) Get(key string) string {
	s.RLock()
	defer s.RUnlock()
	return s.data[key]
}

func (s *Slot) Set(key, val string) {
	s.Lock()
	defer s.Unlock()
	s.data[key] = val
}

func (s *Slot) ToByte() ([]byte, error) {
	s.RLock()
	defer s.RUnlock()
	return json.Marshal(s.data)
}

func (s *Slot) FromIO(serialized io.ReadCloser) error {
	var tmp map[string]string
	if err := json.NewDecoder(serialized).Decode(&tmp); err != nil {
		return err
	}
	s.Lock()
	defer s.Unlock()
	s.data = tmp
	return nil
}