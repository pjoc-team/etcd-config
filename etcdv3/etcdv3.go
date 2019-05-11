package etcdv3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/etcd-io/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type (
	EtcdResponse   *clientv3.WatchResponse
	EtcdBindHandle func(*EtcdBind) error //bind , cur path ,cur value
	BindTargetFunc func() interface{}

	// Bind to key may be with opOpt.Range(or opOpt.Prefix), this should returns multiple instance.
	BindKV struct {
		Result         map[string]interface{}
		BindTargetFunc BindTargetFunc
	}

	BinderFunc interface {
		NewTarget() interface{}
	}

	EtcdConfig struct {
		Cache          string   `yaml:"Cache"`
		Endpoints      []string `yaml:"Endpoints"`
		Root           string   `yaml:"Root"`
		TimeoutSeconds int64    `yaml:"TimeoutSeconds"`
	}

	EtcdClient struct {
		BindMap    map[string]EtcdBind
		Config     clientv3.Config
		EtcdConfig EtcdConfig
		Client     clientv3.Client
	}

	EtcdBind struct {
		Mutex    *sync.Mutex
		KV       *BindKV
		Response EtcdResponse
		Key      string
		Process  []EtcdBindHandle
		Opts     []clientv3.OpOption
	}
)

var logger = log.New()

// Support json value binder
var JsonBindHandle = func(bind *EtcdBind) error {
	response := bind.Response
	logger.Debugf("JsonBindHandle: bind: %v ", bind)
	for _, e := range response.Events {
		logger.Debugf("e: %v ", e)

		eventType := e.Type
		switch eventType {
		case clientv3.EventTypeDelete:
			i := bind.KV.Result
			delete(i, string(e.Kv.Key))
			err := fmt.Errorf("Etcd is deleted: %s \n", bind.Key)
			logger.Errorf(err.Error())
			return err
		default:
			kv := e.Kv
			//bind.ParseFromKv(kv)
			//bind.Mutex.Lock()
			//defer bind.Mutex.Unlock()
			if kv == nil {
				logger.Errorf("kv is nil!!")
				return errors.New("kv is nil!!")
			}
			bindKV := bind.KV
			key := string(kv.Key)
			v := kv.Value
			target := bindKV.BindTargetFunc()
			json.Unmarshal(v, target)
			bind.KV.Result[key] = target
			logger.Debugf("Bind key: %s value: %s to target: %v \n", key, string(v), target)
		}
	}
	return nil
}

// Support string value binder
var StringBindFunc = func(bind *EtcdBind) error {
	response := bind.Response
	for _, e := range response.Events {
		eventType := e.Type
		switch eventType {
		case clientv3.EventTypeDelete:
			err := fmt.Errorf("Etcd is deleted: %s \n", bind.Key)
			logger.Errorf(err.Error())
			return err
		default:
			kv := e.Kv
			key := string(kv.Key)
			ptr := bind.KV.BindTargetFunc()
			logger.Debugf("bind key: %s to target: %v \n", key, ptr)
			strPtr := ptr.(*string)
			*strPtr = string(kv.Value)
			logger.Debugf("binded key: %s target: %v \n", key, *strPtr)
			bind.KV.Result[key] = strPtr
		}
	}
	return nil
}

func Init(etcdConfig *EtcdConfig) EtcdClient {
	c := clientv3.Config{Endpoints: etcdConfig.Endpoints, DialTimeout: time.Duration(etcdConfig.TimeoutSeconds) * time.Second}
	client := EtcdClient{}
	client.Config = c
	client.EtcdConfig = *etcdConfig
	client.BindMap = make(map[string]EtcdBind)

	cli, err := clientv3.New(c)
	if err != nil {
		logger.Errorf("Error when connection: %s error: %s \n", client.EtcdConfig.Endpoints, err.Error())
		panic(err)
	}
	client.Client = *cli
	return client
}

func (client *EtcdClient) Close() {
	client.Client.Close()
}

func (client *EtcdClient) Bind(target interface{}, path string, opts []clientv3.OpOption, handle ...EtcdBindHandle) {
	client.BindMux(target, &sync.Mutex{}, path, opts, handle...)
}

func (client *EtcdClient) BindWithMultiResult(f BindTargetFunc, path string, opts []clientv3.OpOption, handle ...EtcdBindHandle) map[string]interface{} {
	result := make(map[string]interface{})
	kv := &BindKV{
		result,
		f,
	}
	client.BindMuxMultiTargetWithKey(kv, &sync.Mutex{}, path, opts, handle...)
	return result
}

