package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"poa-manager/context"
	"poa-manager/event"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type DeviceInfo struct {
	Timestamp int64

	DeviceId   string
	MacAddress string
	PublicIp   string
	PrivateIp  string
	Owner      string
	OwnNumber  int
	DeviceType int
	DeviceDesc string
	Version    string

	Alive bool
}

type Manager struct {
	Devices map[string]*DeviceInfo

	TotalDevices []*DeviceInfo
	DeadDevices  []*DeviceInfo

	serverAddress string
	serverPort    int

	brokerAddress  string
	brokerPort     int
	mqttClient     mqtt.Client
	mqttOpts       *mqtt.ClientOptions
	mqttQos        byte
	mqttClientName string
	mqttUser       string
	mqttPassword   string

	condChan chan int

	nofityUpdatedChan chan int
}

type DeadDevice struct {
	List []*DeviceInfo `json:"List,omitempty"`
	Hash string        `json:"Hash,omitempty"`
	Num  int           `json:"Num,omitempty"`
}

type Device struct {
	List       []*DeviceInfo `json:"List,omitempty"`
	TotalNum   int           `json:"TotalNum,omitempty"`
	Hash       string        `json:"Hash,omitempty"`
	DeadDevice *DeadDevice   `json:"DeadDevice,omitempty"`
}

type Remove struct {
	List []string `json:"List,omitempty"`
}

// server to client
type Response struct {
	Type string

	Device *Device `json:"Device,omitempty"`
	Remove *Remove `json:"Remove,omitempty"`
}

type Update struct {
	ForceUpdate   bool   `json:"ForceUpdate,omitempty"`
	UpdateAddress string `json:"UpdateAddress,omitempty"`
}

type Mqtt struct {
	MqttBrokerAddress string `json:"MqttBrokerAddress,omitempty"`
	MqttPort          int    `json:"MqttPort,omitempty"`
	MqttUser          string `json:"MqttUser,omitempty"`
	MqttPassword      string `json:"MqttPassword,omitempty"`
}

type Restart struct {
	Restart bool `json:"Restart,omitempty"`
}

// server to client
type Command struct {
	Type string

	Update  *Update  `json:"Update,omitempty"`
	Mqtt    *Mqtt    `json:"Mqtt,omitempty"`
	Restart *Restart `json:"Restart,omitempty"`
}

func NewManager() *Manager {
	return &Manager{Devices: make(map[string]*DeviceInfo)}
}

func (manager *Manager) mqttSubscribeHandler(client mqtt.Client, msg mqtt.Message) {
	go func() {
		// fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())

		if match, _ := regexp.MatchString("mine/server/updated", msg.Topic()); match {
			fmt.Println("rise mqtt updated message. start check status")
			manager.condChan <- 0
		} else if match, _ := regexp.MatchString("mine/[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+/.+/poa/info", msg.Topic()); match {
			fmt.Println("rise mqtt poa message. start check status")

			deviceInfo, err := manager.parsePayload(string(msg.Payload()))

			if err != nil {
				log.Println(err)
				return
			}

			var oldDeviceInfo DeviceInfo

			if _, ok := manager.Devices[deviceInfo.DeviceId]; ok {
				oldDeviceInfo = *manager.Devices[deviceInfo.DeviceId]
			}

			// Check changing the information displayed on the screen
			if deviceInfo.Owner != oldDeviceInfo.Owner ||
				deviceInfo.OwnNumber != oldDeviceInfo.OwnNumber ||
				deviceInfo.DeviceDesc != oldDeviceInfo.DeviceDesc ||
				deviceInfo.PublicIp != oldDeviceInfo.PublicIp ||
				deviceInfo.PrivateIp != oldDeviceInfo.PrivateIp ||
				deviceInfo.MacAddress != oldDeviceInfo.MacAddress ||
				deviceInfo.DeviceType != oldDeviceInfo.DeviceType ||
				deviceInfo.Timestamp != oldDeviceInfo.Timestamp {
				manager.condChan <- 0
			}
		}
	}()
}

