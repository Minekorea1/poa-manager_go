package ui

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"poa-manager/context"
	"poa-manager/event"
	"poa-manager/log"
	"poa-manager/manager"
	"poa-manager/res"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var logger log.Logger = log.NewLogger("ui")

type CustomContent interface {
	GetContent() *fyne.Container
	SetMainContent()
}

type Menu struct {
	Title, Intro string
	content      CustomContent
}

var (
	poaContext *context.Context
	poaManager *manager.Manager

	menus     map[string]Menu
	menuIndex map[string][]string

	window *fyne.Window

	parentContainer *fyne.Container
	activeContect   *fyne.Container

	statusContent        *contentStatus
	structureContent     *contentStructure
	deviceControlContent *contentDeviceControl
	configContent        *contentConfig
)

type contentStatus struct {
	content             *fyne.Container
	labelStatus         *widget.Label
	checkDeadDeviceOnly *widget.Check
	listDevices         *widget.List
	listItem            *fyne.Container
	detailContent       *fyne.Container
	labelDetailID       *widget.Label
	labelDetailHeader   *widget.Label
	labelDetailData     *widget.Label
	buttonRemove        *widget.Button

	selectedDevice *manager.DeviceInfo
	deadDeviceOnly bool
}

type contentStructure struct {
	content           *fyne.Container
	treeDevices       *widget.Tree
	treeData          map[string][]string
	detailContent     *fyne.Container
	labelDetailID     *widget.Label
	labelDetailHeader *widget.Label
	labelDetailData   *widget.Label
	buttonRemove      *widget.Button

	selectedDevice *manager.DeviceInfo
}

type contentDeviceControl struct {
	content             *fyne.Container
	buttonMqttUserPwd   *widget.Button
	buttonUpdateAddress *widget.Button
	buttonForceUpdate   *widget.Button
	buttonForceRestart  *widget.Button
}

type contentConfig struct {
	content            *fyne.Container
	serverAddressEntry *widget.Entry
	serverPortEntry    *numericalEntry
	mqttAddressEntry   *widget.Entry
	mqttPortEntry      *numericalEntry
	mqttUserEntry      *widget.Entry
	mqttPasswordEntry  *widget.Entry
}

func Init(_ *fyne.App, win *fyne.Window, ctx *context.Context, m *manager.Manager) {
	window = win

	poaContext = ctx
	poaManager = m

	statusContent = newStatusContent()
	structureContent = newStructureContent()
	deviceControlContent = newCommandDeviceControl()
	configContent = newConfigContent()

	menus = map[string]Menu{
		"status":        {"????????????", "????????? ???????????? ?????? ????????? ???????????????.", statusContent},
		"structure":     {"??????????????? ??????", "????????? ???????????? ????????? ???????????????.", structureContent},
		"deviceControl": {"?????? ??????", "?????? ???????????? ?????? ???????????? ???????????????..", deviceControlContent},
		"configs":       {"??????", "????????? ?????? ????????? ??? ??? ????????????.", configContent},
	}

	menuIndex = map[string][]string{
		"": {"status", "structure", "deviceControl", "configs"},
		// "collections": {"list", "table", "tree"},
	}
}

