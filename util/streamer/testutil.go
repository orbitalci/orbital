package streamer

import "sync"

type testStreamArray struct {
	data [][]byte
	done bool
	mu sync.Mutex
}

func (t *testStreamArray) GetData() [][]byte {
	return t.data
}

func (t *testStreamArray) CheckDone() bool {
	return t.done
}

func (t *testStreamArray) AddToData(data [][]byte) {
	for _, datum := range data {
		t.data = append(t.data, datum)
	}
}

func (t *testStreamArray) Lock() {
	t.mu.Lock()
}

func (t *testStreamArray) Unlock() {
	t.mu.Unlock()
}

func NewTestStreamArray() *testStreamArray {
	var data [][]byte
	var done bool
	return &testStreamArray{
		data: data,
		done: done,
		mu: sync.Mutex{},
	}
}
