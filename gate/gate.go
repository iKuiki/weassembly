package gate

import (
	"encoding/json"
	"fmt"
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/liangdas/mqant/utils/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"sync"
	"time"
	"weassembly/conf"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"wegate/common"
	commontest "wegate/common/test"
)

// Gate 插件网关服务模块
type Gate interface {
	Serve(modules ...Module)
}

// MustNewGate 创建新的插件网关服务模块，若创建失败则直接panic
func MustNewGate(conf conf.Conf) (g Gate) {
	g, err := NewGate(conf)
	if err != nil {
		panic(err)
	}
	return
}

// NewGate 创建新的插件网关服务模块
func NewGate(conf conf.Conf) (g Gate, err error) {
	gt := &gate{
		conf:    conf,
		modules: make(map[string]Module),
	}
	err = gt.prepareConnect(conf)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	g = gt
	return
}

type gate struct {
	conf       conf.Conf
	w          commontest.Work // mqtt客户端
	token      string          // 微信plugin注册后的token
	modules    map[string]Module
	callerLock sync.Mutex
	connLock   sync.RWMutex // 连接是否可用的锁，当连接不可用时会锁起
}

// Serve 运行
func (g *gate) Serve(modules ...Module) {
	// 给module分配caller
	for _, module := range modules {
		module.SetCaller(g)
		logger := g.conf.GetLogger().Child("[" + module.GetName() + "]")
		logger.NewLine = true
		module.SetLogger(logger)
		g.modules[uuid.Rand().Hex()] = module
		// 运行它！
		go module.Run()
	}
	loginChan := make(chan wwdk.LoginChannelItem)
	contactChan := make(chan datastruct.Contact)
	messageChan := make(chan datastruct.Message)
	g.w.On("LoginStatus", func(client MQTT.Client, msg MQTT.Message) {
		var loginItem wwdk.LoginChannelItem
		if e := json.Unmarshal(msg.Payload(), &loginItem); e == nil {
			loginChan <- loginItem
		}
	})
	g.w.On("ModifyContact", func(client MQTT.Client, msg MQTT.Message) {
		var contact datastruct.Contact
		if e := json.Unmarshal(msg.Payload(), &contact); e == nil {
			contactChan <- contact
		}
	})
	g.w.On("NewMessage", func(client MQTT.Client, msg MQTT.Message) {
		var message datastruct.Message
		if e := json.Unmarshal(msg.Payload(), &message); e == nil {
			messageChan <- message
		}
	})
	for {
		select {
		case loginItem := <-loginChan:
			g.conf.GetLogger().Info("new loginItem: ", loginItem.Code)
			for _, module := range g.modules {
				go module.LoginStatusChange(loginItem)
			}
		case contact := <-contactChan:
			g.conf.GetLogger().Info("new contact: ", contact.NickName)
			for _, module := range g.modules {
				go module.ModifyContact(contact)
			}
		case message := <-messageChan:
			g.conf.GetLogger().Infof("new message[%s]{%v}: %s", message.FromUserName, message.MsgType, message.GetContent())
			for _, module := range g.modules {
				go module.ReciveMessage(message)
			}
		}
	}
}

// 准备连接
func (g *gate) prepareConnect(conf conf.Conf) (err error) {
	opts := g.w.GetDefaultOptions(conf.GetWegateURL())
	opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		conf.GetLogger().Info("ConnectionLost", err.Error())
		// 连接不可用，锁定连接锁
		g.connLock.Lock()
	})
	opts.SetOnConnectHandler(func(client MQTT.Client) {
		conf.GetLogger().Info("检测到已连接，开始执行登陆、注册方法流程")

		pass := conf.GetWegatePassword() + time.Now().Format(time.RFC822)
		hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			panic(errors.WithStack(err))
		}
		resp, _ := g.w.Request("Login/HD_Login", []byte(`{"username":"gate","password":"`+string(hashedPass)+`"}`))
		if resp.Ret != common.RetCodeOK {
			panic(errors.Errorf("登录失败: %s", resp.Msg))
		}
		resp, _ = g.w.Request("Wechat/HD_Wechat_RegisterMQTTPlugin", []byte(fmt.Sprintf(
			`{"name":"%s","description":"%s","loginListenerTopic":"%s","contactListenerTopic":"%s","msgListenerTopic":"%s"}`,
			"weAssembly",
			"子模块集合",
			"LoginStatus",
			"ModifyContact",
			"NewMessage",
		)))
		if resp.Ret != common.RetCodeOK {
			panic(errors.Errorf("注册plugin失败: %s", resp.Msg))
		}
		g.token = resp.Msg
		conf.GetLogger().Info("注册完成，获取到token: " + g.token)
		// 连接完成，则释放连接锁
		g.connLock.Unlock()
	})
	opts.SetAutoReconnect(true)
	// 先锁定连接状态
	g.connLock.Lock()
	err = g.w.Connect(opts)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	return
}