func (menu *Menu) MakeMenu(parent *fyne.Container) *fyne.Container {
	parentContainer = parent
	statusContent.SetMainContent()

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return menuIndex[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := menuIndex[uid]

			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			t, ok := menus[uid]
			if !ok {
				fyne.LogError("Missing panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(t.Title)
		},
		OnSelected: func(uid string) {
			if m, ok := menus[uid]; ok {
				m.content.SetMainContent()

				if activeContect == statusContent.content {
					statusContent.listDevices.UnselectAll()
					statusContent.detailContent.Hide()
					statusContent.selectedDevice = nil
				} else if activeContect == structureContent.content {
					structureContent.updateTreeView()
					structureContent.treeDevices.OpenAllBranches()

					structureContent.treeDevices.UnselectAll()
					structureContent.detailContent.Hide()
					structureContent.selectedDevice = nil
				} else if activeContect == configContent.content {
					configContent.serverAddressEntry.SetText(poaContext.Configs.PoaServerAddress)
					configContent.serverPortEntry.SetText(strconv.FormatInt(int64(poaContext.Configs.PoaServerPort), 10))
					configContent.mqttAddressEntry.SetText(poaContext.Configs.MqttBrokerAddress)
					configContent.mqttPortEntry.SetText(strconv.FormatInt(int64(poaContext.Configs.MqttPort), 10))
					configContent.mqttUserEntry.SetText(poaContext.Configs.MqttUser)
					configContent.mqttPasswordEntry.SetText(poaContext.Configs.MqttPassword)
				}
			}
		},
	}

	return container.NewMax(tree)
}

func RunUpdateThread() {
	go func() {
		for {
			poaManager.WaitUpdated()

			if activeContect == statusContent.content {
				totalCount := len(poaManager.TotalDevices)
				deadCount := len(poaManager.DeadDevices)
				statusContent.labelStatus.SetText(fmt.Sprintf("??????: %d ???, ??????: %d ???, ?????? ??????: %d ???", totalCount, totalCount-deadCount, deadCount))

				statusContent.listDevices.Refresh()
				statusContent.updateDetailView(statusContent.selectedDevice)

				// status.labelOwner.SetText(fmt.Sprintf("?????????: %s", deviceInfo.Owner))
				// status.labelOwnNumber.SetText(fmt.Sprintf("?????? ??????: %d", deviceInfo.OwnNumber))
				// status.labelDesc.SetText(fmt.Sprintf("??????: %s", deviceInfo.DeviceDesc))
				// status.labelPublicIp.SetText(fmt.Sprintf("??????IP: %s", deviceInfo.PublicIp))
				// status.labelPrivateIp.SetText(fmt.Sprintf("??????IP: %s", deviceInfo.PrivateIp))
				// status.labelMacAddress.SetText(fmt.Sprintf("?????????: %s", deviceInfo.MacAddress))
				// status.labelDeviceId.SetText(fmt.Sprintf("?????? ????????????: %s", deviceInfo.DeviceId))
				// status.labelLastPoaTime.SetText(fmt.Sprintf("????????? ?????? ??????: %s", time.Unix(deviceInfo.Timestamp, 0).Format("2006-01-02 15:04")))
				// status.labelVersion.SetText(fmt.Sprintf("??????: %s", poaContext.Version))
			} else if activeContect == structureContent.content {
				structureContent.updateTreeView()
				structureContent.treeDevices.Refresh()

				structureContent.updateDetailView(structureContent.selectedDevice)
			}
		}
	}()
}

func newStatusContent() *contentStatus {
	status := contentStatus{deadDeviceOnly: false}

	// status.content = container.NewVBox()
	status.content = container.NewMax()

	status.labelStatus = widget.NewLabel("?????? ??????:")
	status.checkDeadDeviceOnly = widget.NewCheck("?????? ?????? ????????? ??????", func(check bool) {
		status.deadDeviceOnly = check
		status.listDevices.Refresh()
		// if status.deadDeviceOnly && status.selectedDevice != nil && status.selectedDevice.Alive {
		// 	status.detailContent.Hide()
		// 	status.listDevices.UnselectAll()
		// 	status.selectedDevice = nil
		// }
		status.detailContent.Hide()
		status.listDevices.UnselectAll()
		status.selectedDevice = nil
	})
	status.listDevices = widget.NewList(
		func() int {
			if status.deadDeviceOnly {
				return len(poaManager.DeadDevices)
			} else {
				return len(poaManager.TotalDevices)
			}
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(res.Ic_error), widget.NewLabel("Template Object"), layout.NewSpacer())
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			var device *manager.DeviceInfo

			if status.deadDeviceOnly {
				device = poaManager.DeadDevices[id]
			} else {
				device = poaManager.TotalDevices[id]
			}

			if device.Alive {
				item.(*fyne.Container).Objects[0].Hide()
			} else {
				item.(*fyne.Container).Objects[0].Show()
			}

			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(fmt.Sprintf("%s[%d]: %s", device.Owner, device.OwnNumber, device.DeviceDesc))
		})
	status.listDevices.OnSelected = func(id widget.ListItemID) {
		var device *manager.DeviceInfo

		if status.deadDeviceOnly {
			device = poaManager.DeadDevices[id]
		} else {
			device = poaManager.TotalDevices[id]
		}

		status.selectedDevice = device
		status.updateDetailView(device)
		status.detailContent.Show()

		// remove device button
		status.buttonRemove.OnTapped = func() {
			logger.LogD("remove device: ", device)
			if ok, _ := poaManager.RemoveDevices(device.DeviceId); ok {
				status.detailContent.Hide()
				status.listDevices.UnselectAll()
				status.selectedDevice = nil
			}
		}
	}
	status.listDevices.OnUnselected = func(id widget.ListItemID) {
	}

	status.detailContent = container.NewVBox()
	status.labelDetailID = widget.NewLabel("")
	status.labelDetailHeader = widget.NewLabel("")
	status.labelDetailData = widget.NewLabel("")
	status.buttonRemove = widget.NewButton("???????????? ??????", nil)

	status.detailContent.Add(status.labelDetailID)
	status.detailContent.Add(widget.NewSeparator())
	status.detailContent.Add(status.labelDetailHeader)
	status.detailContent.Add(status.labelDetailData)
	status.detailContent.Add(layout.NewSpacer())
	status.detailContent.Add(status.buttonRemove)
	status.detailContent.Hide()

	status.content.Add(container.NewBorder(container.NewHBox(status.labelStatus, status.checkDeadDeviceOnly), nil, nil, nil,
		container.NewHSplit(status.listDevices, status.detailContent)))

	// status.content.Add(container.NewVBox(status.labelOwner, status.labelOwnNumber))
	// status.content.Add(status.labelDesc)
	// status.content.Add(container.NewVBox(widget.NewSeparator(), status.labelPublicIp, status.labelPrivateIp, status.labelMacAddress,
	// status.labelDeviceId, status.labelLastPoaTime, layout.NewSpacer(), status.labelVersion))

	return &status
}

