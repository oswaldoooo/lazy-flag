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
		gg  = Getter{"name": "jesko", "ip_address": "127.0.0.1", "name_info": "this is info", "vage": "23"}
		obj struct {
			Name      string
			IpAddress string
			NInfo     string `json:"name_info"`
			Age       int    `json:"age"`
		}
	)
	err := lazyflag.StringLoad(gg.Get, lazyflag.NewLoaderAttr("", nil), &obj, lazyflag.Alias{"age": "vage"})
	if err != nil {
		t.Fatal("error", err)
	}
	t.Log(obj)
}

func TestObjectBind(t *testing.T) {
	var data struct {
		ID   string
		Info struct {
			Name string `yaml:"iname"`
		}
		From string
	}
	err := lazyflag.ObjectBind(map[string]any{
		"id": "1234",
		"info": map[string]any{
			"ename": "jesko",
		},
		"from": "ca",
	}, &data, lazyflag.NewLoaderAttr("yaml", nil), lazyflag.Alias{"info.iname": "ename"})
	if err != nil {
		t.Fatal("bind error", err)
	}
	t.Logf("data %+v\n", data)
}