func (e *EtcdClient) GetKey(subkey string) string {
	if len(subkey) > 0 && subkey[0] == '/' {
		return subkey
	}
	if subkey == "" {
		return e.EtcdConfig.Root
	}
	return e.EtcdConfig.Root + "/" + subkey
}

func (e *EtcdClient) BindMux(target interface{}, mu *sync.Mutex, path string, opts []clientv3.OpOption, process ...EtcdBindHandle) {
	result := make(map[string]interface{})
	f := func() interface{} { return target }
	bindKv := &BindKV{
		BindTargetFunc: f,
		Result:         result,
	}
	bind := &EtcdBind{
		KV:      bindKv,
		Mutex:   mu,
		Process: process,
		Key:     e.GetKey(path),
		Opts:    opts,
	}
	go e.Process(bind)
}

func (e *EtcdClient) BindMuxMultiTargetWithKey(kv *BindKV, mu *sync.Mutex, path string, opts []clientv3.OpOption, process ...EtcdBindHandle) {
	bind := &EtcdBind{
		Mutex:   mu,
		KV:      kv,
		Process: process,
		Key:     e.GetKey(path),
		Opts:    opts,
	}
	go e.Process(bind)
}

func (e *EtcdClient) Process(bind *EtcdBind) error {
	e.Client.Get(context.Background(), bind.Key, bind.Opts...)
	resp, err := e.Client.Get(context.Background(), bind.Key, bind.Opts...)
	logger.Debugf("resp: %s kvs: %s \n", resp, resp.Kvs)
	if err != nil {
		return err
	}
	for _, kv := range resp.Kvs {
		if err := bind.ParseFromKv(kv); err != nil {
			logger.Errorf("Error to parse kv: %v error: %s", kv, err)
		}

	}

	watchResponse := e.Client.Watch(context.Background(), bind.Key, bind.Opts...)
	for response := range watchResponse {
		bind.ProcessAll(&response)
		for _, ev := range response.Events {
			logger.Debugf("%s %q : %q , version: %v pre: %q \n", ev.Type, ev.Kv.Key, ev.Kv.Value, ev.Kv.Version, ev.PrevKv)
			//fmt.Printf("pre: %v : %v\n", ev.PrevKv.Key, ev.PrevKv.Value)
		}
	}
	return nil
}

func (bind *EtcdBind) GetValue(kvs []*mvccpb.KeyValue) {
	for _, v := range kvs {
		bytes := v.Value
		fmt.Printf("Got value: %s from key: %s \n", string(bytes), v.Key)
	}
}

func (bind *EtcdBind) ParseFromKv(kv *mvccpb.KeyValue) error {
	var err error
	bind.Mutex.Lock()
	defer bind.Mutex.Unlock()
	if kv == nil {
		return err
	}
	//bindKV := bind.KV
	//key := string(kv.Key)
	//v := kv.Value
	//target := bindKV.BindTargetFunc()

	// fake response
	tmpResponse := &clientv3.WatchResponse{}
	e := &clientv3.Event{Type: clientv3.EventTypePut, Kv: kv}
	tmpResponse.Events = []*clientv3.Event{e}
	bind.Response = tmpResponse

	for _, process := range bind.Process {
		logger.Debugf("ParseFromKv... Binder: %v Process: %v is processing...\n", bind, process)
		err := process(bind)
		if err != nil {
			logger.Errorf("process error %s %s %v", bind.Key, err, tmpResponse)
			return err
		}
		//logger.Debugf("Use processor: %v parse key: %s value: %s to target type: %T target value: %s \n", process, key, v, target, target)
	}
	//err = json.Unmarshal(v, target)
	//logger.Debugf("parse key: %s value: %s to target type: %T target value: %s \n", key, v, target, target)
	//if err == nil {
	//	m := bindKV.Result
	//	m[key] = target
	//}
	return err
}

func (bind *EtcdBind) ProcessAll(resp *clientv3.WatchResponse) error {
	bind.Mutex.Lock()
	defer bind.Mutex.Unlock()
	bind.Response = resp
	for _, process := range bind.Process {
		logger.Debugf("Process: %v is processing response: %v \n", process, *resp)
		err := process(bind)
		if err != nil {
			//Log.Errorf("process error %s %s %v", bind.Key, err, *resp)
			return err
		}
	}
	return nil
}
