package context

import "context"

type (
	Key string
)

const (
	ClientVersion Key = "client_version"
	DeviceId      Key = "device_id"
)

func GetStringValue(ctx context.Context, key Key) string {
	value := ctx.Value(key)
	if value == "" {
		return ""
	}
	if strValue, ok := value.(string); ok {
		return strValue
	}

	return ""
}
