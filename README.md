# Example

### Flag Use Example
```go
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	lazyflag "github.com/oswaldoooo/lazy-flag"
)

type Pair struct {
	Key   string
	Value string
}

func (p *Pair) UnmarshalText(text []byte) error {
	pairs := strings.Split(string(text), "-")
	if len(pairs) != 2 {
		return errors.New("pair is incorrect")
	}
	p.Key = pairs[0]
	p.Value = pairs[1]
	return nil
}

func main() {
	type Iface struct {
		Name             string
		InterfaceAddress string
	}
  // load p with short flag
	fmt.Println(lazyflag.Default.LoadAsBool(lazyflag.Short, "p"))
	var p Pair
	fmt.Println(lazyflag.Default.LoadAs(lazyflag.Long, "pair", &p))
	fmt.Println(p)
	data, err := lazyflag.LoadAsSlice[*Pair](lazyflag.Default, lazyflag.Long, "pairs")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error", err)
		return
	}
	for _, d := range data {
		fmt.Println(d)
	}
	var ifa Iface
  // bind Iface object with flag
	err = lazyflag.Default.Bind(&ifa)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error", err)
		return
	}
	fmt.Println(ifa)
}
```

### String Loader Example

```go
package lazyflag_test

import (
	"testing"

	lazyflag "github.com/oswaldoooo/lazy-flag"
)

type Getter map[string]string

func (g Getter) Get(k string) string {
	if v, ok := g[k]; ok {
		return v
	}
	return ""
}

func TestStringLoad(t *testing.T) {
	var (
		gg  = Getter{"name": "jesko", "ip_address": "127.0.0.1", "name_info": "this is info"}
		obj struct {
			Name      string
			IpAddress string
			NInfo     string `json:"name_info"`
		}
	)
	err := lazyflag.StringLoad(gg.Get, lazyflag.NewLoaderAttr("", nil), &obj)
	if err != nil {
		t.Fatal("error", err)
	}
	t.Log(obj)
}

```