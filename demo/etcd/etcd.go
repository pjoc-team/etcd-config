package main

import (
	"encoding/json"
	"fmt"
	"github.com/etcd-io/etcd/clientv3"
	"github.com/pjoc-team/etcd-config/etcdv3"
	"github.com/pjoc-team/etcd-config/yaml"
	"sync"
	"time"
)

type MyServiceConfig struct {
	Listen string            `yaml:"Listen" json:"Listen"`
	Etcd   etcdv3.EtcdConfig `yaml:"Etcd" json:"Etcd"`
}

type Student struct {
	Age    uint   `json:"age"`
	Name   string `json:"name"`
	Client *Client
	sync.Mutex
}

func (student *Student) GetClient() *Client {
	if student.Client == nil {
		student.Lock()
		if student.Client == nil {
			student.Client = &Client{Name: time.Now().String()}
		}
		student.Unlock()
	}
	return student.Client
}

type Client struct {
	Name string
}

func (student *Student) String() string {
	bytes, _ := json.Marshal(student)
	return string(bytes)
}

func main() {
	/**
	export HostIP="127.0.0.1"
	docker run --restart always -d -v /usr/share/ca-certificates/:/etc/ssl/certs -p 14001:4001 -p 2380:2380 -p 2379:2379 \
	 -e ETCDCTL_API=3 \
	 --name etcd quay.io/coreos/etcd:v3.2.18 \
	 etcd \
	 -name etcd0 \
	 -advertise-client-urls http://127.0.0.1:2379,http://127.0.0.1:4001 \
	 -listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:4001 \
	 -initial-advertise-peer-urls http://127.0.0.1:2380 \
	 -listen-peer-urls http://0.0.0.0:2380 \
	 -initial-cluster-token etcd-cluster-1 \
	 -initial-cluster etcd0=http://127.0.0.1:2380 \
	 -initial-cluster-state new
	*/
	/**
	docker exec -it etcd etcdctl put /pub/pjoc/go_test '{"name":"zhangsan"}'
	docker exec -it etcd etcdctl put /pub/pjoc/go_test123 '{"name":"哈哈哈哈"}'
	docker exec -it etcd etcdctl put /pub/pjoc/go_testa '{"name":"哈哈哈哈","age":18}'
	docker exec -it etcd etcdctl del /pub/pjoc/go_test
	**/
	config := &MyServiceConfig{}
	yaml.UnmarshalFromFile("./demo/etcd/config.yaml", config)

	fmt.Printf("watch: %v \n", config.Etcd.Root)
	client := etcdv3.Init(&config.Etcd)
	student := &Student{}
	os := []clientv3.OpOption{clientv3.WithPrefix()}
	bindFunc := func() interface{} { return &Student{} }
	result := client.BindWithMultiResult(bindFunc, "go_test", os, etcdv3.JsonBindHandle)

	time.Sleep(2 * time.Second)

	fmt.Printf("config: %s \n", student)

	for {
		time.Sleep(1 * time.Second)
		fmt.Println("result...", result)

		for k, v := range result {
			s := v.(*Student)
			fmt.Println("k: ", k, " v: ", v)
			fmt.Println(" client: ", s.GetClient())
		}
	}
}
