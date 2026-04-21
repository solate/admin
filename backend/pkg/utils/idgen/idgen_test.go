package idgen

import (
	"testing"
)

func TestGenerateUUID_Success(t *testing.T) {
	uuid, err := GenerateUUID()
	if err != nil {
		t.Errorf("GenerateUUID returned an error: %v", err)
	}
	if uuid == "" {
		t.Errorf("GenerateUUID returned a zero UUID")
	}
}

func TestGenerateUUIDs_Success(t *testing.T) {
	n := 10
	uuids, err := GenerateUUIDs(n)
	if err != nil {
		t.Errorf("GenerateUUIDs returned an error: %v", err)
	}
	if len(uuids) != n {
		t.Errorf("GenerateUUIDs returned %d UUIDs, expected %d", len(uuids), n)
	}
	for _, uuid := range uuids {
		if uuid == "" {
			t.Errorf("GenerateUUIDs returned a zero UUID")
		}
		t.Log(uuid)
	}
}

func TestGenerateUUIDs_ErrorPropagation(t *testing.T) {
	// 模拟 GenerateUUID 返回错误
	// 这里我们假设有一个方法可以模拟错误，但实际实现中没有这样的方法。
	// 因此，我们无法直接测试错误传播，除非修改 GenerateUUID 的行为。
	// 为了演示，我们假设有一个方法可以模拟错误。
	// uuids, err := GenerateUUIDs(1)
	// if err == nil {
	// 	t.Errorf("GenerateUUIDs should have returned an error")
	// }
}
