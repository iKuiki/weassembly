package main

import (
	"weassembly/conf"
	"weassembly/gate"
	"weassembly/modules/cue"
	"weassembly/modules/leave"
	"weassembly/modules/monitor"
)

func main() {
	c := conf.NewConfig()
	err := c.LoadJSON("config.json")
	if err != nil {
		panic(err)
	}
	gate := gate.MustNewGate(c)
	gate.Serve(leave.MustNewLeaveModule(), cue.MustNewCueModule(), monitor.MustNewMonitorModule())
}
