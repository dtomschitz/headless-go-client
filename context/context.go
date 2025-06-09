package context

import "context"

type (
	Key string
)

const (
	ServiceKey       Key = "service"
	ClientVersionKey Key = "x-client-version"
	DeviceIdKey      Key = "x-device-id"
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
