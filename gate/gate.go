package gate

import (
	"encoding/json"
	"fmt"
	"github.com/ikuiki/go-component/language"
	"github.com/ikuiki/wwdk"
	"github.com/ikuiki/wwdk/datastruct"
	"github.com/liangdas/mqant/utils/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"sync"
	"time"
	"weassembly/conf"
	"wegate/wechat"

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
		conf:       conf,
		modules:    make(map[string]Module),
		contactMap: make(map[string]datastruct.Contact),
	}
	err = gt.prepareConnect(conf)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// 准备联系人
	gt.prepareContact()
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
	// 以下属性是为了服务Caller用的属性
	contactMap     map[string]datastruct.Contact // 联系人map
	startedContact []string                      // 星标联系人
}

// Serve 运行
func (g *gate) Serve(modules ...Module) {
	// 给module分配caller
	for _, module := range modules {
		logger := g.conf.GetLogger().Child("[" + module.GetName() + "]")
		logger.NewLine = true
		moduleConf, _ := g.conf.GetModuleConf(module.GetName())
		module.Set(g, logger, moduleConf)
		g.modules[uuid.Rand().Hex()] = module
		// 运行它！
		go module.Run()
	}
	loginChan := make(chan wwdk.LoginChannelItem)
	contactChan := make(chan datastruct.Contact)
	messageChan := make(chan datastruct.Message)
	addPluginChan := make(chan wechat.PluginDesc)
	removePluginChan := make(chan wechat.PluginDesc)
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
	g.w.On("AddPlugin", func(client MQTT.Client, msg MQTT.Message) {
		var pluginDesc wechat.PluginDesc
		if e := json.Unmarshal(msg.Payload(), &pluginDesc); e == nil {
			addPluginChan <- pluginDesc
		}
	})
	g.w.On("RemovePlugin", func(client MQTT.Client, msg MQTT.Message) {
		var pluginDesc wechat.PluginDesc
		if e := json.Unmarshal(msg.Payload(), &pluginDesc); e == nil {
			removePluginChan <- pluginDesc
		}
	})
	for {
		select {
		case loginItem := <-loginChan:
			g.conf.GetLogger().Info("new loginItem: ", loginItem.Code)
			// 对loginItem预处理
			switch loginItem.Code {
			case wwdk.LoginStatusGotBatchContact:
				g.prepareContact()
			}
			for _, module := range g.modules {
				go module.LoginStatusChange(loginItem)
			}
		case contact := <-contactChan:
			g.conf.GetLogger().Info("new contact: ", contact.NickName)
			// 更新本地联系人map
			g.contactMap[contact.UserName] = contact
			// 检测是否为星标联系人，并有所处理
			if contact.IsStar() {
				// 添加星标联系人
				g.startedContact = language.ArrayUnique(append(g.startedContact, contact.UserName)).([]string)
			} else {
				// 从已有星标联系人中去除此非星标的联系人
				g.startedContact = language.ArrayDiff(g.startedContact, []string{contact.UserName}).([]string)
			}
			// 逐一通知各module
			for _, module := range g.modules {
				go module.ModifyContact(contact)
			}
		case message := <-messageChan:
			contact, _ := g.GetContactByUserName(message.FromUserName)
			g.conf.GetLogger().Infof("new message[%s]{%v}: %s", contact.NickName, message.MsgType, message.GetContent())
			for _, module := range g.modules {
				go module.ReciveMessage(message)
			}
		case pluginDesc := <-addPluginChan:
			g.conf.GetLogger().Infof("new plugin added: [%s]%s", pluginDesc.Name, pluginDesc.Description)
			for _, module := range g.modules {
				go module.AddPlugin(pluginDesc)
			}
		case pluginDesc := <-removePluginChan:
			g.conf.GetLogger().Infof("plugin removed: [%s]%s", pluginDesc.Name, pluginDesc.Description)
			for _, module := range g.modules {
				go module.RemovePlugin(pluginDesc)
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
			`{"name":"%s","description":"%s","loginListenerTopic":"%s","contactListenerTopic":"%s","msgListenerTopic":"%s","addPluginListenerTopic":"%s","removePluginListenerTopic":"%s"}`,
			"weAssembly",
			"子模块集合",
			"LoginStatus",
			"ModifyContact",
			"NewMessage",
			"AddPlugin",
			"RemovePlugin",
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

// 准备联系人
func (g *gate) prepareContact() (err error) {
	// 初始化联系人相关字段
	g.contactMap = make(map[string]datastruct.Contact)
	g.startedContact = []string{}
	// 更新联系人
	contacts, err := g.getContactList()
	if err != nil {
		return errors.WithStack(err)
	}
	for _, contact := range contacts {
		g.contactMap[contact.UserName] = contact
		if contact.IsStar() {
			g.startedContact = append(g.startedContact, contact.UserName)
		}
	}
	return
}
