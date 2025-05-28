package contextkeys

type contextKey string

const (
	UserKey   contextKey = "user"
	DeviceKey contextKey = "device"

	DeviceIDKey   contextKey = "deviceID"
	DeviceNameKey contextKey = "deviceName"
)
