package lazyflag

type context struct {
	d map[string]any
}

func newcontext() *context {
	return &context{d: make(map[string]any)}
}
func (c *context) get(key string) any {
	return c.d[key]
}
func (c *context) set(key string, val any) {
	c.d[key] = val
}
