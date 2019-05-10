package gate

import (
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
)

// Module 模块的定义
type Module interface {
	// LoginStatusChange 可以处理登陆状态发生改变
	LoginStatusChange(loginItem wwdk.LoginChannelItem)
	// ModifyContact 可以处理联系人发生变更
	ModifyContact(contact datastruct.Contact)
	// ReciveMessage 可以处理接受到信息
	ReciveMessage(msg datastruct.Message)
	// SetCaler 设置微信调用者
	SetCaller(caller Caller)
	// 运行对应的module
	Run()
}

// BaseModule 基础模块，提供了空的module方便组合
type BaseModule struct {
	Caller Caller
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

// SetCaller 设置微信调用者
func (m *BaseModule) SetCaller(caller Caller) {
	m.Caller = caller
	return
}

// Run 运行module
func (m *BaseModule) Run() {
	return
}