func (manager *Manager) Init(poaContext *context.Context) {
	rand.Seed(time.Now().UnixNano())

	manager.serverAddress = poaContext.Configs.PoaServerAddress
	manager.serverPort = poaContext.Configs.PoaServerPort

	manager.brokerAddress = poaContext.Configs.MqttBrokerAddress
	manager.brokerPort = poaContext.Configs.MqttPort
	manager.mqttQos = 1
	manager.mqttClientName = fmt.Sprintf("poa-manager-%d", rand.Int31n(10000000))
	manager.mqttUser = poaContext.Configs.MqttUser
	manager.mqttPassword = poaContext.Configs.MqttPassword

	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)

	manager.mqttOpts = mqtt.NewClientOptions()
	manager.mqttOpts.AddBroker(fmt.Sprintf("tcp://%s:%d", manager.brokerAddress, manager.brokerPort))
	manager.mqttOpts.SetClientID(manager.mqttClientName)
	manager.mqttOpts.SetUsername(manager.mqttUser)
	manager.mqttOpts.SetPassword(manager.mqttPassword)
	manager.mqttOpts.SetDefaultPublishHandler(manager.mqttSubscribeHandler)
	manager.mqttOpts.SetAutoReconnect(true)
	manager.mqttOpts.OnConnect = func(client mqtt.Client) {
		fmt.Println("MQTT Connected")

		fmt.Println("Subscribe mine/#")
		token := manager.mqttClient.Subscribe("mine/#", manager.mqttQos, nil)
		token.Wait()
	}
	manager.mqttOpts.OnConnectionLost = func(client mqtt.Client, err error) {
		fmt.Printf("MQTT Connect lost: %v", err)
	}

	manager.condChan = make(chan int, 100)
	manager.nofityUpdatedChan = make(chan int)

	poaContext.EventLooper.RegisterEventHandler(event.MANAGER, manager.eventListener)
}

func (manager *Manager) Start() {
	var mqttInit func()
	mqttInit = func() {
		manager.mqttClient = mqtt.NewClient(manager.mqttOpts)

		if token := manager.mqttClient.Connect(); token.Wait() && token.Error() != nil {
			log.Println(token.Error())
			time.AfterFunc(time.Second*60, mqttInit)
			return
		}
	}
	go mqttInit()

	go func() {
		// run once at startup
		go func() {
			manager.condChan <- 0
		}()

		for {
			<-manager.condChan

			// get device status
			manager.TotalDevices, manager.DeadDevices = manager.getTotalDevices()

			for _, device := range manager.TotalDevices {
				manager.Devices[device.DeviceId] = device
			}

			manager.nofityUpdatedChan <- 0
		}
	}()
}

func (manager *Manager) getDeviceStatus() (total int, dead int) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/device/status", manager.serverAddress, manager.serverPort))
	if err == nil {
		bytes, _ := ioutil.ReadAll(resp.Body)
		str := string(bytes)
		fmt.Println(str)

		response := Response{Device: &Device{DeadDevice: &DeadDevice{}}}
		json.Unmarshal(bytes, &response)

		total = response.Device.TotalNum
		dead = response.Device.DeadDevice.Num
	} else {
		log.Println(err)
	}

	return
}

func (manager *Manager) getTotalDevices() ([]*DeviceInfo, []*DeviceInfo) {
	totalDevices := []*DeviceInfo{}
	deadDevices := []*DeviceInfo{}

	resp, err := http.Get(fmt.Sprintf("http://%s:%d/device/list", manager.serverAddress, manager.serverPort))
	if err == nil {
		bytes, _ := ioutil.ReadAll(resp.Body)
		// str := string(bytes)
		// fmt.Println(str)

		response := Response{}
		json.Unmarshal(bytes, &response)

		totalDevices = response.Device.List
		deadDevices = response.Device.DeadDevice.List
	} else {
		log.Println(err)
	}

	return totalDevices, deadDevices
}

func (manager *Manager) getDeadDevices() []*DeviceInfo {
	deadDevices := []*DeviceInfo{}

	resp, err := http.Get(fmt.Sprintf("http://%s:%d/device/dead/list", manager.serverAddress, manager.serverPort))
	if err == nil {
		bytes, _ := ioutil.ReadAll(resp.Body)
		str := string(bytes)
		fmt.Println(str)

		response := Response{Device: &Device{}}
		json.Unmarshal(bytes, &response)

		deadDevices = response.Device.DeadDevice.List
	} else {
		log.Println(err)
	}

	return deadDevices
}