func (status *contentStatus) GetContent() *fyne.Container {
	return status.content
}

func (status *contentStatus) SetMainContent() {
	if parentContainer != nil {
		parentContainer.Objects = []fyne.CanvasObject{status.content}
		activeContect = status.content
	}
}

func (status *contentStatus) updateDetailView(device *manager.DeviceInfo) {
	if device == nil {
		return
	}

	device = poaManager.Devices[device.DeviceId]
	if device == nil {
		return
	}

	var aliveText string
	if device.Alive {
		aliveText = "??????"
	} else {
		aliveText = "?????? ??????"
	}

	status.labelDetailID.SetText(fmt.Sprintf("?????? ????????????: %s", device.DeviceId))
	status.labelDetailHeader.SetText(fmt.Sprintf("?????????: %s\n????????????: %d\n??????: %s", device.Owner, device.OwnNumber, device.DeviceDesc))
	status.labelDetailData.SetText(fmt.Sprintf("??????IP: %s\n??????IP: %s\n?????????: %s\n\n????????? ?????? ??????: %s\n????????????: %s\n\n??????:%s",
		device.PublicIp, device.PrivateIp, device.MacAddress, time.Unix(device.Timestamp, 0).Format("2006-01-02 15:04:05"), aliveText, device.Version))
}

