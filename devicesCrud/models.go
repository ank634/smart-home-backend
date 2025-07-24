package devicesCrud

type SmartHomeDevice struct {
	DeviceID    *string
	DeviceName  *string
	DeviceType  *string
	ServiceType *string
	Manufactor  *string
	SetTopic    *string
	GetTopic    *string
	DeviceUrl   *string
}

type LightDevice struct {
	DeviceID    *string
	DeviceName  *string
	DeviceType  *string
	ServiceType *string
	Manufactor  *string
	SetTopic    *string
	GetTopic    *string
	EndPoint    *string
	RoomID      *int

	IsDimmable *bool
	IsRgb      *bool
}

func newLightDevice(id string, name string, deviceType string,
	serviceType string, manufactor string, setTopic string,
	getTopic string, endpoint string, roomId *int,
	isDimmable bool, isRgb bool) *LightDevice {
	if roomId == nil {
		return &LightDevice{&id, &name, &deviceType,
			&serviceType, &manufactor, &setTopic,
			&getTopic, &endpoint, roomId, &isDimmable, &isRgb}
	}
	return &LightDevice{&id, &name, &deviceType,
		&serviceType, &manufactor, &setTopic,
		&getTopic, &endpoint, roomId, &isDimmable, &isRgb}

}

func equalStrings(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a != nil && b != nil {
		return *a == *b
	}
	return false
}

func equalInts(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a != nil && b != nil {
		return *a == *b
	}
	return false
}

func equalBools(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a != nil && b != nil {
		return *a == *b
	}
	return false
}

func EqualLightDevices(a, b *LightDevice) bool {
	if a == nil || b == nil {
		return a == b // true if both nil
	}

	return equalStrings(a.DeviceID, b.DeviceID) &&
		equalStrings(a.DeviceName, b.DeviceName) &&
		equalStrings(a.DeviceType, b.DeviceType) &&
		equalStrings(a.ServiceType, b.ServiceType) &&
		equalStrings(a.Manufactor, b.Manufactor) &&
		equalStrings(a.SetTopic, b.SetTopic) &&
		equalStrings(a.GetTopic, b.GetTopic) &&
		equalStrings(a.EndPoint, b.EndPoint) &&
		equalInts(a.RoomID, b.RoomID) &&
		equalBools(a.IsDimmable, b.IsDimmable) &&
		equalBools(a.IsRgb, b.IsRgb)
}

type SmartHomeDevicePatch struct {
	DeviceID   string
	DeviceName string
}
