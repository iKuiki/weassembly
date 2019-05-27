package gate

import (
	"encoding/json"
	"fmt"
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
}

// 为gate实现Caller接口

// 获取联系人列表
func (g *gate) GetContactList() (contacts []datastruct.Contact, err error) {
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
