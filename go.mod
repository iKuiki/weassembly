module weassembly

go 1.12

replace wegate => github.com/ikuiki/wegate v1.0.1

replace github.com/liangdas/mqant => github.com/ikuiki/mqant v1.8.1-0.20190427142930-7dabfa32d064

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/ikuiki/go-component v0.0.0-20171218165758-b9f2562e71d1
	github.com/ikuiki/wwdk v2.4.0+incompatible
	github.com/jinzhu/configor v1.0.0
	github.com/kataras/golog v0.0.0-20180321173939-03be10146386
	github.com/liangdas/mqant v1.8.1
	github.com/pkg/errors v0.8.1
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	gopkg.in/yaml.v2 v2.2.2 // indirect
	wegate v0.0.0-00010101000000-000000000000
)
