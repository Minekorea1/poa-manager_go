package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"poa-manager/context"
	"poa-manager/manager"
	"poa-manager/res"
	"poa-manager/ui"
	"strings"

	poaUpdater "poa-manager/updater"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

const (
	VERSION_NAME                          = "v0.2.1"
	APPLICATION_UPDATE_ADDRESS            = "github.com/Minekorea1/poa-manager_go"
	APPLICATION_UPDATE_CHECK_INTERVAL_SEC = 3600
	POA_SERVER_ADDRESS                    = "minekorea.asuscomm.com"
	POA_SERVER_PORT                       = 8888
	MQTT_BROKER_ADDRESS                   = "minekorea.asuscomm.com"
	MQTT_PORT                             = 1883
)

func ternaryOP(cond bool, valTrue, valFalse interface{}) interface{} {
	if cond {
		return valTrue
	} else {
		return valFalse
	}
}

func emptyString(str string) bool {
	return strings.TrimSpace(str) == ""
}

func Initialize() *context.Context {
	context := context.NewContext()

	context.Version = VERSION_NAME
	context.Configs.UpdateAddress = ternaryOP(emptyString(context.Configs.UpdateAddress),
		APPLICATION_UPDATE_ADDRESS, context.Configs.UpdateAddress).(string)
	context.Configs.UpdateCheckIntervalSec = ternaryOP(context.Configs.UpdateCheckIntervalSec <= 0,
		APPLICATION_UPDATE_CHECK_INTERVAL_SEC, context.Configs.UpdateCheckIntervalSec).(int)
	context.Configs.PoaServerAddress = ternaryOP(emptyString(context.Configs.PoaServerAddress),
		POA_SERVER_ADDRESS, context.Configs.PoaServerAddress).(string)
	context.Configs.PoaServerPort = ternaryOP(context.Configs.PoaServerPort <= 0,
		POA_SERVER_PORT, context.Configs.PoaServerPort).(int)
	context.Configs.MqttBrokerAddress = ternaryOP(emptyString(context.Configs.MqttBrokerAddress),
		MQTT_BROKER_ADDRESS, context.Configs.MqttBrokerAddress).(string)
	context.Configs.MqttPort = ternaryOP(context.Configs.MqttPort <= 0,
		MQTT_PORT, context.Configs.MqttPort).(int)

	return context
}

func main() {
	versionFlag := false
	flag.BoolVar(&versionFlag, "version", false, "prints the version and exit")
	flag.Parse()

	if versionFlag {
		fmt.Println(VERSION_NAME)
		return
	}

	fmt.Printf("version: %s\n", VERSION_NAME)

	context := Initialize()

	updater := poaUpdater.NewUpdater()
	updater.Init(context)
	updater.Start()

	manager := manager.NewManager()
	manager.Init(context)
	manager.Start()

	// ui
	os.Setenv("FYNE_THEME", "light") // light or dark
	a := app.NewWithID("PoA-Manager")
	a.SetIcon(res.Ic_main)
	a.Settings().SetTheme(&ui.MyTheme{})
	a.Lifecycle().SetOnStarted(func() {
		log.Println("Lifecycle: Started")
	})
	a.Lifecycle().SetOnStopped(func() {
		log.Println("Lifecycle: Stopped")
	})
	win := a.NewWindow("PoA Manager " + VERSION_NAME)
	a.Settings().SetTheme(&ui.MyTheme{})
	win.SetMaster()

	ui.Init(&a, &win, context, manager)
	uiMenu := ui.Menu{}
	subContent := container.NewMax()

	// mainContent := container.NewHSplit(uiMenu.MakeMenu(), uiStatus.GetContainer())
	// mainContent.Offset = 0.2
	mainContent := container.NewBorder(nil, nil, uiMenu.MakeMenu(subContent), nil, subContent)
	win.SetContent(mainContent)
	win.Resize(fyne.NewSize(1120, 720))
	ui.RunUpdateThread()
	win.ShowAndRun()
}
