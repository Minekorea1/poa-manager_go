package ui

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"poa-manager/context"
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

	statusContent    *contentStatus
	structureContent *contentStructure
	configContent    *contentConfig
)

type contentStatus struct {
	content     *fyne.Container
	labelStatus *widget.Label
	listDevices *widget.List
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
	configContent = newConfigContent()

	menus = map[string]Menu{
		"status":    {"상태", "등록된 장치들의 현재 상태를 표시합니다.", statusContent},
		"structure": {"구조", "등록된 장치들의 목록을 표시합니다.", structureContent},
		"configs":   {"설정", "매니저 환경 설정을 할 수 있습니다.", configContent},
	}

	menuIndex = map[string][]string{
		"": {"status", "structure", "configs"},
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

				if activeContect == configContent.content {
					configContent.serverAddressEntry.SetText(poaContext.Configs.PoaServerAddress)
					configContent.serverPortEntry.SetText(strconv.FormatInt(int64(poaContext.Configs.PoaServerPort), 10))
					configContent.mqttAddressEntry.SetText(poaContext.Configs.MqttBrokerAddress)
					configContent.mqttPortEntry.SetText(strconv.FormatInt(int64(poaContext.Configs.MqttPort), 10))
				} else if activeContect == structureContent.content {
					// structureContent.treeData = map[string][]string{}

					// publicIps := map[string]manager.DeviceInfo{}
					// for _, device := range poaManager.TotalDevices {
					// 	publicIps[device.PublicIp] = *device
					// }

					// roots := []string{}
					// for publicIp := range publicIps {
					// 	roots = append(roots, publicIp)

					// 	for _, device := range poaManager.TotalDevices {
					// 		if publicIp == device.PublicIp {
					// 			structureContent.treeData[publicIp] = append(structureContent.treeData[publicIp], structureContent.makeUid(device))
					// 		}
					// 	}
					// }

					// sort.Slice(roots, func(i, j int) bool {
					// 	return roots[i] < roots[j]
					// })
					// structureContent.treeData[""] = roots

					structureContent.updateTreeView()
					structureContent.treeDevices.OpenAllBranches()

					structureContent.treeDevices.UnselectAll()
					structureContent.detailContent.Hide()
					structureContent.selectedDevice = nil
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

			if activeContect == structureContent.content {
				structureContent.updateTreeView()
				structureContent.treeDevices.Refresh()

				structureContent.updateDetailView(structureContent.selectedDevice)
			}
		}
	}()
}

func newStatusContent() *contentStatus {
	status := contentStatus{}

	// status.content = container.NewVBox()
	status.content = container.NewMax()

	status.labelStatus = widget.NewLabel("장치 정보:")
	status.listDevices = widget.NewList(
		func() int { return len(poaManager.TotalDevices) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(res.Ic_main), widget.NewLabel("Template Object"), layout.NewSpacer(),
				widget.NewButton("삭제", nil))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(poaManager.TotalDevices[id].Owner)
			item.(*fyne.Container).Objects[3].(*widget.Button).OnTapped = func() { fmt.Println("tab: ", id) }
		})
	status.listDevices.OnSelected = func(id widget.ListItemID) {
		fmt.Println("item selected")
	}
	status.listDevices.OnUnselected = func(id widget.ListItemID) {
		fmt.Println("Select An Item From The List")
	}

	status.content.Add(container.NewBorder(status.labelStatus, nil, nil, nil, status.listDevices))

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
	structure.labelDetailData.SetText(fmt.Sprintf("공인IP: %s\n내부IP: %s\n맥주소: %s\n\n마지막 통신 시간: %s\n통신상태: %s",
		device.PublicIp, device.PrivateIp, device.MacAddress, time.Unix(device.Timestamp, 0).Format("2006-01-02 15:04"), aliveText))

	structure.treeDevices.Select(structure.makeUid(device))
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

func Refresh() {
	if activeContect == statusContent.content {
		totalCount := len(poaManager.TotalDevices)
		deadCount := len(poaManager.DeadDevices)
		statusContent.labelStatus.SetText(fmt.Sprintf("전체: %d 대, 정상: %d 대, 응답 없음: %d 대", totalCount, totalCount-deadCount, deadCount))

		// status.labelOwner.SetText(fmt.Sprintf("사용자: %s", deviceInfo.Owner))
		// status.labelOwnNumber.SetText(fmt.Sprintf("장치 번호: %d", deviceInfo.OwnNumber))
		// status.labelDesc.SetText(fmt.Sprintf("설명: %s", deviceInfo.DeviceDesc))
		// status.labelPublicIp.SetText(fmt.Sprintf("공인IP: %s", deviceInfo.PublicIp))
		// status.labelPrivateIp.SetText(fmt.Sprintf("내부IP: %s", deviceInfo.PrivateIp))
		// status.labelMacAddress.SetText(fmt.Sprintf("맥주소: %s", deviceInfo.MacAddress))
		// status.labelDeviceId.SetText(fmt.Sprintf("장치 고유번호: %s", deviceInfo.DeviceId))
		// status.labelLastPoaTime.SetText(fmt.Sprintf("마지막 통신 시간: %s", time.Unix(deviceInfo.Timestamp, 0).Format("2006-01-02 15:04")))
		// status.labelVersion.SetText(fmt.Sprintf("버전: %s", poaContext.Version))
	}
}