func newStructureContent() *contentStructure {
	structure := contentStructure{treeData: map[string][]string{}}

	structure.content = container.NewMax()
	structure.treeDevices = &widget.Tree{
		ChildUIDs: func(uid string) (c []string) {
			c = structure.treeData[uid]
			return
		},
		IsBranch: func(uid string) (b bool) {
			_, b = structure.treeData[uid]
			return
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			label := widget.NewLabel("Template Object")
			label.Resize(fyne.Size{Height: 200})
			return container.NewHBox(widget.NewIcon(res.Ic_error), label)
		},
		UpdateNode: func(uid string, branch bool, node fyne.CanvasObject) {
			// node.(*widget.Label).SetText(uid)
			if branch {
				node.(*fyne.Container).Objects[0].(*widget.Icon).Hide()
			}

			if match, _ := regexp.MatchString("[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]", uid); match {
				node.(*fyne.Container).Objects[1].(*widget.Label).SetText(uid)
			} else {
				// device.DeviceId, owner, device.OwnNumber, desc
				nodeInfos := strings.Split(uid, " \\ ")
				owner := strings.ReplaceAll(nodeInfos[1], "\\\\", "\\")
				ownNumber := nodeInfos[2]
				desc := strings.ReplaceAll(nodeInfos[3], "\\\\", "\\")
				node.(*fyne.Container).Objects[1].(*widget.Label).SetText(fmt.Sprintf("%s[%s]: %s", owner, ownNumber, desc))

				if poaManager.Devices[nodeInfos[0]].Alive {
					node.(*fyne.Container).Objects[0].(*widget.Icon).Hide()
					// node.(*fyne.Container).Objects[0].(*widget.Icon).SetResource(res.Ic_connected)
				} else {
					node.(*fyne.Container).Objects[0].(*widget.Icon).Show()
					// node.(*fyne.Container).Objects[0].(*widget.Icon).SetResource(res.Ic_error)
				}
			}
		},
	}
	structure.treeDevices.ExtendBaseWidget(structure.treeDevices)

	structure.treeDevices.OnSelected = func(uid string) {
		logger.LogD("Tree node selected:", uid)

		if match, _ := regexp.MatchString("[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]", uid); match {
			structure.detailContent.Hide()
			structure.selectedDevice = nil
		} else {
			// nodeInfos: device.DeviceId \ owner \ device.OwnNumber \ desc
			nodeInfos := strings.Split(uid, " \\ ")
			deviceId := nodeInfos[0]
			device := poaManager.Devices[deviceId]

			structure.selectedDevice = device

			structure.detailContent.Show()
			structure.updateDetailView(device)

			// remove device button
			structure.buttonRemove.OnTapped = func() {
				logger.LogD("remove device: ", device)
				if ok, _ := poaManager.RemoveDevices(deviceId); ok {
					treeDeviceItems := structure.treeData[device.PublicIp]
					structure.treeData[device.PublicIp] = []string{}

					for _, treeDeviceItem := range treeDeviceItems {
						nodeInfos := strings.Split(treeDeviceItem, " \\ ")
						if len(nodeInfos) > 0 {
							treeNodeId := nodeInfos[0]

							if treeNodeId != device.DeviceId {
								structure.treeData[device.PublicIp] = append(structure.treeData[device.PublicIp], treeDeviceItem)
							}
						}
					}

					structure.treeDevices.Refresh()
					structure.treeDevices.UnselectAll()

					structure.detailContent.Hide()
					structure.selectedDevice = nil
				}
			}
		}
	}

	structure.treeDevices.OnUnselected = func(id string) {
		logger.LogD("Tree node unselected:", id)
	}

	structure.detailContent = container.NewVBox()
	structure.labelDetailID = widget.NewLabel("")
	structure.labelDetailHeader = widget.NewLabel("")
	structure.labelDetailData = widget.NewLabel("")
	structure.buttonRemove = widget.NewButton("???????????? ??????", nil)

	structure.detailContent.Add(structure.labelDetailID)
	structure.detailContent.Add(widget.NewSeparator())
	structure.detailContent.Add(structure.labelDetailHeader)
	// structure.detailContent.Add(widget.NewSeparator())
	structure.detailContent.Add(structure.labelDetailData)
	structure.detailContent.Add(layout.NewSpacer())
	structure.detailContent.Add(structure.buttonRemove)

	structure.content.Add(container.NewHSplit(container.NewBorder(nil, nil, nil, nil, structure.treeDevices), structure.detailContent))

	// status.content.Add(container.NewVBox(status.labelOwner, status.labelOwnNumber))
	// status.content.Add(status.labelDesc)
	// status.content.Add(container.NewVBox(widget.NewSeparator(), status.labelPublicIp, status.labelPrivateIp, status.labelMacAddress,
	// status.labelDeviceId, status.labelLastPoaTime, layout.NewSpacer(), status.labelVersion))

	return &structure
}

