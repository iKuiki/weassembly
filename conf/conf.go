package conf

import (
	"github.com/jinzhu/configor"

	"github.com/kataras/golog"

	"github.com/pkg/errors"
)

// Conf 配置
type Conf interface {
	// LoadJSON 从json文件载入配置
	LoadJSON(jsonFilepath string) (err error)
	GetLogger() *golog.Logger
	GetWegateURL() string
	GetWegatePassword() string
	GetModuleConf(moduleName string) (moduleConf ModuleConf, ok bool)
}

// NewConfig 创建新的配置实例
func NewConfig() Conf {
	return new(conf)
}

// ModuleConf 模块配置
type ModuleConf map[string]string

// conf 配置文件实现
type conf struct {
	// WegateURL wegate微信网关的url地址
	WegateURL string
	// WegatePassword wegate微信网关的接入密码
	WegatePassword string
	logger         *golog.Logger
	ModuleConfs    map[string]ModuleConf // ModuleType=>map[string]string
}

// LoadJSON 从json文件载入配置
func (c *conf) LoadJSON(jsonFilepath string) (err error) {
	err = configor.Load(c, jsonFilepath)
	if err != nil {
		return errors.WithStack(err)
	}
	err = c.Init()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Init 根据载入的参数初始化
func (c *conf) Init() (err error) {
	c.logger = golog.New()
	return nil
}

func (c *conf) GetLogger() *golog.Logger {
	return c.logger
}

func (c *conf) GetWegateURL() string {
	return c.WegateURL
}

func (c *conf) GetWegatePassword() string {
	return c.WegatePassword
}

func (c *conf) GetModuleConf(moduleName string) (moduleConf ModuleConf, ok bool) {
	moduleConf, ok = c.ModuleConfs[moduleName]
	if !ok {
		moduleConf = make(ModuleConf)
	}
	return
}
