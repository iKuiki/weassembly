package gate

import (
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/kataras/golog"
	"weassembly/conf"
)

// Module 模块的定义
type Module interface {
	// GetName 获取模块名称
	GetName() string
	// LoginStatusChange 可以处理登陆状态发生改变
	LoginStatusChange(loginItem wwdk.LoginChannelItem)
	// ModifyContact 可以处理联系人发生变更
	ModifyContact(contact datastruct.Contact)
	// ReciveMessage 可以处理接受到信息
	ReciveMessage(msg datastruct.Message)
	// Set 传入配置
	Set(configs ...interface{})
	// 运行对应的module
	Run()
}

// BaseModule 基础模块，提供了空的module方便组合
type BaseModule struct {
	Caller     Caller
	Logger     *golog.Logger
	ModuleConf conf.ModuleConf
}

// LoginStatusChange 可以处理登陆状态发生改变
func (m *BaseModule) LoginStatusChange(loginItem wwdk.LoginChannelItem) {
	// 不处理
	return
}

// ModifyContact 可以处理联系人发生变更
func (m *BaseModule) ModifyContact(contact datastruct.Contact) {
	// 不处理
	return
}

// ReciveMessage 可以处理接受到信息
func (m *BaseModule) ReciveMessage(msg datastruct.Message) {
	// 不处理
	return
}

// Set 传入设置，自动识别为何种配置
func (m *BaseModule) Set(configs ...interface{}) {
	for _, config := range configs {
		switch config.(type) {
		case Caller:
			// Caller 设置微信调用者
			m.Caller = config.(Caller)
		case *golog.Logger:
			// Logger 设置日志输出器
			m.Logger = config.(*golog.Logger)
		case conf.ModuleConf:
			m.ModuleConf = config.(conf.ModuleConf)
		}
	}
}

// SetCaller 设置微信调用者
func (m *BaseModule) SetCaller(caller Caller) {
	m.Caller = caller
	return
}

// SetLogger 设置Logger
func (m *BaseModule) SetLogger(logger *golog.Logger) {
	m.Logger = logger
	return
}

// Run 运行module
func (m *BaseModule) Run() {
	return
}
