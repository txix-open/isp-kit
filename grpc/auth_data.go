package grpc

import (
	"strconv"

	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

const (
	ApplicationIdHeader = "x-application-identity"
	UserIdHeader        = "x-user-identity"
	DeviceIdHeader      = "x-device-identity"
	ServiceIdHeader     = "x-service-identity"
	DomainIdHeader      = "x-domain-identity"
	SystemIdHeader      = "x-system-identity"
	UserTokenHeader     = "x-user-token"
	DeviceTokenHeader   = "x-device-token"
)

type AuthData metadata.MD

func (i AuthData) SystemId() (int, error) {
	return intFromMd(SystemIdHeader, metadata.MD(i))
}

func (i AuthData) DomainId() (int, error) {
	return intFromMd(DomainIdHeader, metadata.MD(i))
}

func (i AuthData) ServiceId() (int, error) {
	return intFromMd(ServiceIdHeader, metadata.MD(i))
}

func (i AuthData) ApplicationId() (int, error) {
	return intFromMd(ApplicationIdHeader, metadata.MD(i))
}

func (i AuthData) UserId() (int, error) {
	return intFromMd(UserIdHeader, metadata.MD(i))
}

func (i AuthData) DeviceId() (int, error) {
	return intFromMd(DeviceIdHeader, metadata.MD(i))
}

func (i AuthData) UserToken() (string, error) {
	return StringFromMd(UserTokenHeader, metadata.MD(i))
}

func (i AuthData) DeviceToken() (string, error) {
	return StringFromMd(DeviceTokenHeader, metadata.MD(i))
}

func StringFromMd(key string, md metadata.MD) (string, error) {
	values := md[key]
	if len(values) == 0 {
		return "", errors.Errorf("'%s' is expected id metadata", key)
	}
	return values[0], nil
}

func intFromMd(key string, md metadata.MD) (int, error) {
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
