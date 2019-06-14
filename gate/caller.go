package gate

import (
	"encoding/json"
	"fmt"
	"github.com/ikuiki/go-component/language"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/pkg/errors"
	"wegate/common"
	"wegate/wechat/wechatstruct"
)

// Caller 对微信的调用服务
type Caller interface {
	// GetContactList 获取联系人列表
	GetContactList() (contacts []datastruct.Contact, err error)
	// SendTextMessage 发送文字消息
	SendTextMessage(toUserName, content string) (sendMessageRespond wechatstruct.SendMessageRespond, err error)
	// BroadcaseToStartedContact 向星标联系人广播消息
	BroadcaseToStartedContact(text string) error
	// ContactIsStarByUserName 根据userName判定联系人是否为星标联系人
	ContactIsStarByUserName(userName string) bool
	// GetContactByUserName 根据UserName获取Contact
	GetContactByUserName(userName string) (contact datastruct.Contact, err error)
}

// 从远程服务器获取联系人列表
func (g *gate) getContactList() (contacts []datastruct.Contact, err error) {
	g.callerLock.Lock()
	defer g.callerLock.Unlock()
	g.connLock.RLock()
	defer g.connLock.RUnlock()
	resp, _ := g.w.Request("Wechat/HD_Wechat_CallWechat", []byte(`{"fnName":"GetContactList","token":"`+g.token+`"}`))
	if resp.Ret != common.RetCodeOK {
		err = errors.Errorf("GetContactList失败: %s", resp.Msg)
		return
	}
	err = errors.WithStack(json.Unmarshal([]byte(resp.Msg), &contacts))
	return
}

// 为gate实现Caller接口

// 获取联系人列表
func (g *gate) GetContactList() (contacts []datastruct.Contact, err error) {
	for _, contact := range g.contactMap {
		contacts = append(contacts, contact)
	}
	return
}

// SendTextMessage 发送文本信息
func (g *gate) SendTextMessage(toUserName, content string) (sendMessageRespond wechatstruct.SendMessageRespond, err error) {
	g.callerLock.Lock()
	defer g.callerLock.Unlock()
	g.connLock.RLock()
	defer g.connLock.RUnlock()
	resp, _ := g.w.Request("Wechat/HD_Wechat_CallWechat", []byte(fmt.Sprintf(
		`{"fnName":"SendTextMessage","token":"%s","toUserName":"%s","content":"%s"}`,
		g.token,
		toUserName,
		content,
	)))
	if resp.Ret != common.RetCodeOK {
		err = errors.Errorf("SendTextMessage失败[%v]: %s", resp.Ret, resp.Msg)
		return
	}
	err = errors.WithStack(json.Unmarshal([]byte(resp.Msg), &sendMessageRespond))
	return
}

// BroadcaseToStartedContact 向星标联系人广播消息
func (g *gate) BroadcaseToStartedContact(text string) error {
	for _, userName := range g.startedContact {
		_, err := g.SendTextMessage(userName, text)
		if err != nil {
			return err
		}
	}
	return nil
}

// ContactIsStarByUserName 根据userName判定联系人是否为星标联系人
func (g *gate) ContactIsStarByUserName(userName string) bool {
	if language.ArrayIn(g.startedContact, userName) != -1 {
		return true
	}
	return false
}

// 根据UserName获取Contact
func (g *gate) GetContactByUserName(userName string) (contact datastruct.Contact, err error) {
	contact, ok := g.contactMap[userName]
	if !ok {
		err = errors.New("user not found")
	}
	return
}
