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
	cues       []string // cue关键字
	newMsgChan chan datastruct.Message
}

func (m *cueModule) GetName() string {
	return "cue"
}

// ReciveMessage 可以处理接受到信息
func (m *cueModule) ReciveMessage(msg datastruct.Message) {
	m.newMsgChan <- msg
	return
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
	m.Logger.Infof("updated cues: %v", m.cues)
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
	m.Logger.Infof("updated cues: %v", m.cues)
}

// 运行对应的module
func (m *cueModule) Run() {
	m.Logger.Info("cue模块开始运行")
	// 从配置中读取cue词
	cues := strings.Split(m.ModuleConf["Cues"], ",")
	m.cues = append(m.cues, cues...)
	m.Logger.Info("cues: ", m.cues)
	m.Logger.Info("cue模块开始服务")
	for {
		// 检测群消息是否含有cue
		msg := <-m.newMsgChan
		if msg.IsChatroom() {
			// 如果是群消息，则检测是否有cue关键字
			for _, cue := range m.cues {
				if strings.Contains(msg.GetContent(), cue) {
					m.Logger.Infof("检测到cue[%s]: %s", cue, msg.GetContent())
					cueMessage := fmt.Sprintf("好像有人cue你\\n关键字: %s\\n原文: %s", cue, msg.GetContent())
					if chatroom, err := m.Caller.GetContactByUserName(msg.FromUserName); err == nil {
						memberUserName, _ := msg.GetMemberUserName()
						if member, err := chatroom.GetMember(memberUserName); err == nil {
							cueMessage = fmt.Sprintf("好像有人cue你\\n%s[%s]\\n关键字: %s\\n原文: %s",
								chatroom.NickName,
								member.NickName,
								cue,
								msg.GetContent())
						}
					}
					// 有cue，向所有星标联系人发送消息
					err := m.Caller.BroadcaseToStartedContact(cueMessage)
					if err != nil {
						m.Logger.Errorf("broadcaseToStartedContact fail: %+v", err)
					}
				}
			}
		} else {
			// 如果不是，则检测是否是星标联系人发的命令
			if m.Caller.ContactIsStarByUserName(msg.FromUserName) {
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
