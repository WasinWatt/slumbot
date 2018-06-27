package cache

// New creates new cache
func New() Cacher {
	return &storage{
		m: make(map[string]interface{}),
	}
}

type Cacher interface {
	Get(string) (interface{}, bool)
	Set(string, interface{})
}

type storage struct {
	m map[string]interface{}
}

func (c *storage) Get(key string) (interface{}, bool) {
	val, ok := c.m[key]
	if !ok {
		return "", false
	}

	return val, ok
}

func (c *storage) Set(key string, val interface{}) {
	c.m[key] = val
}
