package share

import "sync"

type ItemSyncMap struct {
	m sync.Map
}

func (sMap *ItemSyncMap) Delete(key string) {
	sMap.m.Delete(key)
}

func (sMap *ItemSyncMap) Load(key string) (value ShareItem, ok bool) {
	v, ok := sMap.m.Load(key)
	if v != nil {
		value = v.(ShareItem)
	}
	return
}

func (sMap *ItemSyncMap) LoadOrStore(key string, value ShareItem) (actual ShareItem, loaded bool) {
	a, loaded := sMap.m.LoadOrStore(key, value)
	actual = a.(ShareItem)
	return
}

func (sMap *ItemSyncMap) Range(f func(key string, value ShareItem) bool) {
	f1 := func(key, value interface{}) bool {
		return f(key.(string), value.(ShareItem))
	}
	sMap.m.Range(f1)
}

func (sMap *ItemSyncMap) Store(key string, value ShareItem) {
	sMap.m.Store(key, value)
}
