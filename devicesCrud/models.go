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

type SmartHomeDevicePatch struct {
	DeviceID   string
	DeviceName string
}

type ProblemDetail struct {
	ErrorType string
	Title     string
	Status    int
	Detail    string
}
