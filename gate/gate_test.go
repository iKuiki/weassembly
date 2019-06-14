package gate

import (
	"testing"
	"weassembly/conf"
)

var g *gate

func init() {
	c := conf.NewConfig()
	err := c.LoadJSON("../config.json")
	if err != nil {
		panic(err)
	}
	g = &gate{
		conf:    c,
		modules: make(map[string]Module),
	}
	err = g.prepareConnect(c)
	if err != nil {
		panic(err)
	}
}

func TestGetContactList(t *testing.T) {
	contacts, err := g.GetContactList()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(contacts)
}

func BenchmarkGetContactList(b *testing.B) {
	b.StartTimer()
	contacts, err := g.GetContactList()
	b.StopTimer()
	if err != nil {
		b.Fatal(err)
	}
	b.Logf("contact count: %d", len(contacts))
}
