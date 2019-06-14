package leave

import (
	"github.com/ikuiki/go-component/language"
	"github.com/ikuiki/wwdk/datastruct"
	"strings"
	"weassembly/gate"
)

// MustNewLeaveModule 务必创建模块
func MustNewLeaveModule() (module gate.Module) {
	module, err := NewLeaveModule()
	if err != nil {
		panic(err)
	}
	return
}

// NewLeaveModule 创建模块
func NewLeaveModule() (module gate.Module, err error) {
	module = &leaveModule{
		chatrooms:         make(map[string][]string),
		contactModifyChan: make(chan datastruct.Contact),
	}
	return
}

type leaveModule struct {
	gate.BaseModule
	chatrooms         map[string][]string // 记录每个群内所有用户的昵称
	contactModifyChan chan datastruct.Contact
}

func (m *leaveModule) GetName() string {
	return "leave"
}

// ModifyContact 可以处理联系人发生变更
func (m *leaveModule) ModifyContact(contact datastruct.Contact) {
	m.contactModifyChan <- contact
	return
}

// 运行对应的module
func (m *leaveModule) Run() {
	m.BaseModule.Logger.Info("leave模块开始运行")
	// 先同步现有chatroom
	contacts, _ := m.Caller.GetContactList()
	for _, contact := range contacts {
		if contact.IsChatroom() {
			// 统计群成员NickName列表
			var nicknames []string
			for _, member := range contact.MemberList {
				nicknames = append(nicknames, member.NickName)
			}
			// 将更新的群成员记录
			m.chatrooms[contact.UserName] = nicknames
		}
	}
	m.BaseModule.Logger.Info("leave模块开始服务")
	for {
		contact := <-m.contactModifyChan
		if contact.IsChatroom() {
			// 统计群成员NickName列表
			var nicknames []string
			for _, member := range contact.MemberList {
				nicknames = append(nicknames, member.NickName)
			}
			// 检查是否已经存在此联系人
			oNicknames, ok := m.chatrooms[contact.UserName]
			if ok && len(nicknames) < len(oNicknames) {
				// 是旧的联系人，并且发现群成员减少,尝试找出这个人
				leaveMemberList := language.ArrayDiff(oNicknames, nicknames).([]string)
				m.Caller.SendTextMessage(contact.UserName, "检测到「"+strings.Join(leaveMemberList, ",")+"」疑似退出本群")
				m.Caller.BroadcaseToStartedContact("检测到「" + strings.Join(leaveMemberList, ",") + "」疑似退出「" + contact.NickName + "」")
			}
			// 将更新的群成员记录
			m.chatrooms[contact.UserName] = nicknames
		}
	}
}
