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
	"poa-manager/manager"
	"poa-manager/res"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

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
		"status":        {"전체상태", "등록된 장치들의 현재 상태를 표시합니다.", statusContent},
		"structure":     {"네트워크별 보기", "등록된 장치들의 목록을 표시합니다.", structureContent},
		"deviceControl": {"장치 제어", "모든 장치에게 명령 메시지를 전송합니다..", deviceControlContent},
		"configs":       {"설정", "매니저 환경 설정을 할 수 있습니다.", configContent},
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
				statusContent.labelStatus.SetText(fmt.Sprintf("전체: %d 대, 정상: %d 대, 응답 없음: %d 대", totalCount, totalCount-deadCount, deadCount))

				statusContent.listDevices.Refresh()
				statusContent.updateDetailView(statusContent.selectedDevice)

				// status.labelOwner.SetText(fmt.Sprintf("사용자: %s", deviceInfo.Owner))
				// status.labelOwnNumber.SetText(fmt.Sprintf("장치 번호: %d", deviceInfo.OwnNumber))
				// status.labelDesc.SetText(fmt.Sprintf("설명: %s", deviceInfo.DeviceDesc))
				// status.labelPublicIp.SetText(fmt.Sprintf("공인IP: %s", deviceInfo.PublicIp))
				// status.labelPrivateIp.SetText(fmt.Sprintf("내부IP: %s", deviceInfo.PrivateIp))
				// status.labelMacAddress.SetText(fmt.Sprintf("맥주소: %s", deviceInfo.MacAddress))
				// status.labelDeviceId.SetText(fmt.Sprintf("장치 고유번호: %s", deviceInfo.DeviceId))
				// status.labelLastPoaTime.SetText(fmt.Sprintf("마지막 통신 시간: %s", time.Unix(deviceInfo.Timestamp, 0).Format("2006-01-02 15:04")))
				// status.labelVersion.SetText(fmt.Sprintf("버전: %s", poaContext.Version))
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

	status.labelStatus = widget.NewLabel("장치 정보:")
	status.checkDeadDeviceOnly = widget.NewCheck("응답 없는 장치만 보기", func(check bool) {
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

			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(fmt.Sprintf("%s[%d] %s", device.Owner, device.OwnNumber, device.DeviceDesc))
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
			fmt.Println("remove device: ", device)
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
	status.buttonRemove = widget.NewButton("목록에서 제거", nil)

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
		aliveText = "정상"
	} else {
		aliveText = "응답 없음"
	}

	status.labelDetailID.SetText(fmt.Sprintf("장치 고유번호: %s", device.DeviceId))
	status.labelDetailHeader.SetText(fmt.Sprintf("사용자: %s\n장치번호: %d\n설명: %s", device.Owner, device.OwnNumber, device.DeviceDesc))
	status.labelDetailData.SetText(fmt.Sprintf("공인IP: %s\n내부IP: %s\n맥주소: %s\n\n마지막 통신 시간: %s\n통신상태: %s\n\n버전:%s",
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
				node.(*fyne.Container).Objects[1].(*widget.Label).SetText(fmt.Sprintf("%s[%s] %s", owner, ownNumber, desc))

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
		fmt.Println("Tree node selected:", uid)

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
				fmt.Println("remove device: ", device)
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
		fmt.Println("Tree node unselected:", id)
	}

	structure.detailContent = container.NewVBox()
	structure.labelDetailID = widget.NewLabel("")
	structure.labelDetailHeader = widget.NewLabel("")
	structure.labelDetailData = widget.NewLabel("")
	structure.buttonRemove = widget.NewButton("목록에서 제거", nil)

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
		aliveText = "정상"
	} else {
		aliveText = "응답 없음"
	}

	structure.labelDetailID.SetText(fmt.Sprintf("장치 고유번호: %s", device.DeviceId))
	structure.labelDetailHeader.SetText(fmt.Sprintf("사용자: %s\n장치번호: %d\n설명: %s", device.Owner, device.OwnNumber, device.DeviceDesc))
	structure.labelDetailData.SetText(fmt.Sprintf("공인IP: %s\n내부IP: %s\n맥주소: %s\n\n마지막 통신 시간: %s\n통신상태: %s\n\n버전:%s",
		device.PublicIp, device.PrivateIp, device.MacAddress, time.Unix(device.Timestamp, 0).Format("2006-01-02 15:04:04"), aliveText, device.Version))

	structure.treeDevices.Select(structure.makeUid(device))
}

