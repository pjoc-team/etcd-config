package main

import (
	"fmt"
	cfg "github.com/pjoc-team/etcd-config/config"
	"github.com/pjoc-team/etcd-config/config/etcd"
	_ "github.com/pjoc-team/etcd-config/config/etcd"
	log "github.com/sirupsen/logrus"
	"time"
)

var defaultConfig = Config{
	Addr:     ":9090",
	LogLevel: "debug",
	DataSource: map[string]string{
		"sql":   "mysql://user:pass@127.0.0.1:306/db",
		"cache": "redis://127.0.0.1:6379:127.0.0.1:6380",
	},
	Services: []Service{

		Service{
			Name: "xtimer",
			Url:  "http://xtimer:8080/buket/example",
			Hooks: Hooks{
				Url: "http://127.0.0.1:9090/timeup", Key: "http",
			},
		},
		Service{
			Name: "userinfo",
			Url:  "kv://userinfo",
		},
	},
}

type MerchantConfigMap map[string]MerchantConfig

type MerchantConfig struct {
	AppId                string `json:"app_id"`
	GatewayRSAPublicKey  string `json:"gateway_rsa_public_key"`
	GatewayRSAPrivateKey string `json:"gateway_rsa_private_key"`
	MerchantRSAPublicKey string `json:"merchant_rsa_public_key"`
	Md5Key               string `json:"md5_key"`
}

func main() {
	config := &defaultConfig
	err := cfg.Init(cfg.URL("etcd://127.0.0.1:2379/com/test/demo/"), cfg.WithDefault(config))
	if err != nil {
		panic(err)
	}
	cfg.SetFieldListener("Services[0].Hooks", func(old, new interface{}) {
		log.Println("change from", old.(Hooks), "to", new.(Hooks))
	})
	cfg.SetFieldListener("DataSource.cache", func(old, new interface{}) {
		log.Println("todo something by", new.(string))
	})

	log.Println("change complete", config.Addr)

	configMap := &MerchantConfigMap{}
	i := cfg.New(map[string]interface{}{})
	i.AddBackend(etcd.SCHEMA, &etcd.EtcdBackend{})
	err2 := i.Init(cfg.URL("etcd://127.0.0.1:2379/pub/pjoc/pay/merchants/"), cfg.WithDefault(configMap))
	if err2 != nil {
		panic(err2)
	}
	fmt.Println("configMaps: ", configMap)
	i.SetFieldListener("app_id", func(old, new interface{}) {
		log.Println("change from", old.(Hooks), "to", new.(Hooks))
	})

	time.Sleep(time.Hour)
}

type Config struct {
	Addr       string
	LogLevel   string
	DataSource map[string]string
	Services   []Service
}

type Service struct {
	Name  string
	Url   string
	Hooks Hooks
}

type Hooks struct {
	Url string
	Key string
}
