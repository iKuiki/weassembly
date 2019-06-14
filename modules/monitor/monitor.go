package monitor

import (
	"fmt"
	"weassembly/gate"
	"wegate/wechat"
)

// MustNewMonitorModule 务必创建模块
func MustNewMonitorModule() (module gate.Module) {
	module, err := NewMonitorModule()
	if err != nil {
		panic(err)
	}
	return
}

// NewMonitorModule 创建模块
func NewMonitorModule() (module gate.Module, err error) {
	module = &monitorModule{
		addPluginChan:    make(chan wechat.PluginDesc),
		removePluginChan: make(chan wechat.PluginDesc),
	}
	return
}

type monitorModule struct {
	gate.BaseModule
	addPluginChan    chan wechat.PluginDesc
	removePluginChan chan wechat.PluginDesc
}

func (m *monitorModule) GetName() string {
	return "monitor"
}

// AddPlugin 处理添加插件事件
func (m *monitorModule) AddPlugin(plugDesc wechat.PluginDesc) {
	m.addPluginChan <- plugDesc
	return
}

// RemovePlugin 处理移除插件事件
func (m *monitorModule) RemovePlugin(plugDesc wechat.PluginDesc) {
	m.removePluginChan <- plugDesc
	return
}

func (m *monitorModule) Run() {
	m.BaseModule.Logger.Info("monitor模块开始运行")
	for {
		select {
		case pluginDesc := <-m.addPluginChan:
			contacts, _ := m.Caller.GetContactList()
			for _, contact := range contacts {
				if contact.IsStar() {
					m.Caller.SendTextMessage(contact.UserName, fmt.Sprintf("添加插件[%s]%s", pluginDesc.Name, pluginDesc.Description))
				}
			}
		case pluginDesc := <-m.removePluginChan:
			contacts, _ := m.Caller.GetContactList()
			for _, contact := range contacts {
				if contact.IsStar() {
					m.Caller.SendTextMessage(contact.UserName, fmt.Sprintf("移除插件[%s]%s", pluginDesc.Name, pluginDesc.Description))
				}
			}
		}
	}
}
