package grpc

import (
	"strconv"

	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

const (
	// ApplicationIdHeader is the metadata key for application identity.
	ApplicationIdHeader = "x-application-identity"
	// ApplicationNameHeader is the metadata key for application name.
	ApplicationNameHeader = "x-application-name"
	// UserIdHeader is the metadata key for user identity.
	UserIdHeader = "x-user-identity"
	// DeviceIdHeader is the metadata key for device identity.
	DeviceIdHeader = "x-device-identity"
	// ServiceIdHeader is the metadata key for service identity.
	ServiceIdHeader = "x-service-identity"
	// DomainIdHeader is the metadata key for domain identity.
	DomainIdHeader = "x-domain-identity"
	// SystemIdHeader is the metadata key for system identity.
	SystemIdHeader = "x-system-identity"
	// UserTokenHeader is the metadata key for user token.
	UserTokenHeader = "x-user-token"
	// DeviceTokenHeader is the metadata key for device token.
	DeviceTokenHeader = "x-device-token"
)

// AuthData wraps metadata.MD to provide structured access to authentication
// and authorization information from gRPC request metadata.
// All methods return an error if the key is not found or cannot be parsed.
type AuthData metadata.MD

// SystemId extracts the system identity from metadata as an integer.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func (i AuthData) SystemId() (int, error) {
	return IntFromMd(SystemIdHeader, metadata.MD(i))
}

// DomainId extracts the domain identity from metadata as an integer.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func (i AuthData) DomainId() (int, error) {
	return IntFromMd(DomainIdHeader, metadata.MD(i))
}

// ServiceId extracts the service identity from metadata as an integer.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func (i AuthData) ServiceId() (int, error) {
	return IntFromMd(ServiceIdHeader, metadata.MD(i))
}

// ApplicationId extracts the application identity from metadata as an integer.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func (i AuthData) ApplicationId() (int, error) {
	return IntFromMd(ApplicationIdHeader, metadata.MD(i))
}

// ApplicationName extracts the application name from metadata as a string.
// Returns an error if the key is not found.
func (i AuthData) ApplicationName() (string, error) {
	return StringFromMd(ApplicationNameHeader, metadata.MD(i))
}

// UserId extracts the user identity from metadata as an integer.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func (i AuthData) UserId() (int, error) {
	return IntFromMd(UserIdHeader, metadata.MD(i))
}

// DeviceId extracts the device identity from metadata as an integer.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func (i AuthData) DeviceId() (int, error) {
	return IntFromMd(DeviceIdHeader, metadata.MD(i))
}

// UserToken extracts the user token from metadata as a string.
// Returns an error if the key is not found.
func (i AuthData) UserToken() (string, error) {
	return StringFromMd(UserTokenHeader, metadata.MD(i))
}

// DeviceToken extracts the device token from metadata as a string.
// Returns an error if the key is not found.
func (i AuthData) DeviceToken() (string, error) {
	return StringFromMd(DeviceTokenHeader, metadata.MD(i))
}

// StringFromMd extracts a string value from metadata by key.
// Returns an error if metadata is nil or the key is not found.
func StringFromMd(key string, md metadata.MD) (string, error) {
	if md == nil {
		return "", errors.New("metadata is nil")
	}
	values := md[key]
	if len(values) == 0 {
		return "", errors.Errorf("'%s' is expected in metadata", key)
	}
	return values[0], nil
}

// IntFromMd extracts an integer value from metadata by key.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func IntFromMd(key string, md metadata.MD) (int, error) {
	value, err := StringFromMd(key, md)
	if err != nil {
		return 0, err
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.WithMessagef(err, "parse '%s' to int", key)
	}
	return intValue, nil
}
