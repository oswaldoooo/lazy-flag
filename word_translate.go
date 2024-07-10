package lazyflag

import "strings"

// example: HelloWorld => hello_world
func camel2hungarian(s string) string {
	var (
		raw            []string
		last_is_little bool
		last_index     int
	)
	if len(s) > 0 {
		last_is_little = !(s[0] >= 'A' && s[0] <= 'Z')
	}
	for i := 1; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			if last_is_little {
				raw = append(raw, s[last_index:i])
				last_index = i
			}
			last_is_little = false
		} else if s[i] >= 'a' && s[i] <= 'z' {
			last_is_little = true
		}
	}
	if last_index < len(s) {
		raw = append(raw, s[last_index:])
	}
	return strings.ToLower(strings.Join(raw, "_"))
}