func (structure *contentStructure) GetContent() *fyne.Container {
	return structure.content
}

func (structure *contentStructure) SetMainContent() {
	if parentContainer != nil {
		parentContainer.Objects = []fyne.CanvasObject{structure.content}
		activeContect = structure.content
	}
}

func (structure *contentStructure) makeUid(device *manager.DeviceInfo) (uid string) {
	owner := strings.ReplaceAll(device.Owner, "\\", "\\\\")
	desc := strings.ReplaceAll(device.DeviceDesc, "\\", "\\\\")
	uid = fmt.Sprintf("%s \\ %s \\ %d \\ %s", device.DeviceId, owner, device.OwnNumber, desc)

	return
}

func (structure *contentStructure) updateTreeView() {
	structure.treeData = map[string][]string{}

	publicIps := map[string]manager.DeviceInfo{}
	for _, device := range poaManager.TotalDevices {
		publicIps[device.PublicIp] = *device
	}

	roots := []string{}
	for publicIp := range publicIps {
		roots = append(roots, publicIp)

		for _, device := range poaManager.TotalDevices {
			if publicIp == device.PublicIp {
				structure.treeData[publicIp] = append(structure.treeData[publicIp], structure.makeUid(device))
			}
		}
	}

	sort.Slice(roots, func(i, j int) bool {
		return roots[i] < roots[j]
	})
	structure.treeData[""] = roots
}

func (structure *contentStructure) updateDetailView(device *manager.DeviceInfo) {
	if device == nil {
		return
	}

	device = poaManager.Devices[device.DeviceId]
	if device == nil {
		return
	}

	var aliveText string
	if device.Alive {
		aliveText = "??????"
	} else {
		aliveText = "?????? ??????"
	}

	structure.labelDetailID.SetText(fmt.Sprintf("?????? ????????????: %s", device.DeviceId))
	structure.labelDetailHeader.SetText(fmt.Sprintf("?????????: %s\n????????????: %d\n??????: %s", device.Owner, device.OwnNumber, device.DeviceDesc))
	structure.labelDetailData.SetText(fmt.Sprintf("??????IP: %s\n??????IP: %s\n?????????: %s\n\n????????? ?????? ??????: %s\n????????????: %s\n\n??????:%s",
		device.PublicIp, device.PrivateIp, device.MacAddress, time.Unix(device.Timestamp, 0).Format("2006-01-02 15:04:04"), aliveText, device.Version))

	structure.treeDevices.Select(structure.makeUid(device))
}

