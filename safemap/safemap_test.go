package safemap

import (
	"math/rand"
	"sync"
	"testing"
)

// Год (79) взят из известного события: извержение Везувия и гибель Помпей (79 год н.э.).
// Диапазон ключей в тесте: от 1 до 79.

const year = 79

func TestSafeMap_ConcurrentCreateAndIncrement(t *testing.T) {
	t.Parallel()

	m := New()

	ops := make([]int, 0, year*3)
	for rep := 0; rep < 3; rep++ {
		for k := 1; k <= year; k++ {
			ops = append(ops, k)
		}
	}

	r := rand.New(rand.NewSource(20260326))
	r.Shuffle(len(ops), func(i, j int) { ops[i], ops[j] = ops[j], ops[i] })

	inc := true
	dec := true
	for i := 1; i < len(ops); i++ {
		if ops[i] <= ops[i-1] {
			inc = false
		}
		if ops[i] >= ops[i-1] {
			dec = false
		}

		if !inc && !dec {
			break
		}
	}
	if inc || dec {
		t.Fatalf("operations ended up sequential (inc=%v dec=%v)", inc, dec)
	}

	const goroutines = 4
	chunk := (len(ops) + goroutines - 1) / goroutines

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		start := g * chunk
		end := start + chunk
		if start > len(ops) {
			start = len(ops)
		}
		if end > len(ops) {
			end = len(ops)
		}

		part := ops[start:end]
		go func(part []int) {
			defer wg.Done()
			for _, k := range part {
				m.WithValue(k, func(v *int) {
					*v++
				})
			}
		}(part)
	}

	wg.Wait()

	values, accesses, adds := m.Snapshot()

	if adds != year {
		t.Fatalf("adds=%d, want=%d", adds, year)
	}
	if accesses != year*3 {
		t.Fatalf("accesses=%d, want=%d", accesses, year*3)
	}

	for k := 1; k <= year; k++ {
		v, ok := values[k]
		if !ok {
			t.Fatalf("missing key=%d", k)
		}
		if v != 3 {
			t.Fatalf("key=%d value=%d, want=3", k, v)
		}
	}
}
