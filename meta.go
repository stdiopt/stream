package stream

import "sync"

type meta struct {
	mu     sync.Mutex
	values map[string]interface{}
}

func (m *meta) SetAll(values map[string]interface{}) {
	if values == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.values == nil {
		m.values = map[string]interface{}{}
	}
	for k, v := range values {
		m.values[k] = v
	}
}

func (m *meta) Set(k string, v interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.values == nil {
		m.values = map[string]interface{}{}
	}
	if v == nil {
		delete(m.values, k)
		return
	}
	m.values[k] = v
}

func (m *meta) Get(k string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.values == nil {
		return nil
	}
	return m.values[k]
}

func (m *meta) Values() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.values == nil {
		return nil
	}
	values := map[string]interface{}{}
	for k, v := range m.values {
		values[k] = v
	}
	return values
}