func newCommandDeviceControl() *contentDeviceControl {
	deviceControl := contentDeviceControl{}

	deviceControl.content = container.NewMax()

	deviceControl.buttonMqttUserPwd = widget.NewButton("MQTT ?????????/?????? ??????", func() {
		content := container.NewVBox()
		labelMessage := widget.NewLabel("????????? MQTT ?????? ????????? ????????? ?????????.\nMQTT ????????? ?????? ?????? ?????? ????????? ?????? ????????? ????????? ??? ????????????.")
		entryUser := widget.NewEntry()
		entryPassword := widget.NewEntry()
		content.Add(labelMessage)
		content.Add(entryUser)
		content.Add(entryPassword)

		dialog.ShowCustomConfirm("MQTT ?????? ?????? ??????", "??????", "??????", content,
			func(ok bool) {
				if ok && entryUser.Text != "" && entryPassword.Text != "" {
					poaContext.EventLooper.PushEvent(event.MANAGER, event.EVENT_MANAGER_DEVICE_MQTT_CHANGE_USER_PASSWORD, entryUser.Text, entryPassword.Text)
				}
			}, *window)
	})

	deviceControl.buttonUpdateAddress = widget.NewButton("???????????? ?????? ??????", func() {
		content := container.NewVBox()
		labelMessage := widget.NewLabel("???????????? ????????? ????????? ?????????.")
		entryServerAddress := widget.NewEntry()
		content.Add(labelMessage)
		content.Add(entryServerAddress)

		customDialog := dialog.NewCustomConfirm("???????????? ?????? ??????", "??????", "??????", content,
			func(ok bool) {
				if ok && entryServerAddress.Text != "" {
					poaContext.EventLooper.PushEvent(event.MANAGER, event.EVENT_MANAGER_DEVICE_CHANGE_UPDATE_ADDRESS, entryServerAddress.Text)
				}
			}, *window)
		customDialog.Resize(fyne.Size{Width: 640})
		customDialog.Show()
	})

	deviceControl.buttonForceUpdate = widget.NewButton("???????????? ?????? ??????", func() {
		poaContext.EventLooper.PushEvent(event.MANAGER, event.EVENT_MANAGER_DEVICE_FORCE_UPDATE)
		dialog.ShowInformation("???????????? ?????? ??????", "???????????? ????????? ??????????????????.", *window)
	})

	deviceControl.buttonForceRestart = widget.NewButton("?????????????????? ?????????", func() {
		content := container.NewVBox()
		labelMessage := widget.NewLabel("?????????????????? ???????????? ?????????????????????????.")
		content.Add(labelMessage)

		dialog.ShowCustomConfirm("?????????????????? ????????? ??????", "??????", "??????", content,
			func(ok bool) {
				if ok {
					poaContext.EventLooper.PushEvent(event.MANAGER, event.EVENT_MANAGER_DEVICE_RESTART)
				}
			}, *window)
	})

	deviceControl.content.Add(container.NewHBox(layout.NewSpacer(), container.NewVBox(deviceControl.buttonMqttUserPwd, deviceControl.buttonUpdateAddress, deviceControl.buttonForceUpdate, deviceControl.buttonForceRestart), layout.NewSpacer()))

	return &deviceControl
}

func (deviceControl *contentDeviceControl) GetContent() *fyne.Container {
	return deviceControl.content
}

func (deviceControl *contentDeviceControl) SetMainContent() {
	if parentContainer != nil {
		parentContainer.Objects = []fyne.CanvasObject{deviceControl.content}
		activeContect = deviceControl.content
	}
}

func newConfigContent() *contentConfig {
	config := contentConfig{}

	config.content = container.NewPadded()

	config.serverAddressEntry = widget.NewEntry()
	config.serverPortEntry = NewNumericalEntry()
	config.mqttAddressEntry = widget.NewEntry()
	config.mqttPortEntry = NewNumericalEntry()
	config.mqttUserEntry = widget.NewEntry()
	config.mqttPasswordEntry = widget.NewPasswordEntry()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "?????? ??????", Widget: config.serverAddressEntry},
			{Text: "?????? ??????", Widget: config.serverPortEntry},
			{Text: "MQTT ??????", Widget: config.mqttAddressEntry},
			{Text: "MQTT ??????", Widget: config.mqttPortEntry},
			{Text: "MQTT ?????????", Widget: config.mqttUserEntry},
			{Text: "MQTT ????????????", Widget: config.mqttPasswordEntry},
		},
		OnSubmit: func() {
			oldConfigs := poaContext.Configs

			poaContext.Configs.PoaServerAddress = config.serverAddressEntry.Text
			poaContext.Configs.PoaServerPort, _ = strconv.Atoi(config.serverPortEntry.Text)
			poaContext.Configs.MqttBrokerAddress = config.mqttAddressEntry.Text
			poaContext.Configs.MqttPort, _ = strconv.Atoi(config.mqttPortEntry.Text)
			poaContext.Configs.MqttUser = config.mqttUserEntry.Text
			poaContext.Configs.MqttPassword = config.mqttPasswordEntry.Text
			poaContext.WriteConfig()

			if oldConfigs != poaContext.Configs {
				dialog.ShowInformation("?????? ?????? ??????", "?????? ????????? ??????????????? ??????????????? ????????? ????????????.", *window)
			}
		},
		SubmitText: "??????",
	}

	config.content.Add(form)

	return &config
}

func (config *contentConfig) GetContent() *fyne.Container {
	return config.content
}

func (config *contentConfig) SetMainContent() {
	if parentContainer != nil {
		parentContainer.Objects = []fyne.CanvasObject{config.content}
		activeContect = config.content
	}
}
