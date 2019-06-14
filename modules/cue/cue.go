package cue

import (
	"fmt"
	"github.com/ikuiki/go-component/language"
	"github.com/ikuiki/wwdk/datastruct"
	"strings"
	"weassembly/gate"
)

// MustNewCueModule 务必创建模块
func MustNewCueModule() (module gate.Module) {
	module, err := NewCueModule()
	if err != nil {
		panic(err)
	}
	return
}

// NewCueModule 创建模块
func NewCueModule() (module gate.Module, err error) {
	module = &cueModule{
		newMsgChan: make(chan datastruct.Message),
	}
	return
}

type cueModule struct {
	gate.BaseModule
	startedContact []string // 星标联系人
	cues           []string // cue关键字
	newMsgChan     chan datastruct.Message
}

func (m *cueModule) GetName() string {
	return "cue"
}

// ModifyContact 可以处理联系人发生变更
func (m *cueModule) ModifyContact(contact datastruct.Contact) {
	// 检测是否为星标联系人，并有所处理
	if contact.IsStar() {
		// 添加星标联系人
		m.startedContact = language.ArrayUnique(append(m.startedContact, contact.UserName)).([]string)
	} else {
		// 从已有星标联系人中去除此非星标的联系人
		m.startedContact = language.ArrayDiff(m.startedContact, []string{contact.UserName}).([]string)
	}
	return
}

// ReciveMessage 可以处理接受到信息
func (m *cueModule) ReciveMessage(msg datastruct.Message) {
	m.newMsgChan <- msg
	return
}

// 向星标联系人广播消息
func (m *cueModule) broadcaseToStartedContact(text string) error {
	for _, userName := range m.startedContact {
		_, err := m.Caller.SendTextMessage(userName, text)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *cueModule) addCue(cmd string, msg datastruct.Message) {
	keys := strings.Split(cmd, ",")
	for k, key := range keys {
		keys[k] = strings.TrimSpace(key)
	}
	ocues := m.cues
	m.cues = language.ArrayMerge(ocues, keys).([]string)
	m.Caller.SendTextMessage(msg.FromUserName,
		fmt.Sprintf("添加了%d个cue词，当前cue词列表为: %s",
			len(m.cues)-len(ocues),
			strings.Join(m.cues, ",")))
	m.BaseModule.Logger.Infof("updated cues: %v", m.cues)
}

func (m *cueModule) removeCue(cmd string, msg datastruct.Message) {
	keys := strings.Split(cmd, ",")
	for k, key := range keys {
		keys[k] = strings.TrimSpace(key)
	}
	ocues := m.cues
	m.cues = language.ArrayDiff(ocues, keys).([]string)
	m.Caller.SendTextMessage(msg.FromUserName,
		fmt.Sprintf("移除了%d个cue词，当前cue词列表为: %s",
			len(m.cues)-len(ocues),
			strings.Join(m.cues, ",")))
	m.BaseModule.Logger.Infof("updated cues: %v", m.cues)
}

// 运行对应的module
func (m *cueModule) Run() {
	m.BaseModule.Logger.Info("cue模块开始运行")
	// 从配置中读取cue词
	cues := strings.Split(m.ModuleConf["Cues"], ",")
	m.cues = append(m.cues, cues...)
	m.BaseModule.Logger.Info("cues: ", m.cues)
	// 先同步现有chatroom
	contacts, _ := m.Caller.GetContactList()
	for _, contact := range contacts {
		if contact.IsStar() {
			// 记录星标联系人
			m.startedContact = append(m.startedContact, contact.UserName)
		}
	}
	m.BaseModule.Logger.Infof("检测星标联系人结束，共找到%d个星标联系人", len(m.startedContact))
	m.BaseModule.Logger.Info("cue模块开始服务")
	for {
		// 检测群消息是否含有cue
		msg := <-m.newMsgChan
		if msg.IsChatroom() {
			// 如果是群消息，则检测是否有cue关键字
			for _, cue := range m.cues {
				if strings.Contains(msg.GetContent(), cue) {
					m.BaseModule.Logger.Infof("检测到cue[%s]: %s", cue, msg.GetContent())
					// 有cue，向所有星标联系人发送消息
					err := m.broadcaseToStartedContact(fmt.Sprintf("好像有人cue你: 关键字: %s 原文: %s", cue, msg.GetContent()))
					if err != nil {
						m.BaseModule.Logger.Errorf("broadcaseToStartedContact fail: %+v", err)
					}
				}
			}
		} else {
			// 如果不是，则检测是否是星标联系人发的命令
			if language.ArrayIn(m.startedContact, msg.FromUserName) != -1 {
				// 是星标联系人发送的信息，看看是否包含指令
				content := msg.Content
				if strings.HasPrefix(content, "add cue:") {
					cmd := strings.TrimLeft(content, "add cue:")
					m.addCue(cmd, msg)
				} else if strings.HasPrefix(content, "remove cue:") {
					cmd := strings.TrimLeft(content, "remove cue:")
					m.removeCue(cmd, msg)
				} else if strings.HasPrefix(content, "+cue:") {
					cmd := strings.TrimLeft(content, "+cue:")
					m.addCue(cmd, msg)
				} else if strings.HasPrefix(content, "-cue:") {
					cmd := strings.TrimLeft(content, "-cue:")
					m.removeCue(cmd, msg)
				}
			}
		}
	}
}