func (manager *Manager) RemoveDevices(id string) (bool, string) {
	var reqBody string
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s:%d/device/remove/%s", manager.serverAddress, manager.serverPort, id), strings.NewReader(reqBody))
	if err != nil {
		log.Println(err)

		return false, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		bytes, _ := ioutil.ReadAll(resp.Body)
		str := string(bytes)
		fmt.Println(str)

		response := Response{ /* Remove: &Remove{} */ }
		json.Unmarshal(bytes, &response)

		if len(response.Remove.List) > 0 {
			return true, response.Remove.List[0]
		} else {
			return false, ""
		}
	} else {
		log.Println(err)
	}

	return false, ""
}

func (manager *Manager) parsePayload(payload string) (deviceInfo DeviceInfo, err error) {
	err = json.Unmarshal([]byte(payload), &deviceInfo)
	if err != nil {
		log.Println(err)
	}

	return
}

func (manager *Manager) WaitUpdated() {
	<-manager.nofityUpdatedChan
}

func (manager *Manager) eventListener(name event.EventName, args []interface{}) {
	fmt.Println("name:", name, args)

	switch name {
	case event.EVENT_MANAGER_DEVICE_RESTART:
		command := Command{Type: "restart", Restart: &Restart{}}
		command.Restart.Restart = true

		doc, err := json.MarshalIndent(command, "", "    ")
		if err == nil {
			for _, device := range manager.TotalDevices {
				if device.Alive {
					cmdAddress := fmt.Sprintf("mine/%s/%s/poa/command", device.PublicIp, device.DeviceId)

					fmt.Println("cmdAddress:", cmdAddress, " <- ", string(doc))

					token := manager.mqttClient.Publish(cmdAddress, manager.mqttQos, false, string(doc))
					token.Wait()
				}
			}
		} else {
			log.Println(err)
		}

	case event.EVENT_MANAGER_DEVICE_MQTT_CHANGE_USER_PASSWORD:
		if len(args) == 2 {
			command := Command{Type: "mqtt", Mqtt: &Mqtt{}}
			command.Mqtt.MqttUser = args[0].(string)
			command.Mqtt.MqttPassword = args[1].(string)

			doc, err := json.MarshalIndent(command, "", "    ")
			if err == nil {
				for _, device := range manager.TotalDevices {
					if device.Alive {
						cmdAddress := fmt.Sprintf("mine/%s/%s/poa/command", device.PublicIp, device.DeviceId)

						fmt.Println("cmdAddress:", cmdAddress, " <- ", string(doc))

						token := manager.mqttClient.Publish(cmdAddress, manager.mqttQos, false, string(doc))
						token.Wait()
					}
				}
			} else {
				log.Println(err)
			}
		}
	case event.EVENT_MANAGER_DEVICE_FORCE_UPDATE:
		command := Command{Type: "update", Update: &Update{}}
		command.Update.ForceUpdate = true

		doc, err := json.MarshalIndent(command, "", "    ")
		if err == nil {
			for _, device := range manager.TotalDevices {
				if device.Alive {
					cmdAddress := fmt.Sprintf("mine/%s/%s/poa/command", device.PublicIp, device.DeviceId)

					fmt.Println("cmdAddress:", cmdAddress, " <- ", string(doc))

					token := manager.mqttClient.Publish(cmdAddress, manager.mqttQos, false, string(doc))
					token.Wait()
				}
			}
		} else {
			log.Println(err)
		}

	case event.EVENT_MANAGER_DEVICE_CHANGE_UPDATE_ADDRESS:
		if len(args) == 1 {
			command := Command{Type: "update", Update: &Update{}}
			command.Update.UpdateAddress = args[0].(string)

			doc, err := json.MarshalIndent(command, "", "    ")
			if err == nil {
				for _, device := range manager.TotalDevices {
					if device.Alive {
						cmdAddress := fmt.Sprintf("mine/%s/%s/poa/command", device.PublicIp, device.DeviceId)

						fmt.Println("cmdAddress:", cmdAddress, " <- ", string(doc))

						token := manager.mqttClient.Publish(cmdAddress, manager.mqttQos, false, string(doc))
						token.Wait()
					}
				}
			} else {
				log.Println(err)
			}
		}
	}
}