//TODO:
func newCommandDeviceControl() *contentDeviceControl {
	deviceControl := contentDeviceControl{}

	deviceControl.content = container.NewMax()

	deviceControl.buttonMqttUserPwd = widget.NewButton("MQTT 아이디/비번 설정", func() {
		content := container.NewVBox()
		labelMessage := widget.NewLabel("변경할 MQTT 계정 정보를 입력해 주세요.\nMQTT 정보가 틀릴 경우 모든 장치의 서버 접속이 제한될 수 있습니다.")
		entryUser := widget.NewEntry()
		entryPassword := widget.NewEntry()
		content.Add(labelMessage)
		content.Add(entryUser)
		content.Add(entryPassword)

		dialog.ShowCustomConfirm("MQTT 계정 정보 변경", "확인", "취소", content,
			func(ok bool) {
				if ok && entryUser.Text != "" && entryPassword.Text != "" {
					poaContext.EventLooper.PushEvent(event.MANAGER, event.EVENT_MANAGER_DEVICE_MQTT_CHANGE_USER_PASSWORD, entryUser.Text, entryPassword.Text)
				}
			}, *window)
	})

	deviceControl.buttonUpdateAddress = widget.NewButton("업데이트 서버 설정", func() {
		content := container.NewVBox()
		labelMessage := widget.NewLabel("업데이트 주소를 입력해 주세요.")
		entryServerAddress := widget.NewEntry()
		content.Add(labelMessage)
		content.Add(entryServerAddress)

		customDialog := dialog.NewCustomConfirm("업데이트 주소 설정", "확인", "취소", content,
			func(ok bool) {
				if ok && entryServerAddress.Text != "" {
					poaContext.EventLooper.PushEvent(event.MANAGER, event.EVENT_MANAGER_DEVICE_CHANGE_UPDATE_ADDRESS, entryServerAddress.Text)
				}
			}, *window)
		customDialog.Resize(fyne.Size{Width: 640})
		customDialog.Show()
	})

	deviceControl.buttonForceUpdate = widget.NewButton("업데이트 확인 요청", func() {
		poaContext.EventLooper.PushEvent(event.MANAGER, event.EVENT_MANAGER_DEVICE_FORCE_UPDATE)
		dialog.ShowInformation("업데이트 확인 요청", "업데이트 확인을 요청했습니다.", *window)
	})

	deviceControl.buttonForceRestart = widget.NewButton("어플리케이션 재시작", func() {
		content := container.NewVBox()
		labelMessage := widget.NewLabel("어플리케이션 재시작을 요청하시겠습니까?.")
		content.Add(labelMessage)

		dialog.ShowCustomConfirm("어플리케이션 재시작 요청", "확인", "취소", content,
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

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "서버 주소", Widget: config.serverAddressEntry},
			{Text: "서버 포트", Widget: config.serverPortEntry},
			{Text: "MQTT 주소", Widget: config.mqttAddressEntry},
			{Text: "MQTT 포트", Widget: config.mqttPortEntry}},
		OnSubmit: func() {
			oldConfigs := poaContext.Configs

			poaContext.Configs.PoaServerAddress = config.serverAddressEntry.Text
			poaContext.Configs.PoaServerPort, _ = strconv.Atoi(config.serverPortEntry.Text)
			poaContext.Configs.MqttBrokerAddress = config.mqttAddressEntry.Text
			poaContext.Configs.MqttPort, _ = strconv.Atoi(config.mqttPortEntry.Text)
			poaContext.WriteConfig()

			if oldConfigs != poaContext.Configs {
				dialog.ShowInformation("접속 정보 변경", "수정 사항을 적용하려면 프로그램을 재시작 해주세요.", *window)
			}
		},
		SubmitText: "저장",
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
