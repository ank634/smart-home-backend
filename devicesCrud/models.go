package devicesCrud

type SmartHomeDevice struct {
	DeviceID    string
	DeviceName  string
	DeviceType  string
	ServiceType string
	SetTopic    string
	GetTopic    string
	DeviceUrl   string
}

type SmartHomeDevicePatch struct {
	DeviceID   string
	DeviceName string
}
