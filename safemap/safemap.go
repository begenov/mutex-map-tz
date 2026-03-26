package safemap

import "sync"

// Автор: Бегенов Оразали

type SafeMap struct {
	mu sync.Mutex
	m  map[int]int

	accesses int
	adds     int
}

func New() *SafeMap {
	return &SafeMap{
		m: make(map[int]int),
	}
}

func (s *SafeMap) WithValue(key int, fn func(v *int)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.accesses++

	v, ok := s.m[key]
	if !ok {
		s.adds++
		v = 0
	}

	fn(&v)
	s.m[key] = v
}

func (s *SafeMap) Snapshot() (values map[int]int, accesses, adds int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	values = make(map[int]int, len(s.m))
	for k, v := range s.m {
		values[k] = v
	}
	return values, s.accesses, s.adds
}
