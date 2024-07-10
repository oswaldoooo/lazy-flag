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
