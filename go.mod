module weassembly

go 1.12

replace wegate => github.com/ikuiki/wegate v0.0.0-20190509101958-f40ab7f0009c

replace github.com/liangdas/mqant => github.com/ikuiki/mqant v1.8.1-0.20190427142930-7dabfa32d064

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/ikuiki/go-component v0.0.0-20171218165758-b9f2562e71d1
	github.com/ikuiki/wwdk v2.3.0+incompatible
	github.com/kataras/golog v0.0.0-20180321173939-03be10146386
	github.com/liangdas/mqant v1.8.1
	github.com/pkg/errors v0.8.1
	github.com/qiniu/api.v7 v7.2.5+incompatible
	github.com/qiniu/x v7.0.8+incompatible // indirect
	qiniupkg.com/x v7.0.8+incompatible // indirect
	wegate v0.0.0-00010101000000-000000000000
)
