package event

type EventTarget int
type EventName int

const (
	MAIN EventTarget = iota
	MANAGER
)

const (
	EVENT_MANAGER_DEVICE_RESTART EventName = iota
	EVENT_MANAGER_DEVICE_MQTT_CHANGE_USER_PASSWORD
	EVENT_MANAGER_DEVICE_FORCE_UPDATE
	EVENT_MANAGER_DEVICE_CHANGE_UPDATE_ADDRESS
)
