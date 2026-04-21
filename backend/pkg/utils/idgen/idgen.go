package idgen

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/sony/sonyflake"
)

var (
	sf     *sonyflake.Sonyflake
	sfOnce sync.Once
)

// GetSonyflake 返回一个单例的 Sonyflake 实例
func GetSonyflake() (*sonyflake.Sonyflake, error) {
	var err error
	sfOnce.Do(func() {
		settings := sonyflake.Settings{
			StartTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			MachineID: GetMachineID,
		}
		sf = sonyflake.NewSonyflake(settings)
		if sf == nil {
			err = errors.New("failed to create sonyflake instance")
		}
	})
	return sf, err
}

// GenerateUUID 生成一个 UUID, 前端超过10位会截断, 所以转换成string来处理
func GenerateUUID() (string, error) {
	sf, err := GetSonyflake()
	if err != nil {
		return "", err
	}
	id, err := sf.NextID()
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(id, 10), nil
}

// GenerateUUIDs 生成多个 UUIDs
func GenerateUUIDs(n int) ([]string, error) {
	var uuids []string
	for i := 0; i < n; i++ {
		uuid, err := GenerateUUID()
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, uuid)
	}
	return uuids, nil
}
