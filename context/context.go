package context

import (
	"poa-manager/jsonWrapper"
	"sync"
)

type Context struct {
	Version string

	Configs Configs

	mutexConfig *sync.Mutex
}

type Configs struct {
	UpdateAddress          string
	UpdateCheckIntervalSec int
	PoaServerAddress       string
	PoaServerPort          int
	MqttBrokerAddress      string
	MqttPort               int
}

type DeviceType int

const (
	DeviceTypeNormal DeviceType = iota
	DeviceTypeDeeper
)

func NewContext() *Context {
	context := Context{
		// Configs: Configs{},
		mutexConfig: &sync.Mutex{},
	}
	context.Configs.ReadFile("config.json")
	return &context
}

func (context *Context) WriteConfig() {
	go func() {
		context.mutexConfig.Lock()
		context.Configs.WriteFile("config.json")
		context.mutexConfig.Unlock()
	}()
}

func (configs *Configs) ToJson() string {
	jsonConfig := jsonWrapper.NewJsonWrapper()
	if jsonConfig.MarshalValue(configs) {
		return jsonConfig.ToString()
	}
	return ""
}

func (configs *Configs) ReadFile(path string) {
	jsonConfig := jsonWrapper.NewJsonWrapper()
	jsonConfig.ReadJsonTo(path, configs)
}

func (configs *Configs) WriteFile(path string) {
	jsonConfig := jsonWrapper.NewJsonWrapper()
	if jsonConfig.MarshalValue(configs) {
		jsonConfig.WriteJson(path)
	}
}
