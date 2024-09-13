package resetableonce

import "sync"

type ResettableOnce struct {
	m    sync.Mutex
	once *sync.Once
}

func NewResettableOnce() *ResettableOnce {
	return &ResettableOnce{once: new(sync.Once)}
}

func (ro *ResettableOnce) Do(f func()) {
	ro.once.Do(f)
}

func (ro *ResettableOnce) Reset() {
	ro.m.Lock()
	defer ro.m.Unlock()
	ro.once = new(sync.Once)
}
