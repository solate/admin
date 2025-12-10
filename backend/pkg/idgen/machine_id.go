package idgen

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"net"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// GetMachineID 根据网络信息生成唯一的机器ID (当机器数量达到 300 台时，冲突概率就超过 50%)
func GetMachineID() (uint16, error) {

	// 1. 尝试从环境变量读取
	if envID := os.Getenv("MACHINE_ID"); envID != "" {
		if id, err := strconv.ParseUint(envID, 10, 16); err == nil {
			return uint16(id), nil
		}
	}

	// 2. 使用系统提供的machine-id（Linux）
	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/etc/machine-id"); err == nil && len(data) >= 16 {
			// 使用前16个字节的哈希值
			hash := sha256.Sum256(data)
			machineID := binary.BigEndian.Uint32(hash[:4])
			return uint16(machineID & 0xFFFF), nil
		}
	}

	// 3. 使用网络接口的MAC地址组合
	if machineID, err := getMacBasedID(); err == nil {
		return machineID, nil
	}

	return 0, errors.New("unable to generate machine ID")

}

func getMacBasedID() (uint16, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return 0, err
	}

	var macs []string
	for _, iface := range interfaces {
		if len(iface.HardwareAddr) > 0 && (iface.Flags&net.FlagUp) != 0 && (iface.Flags&net.FlagLoopback) == 0 {
			macs = append(macs, iface.HardwareAddr.String())
		}
	}

	if len(macs) == 0 {
		return 0, errors.New("no valid MAC address")
	}

	// 对MAC地址进行排序并组合
	sort.Strings(macs)
	combined := strings.Join(macs, "")

	// 使用哈希函数
	hash := sha256.Sum256([]byte(combined))
	machineID := binary.BigEndian.Uint32(hash[:4])

	return uint16(machineID & 0xFFFF), nil
}
