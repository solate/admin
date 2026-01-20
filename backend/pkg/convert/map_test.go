package convert

import (
	"testing"
)

type testItem struct {
	ID   string
	Name string
}

func TestToMap(t *testing.T) {
	items := []testItem{
		{ID: "1", Name: "First"},
		{ID: "2", Name: "Second"},
		{ID: "3", Name: "Third"},
	}

	result := ToMap(items, func(item testItem) string {
		return item.ID
	})

	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}

	if result["1"].Name != "First" {
		t.Errorf("expected First, got %s", result["1"].Name)
	}

	if result["2"].Name != "Second" {
		t.Errorf("expected Second, got %s", result["2"].Name)
	}

	if result["3"].Name != "Third" {
		t.Errorf("expected Third, got %s", result["3"].Name)
	}
}

func TestToMapEmpty(t *testing.T) {
	items := []testItem{}

	result := ToMap(items, func(item testItem) string {
		return item.ID
	})

	if len(result) != 0 {
		t.Errorf("expected empty map, got %d items", len(result))
	}
}

func TestToMapIntKey(t *testing.T) {
	items := []testItem{
		{ID: "a", Name: "Apple"},
		{ID: "b", Name: "Banana"},
	}

	type pair struct {
		idx  int
		item testItem
	}

	pairs := []pair{
		{idx: 1, item: items[0]},
		{idx: 2, item: items[1]},
	}

	result := ToMap(pairs, func(p pair) int {
		return p.idx
	})

	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}

	if result[1].item.Name != "Apple" {
		t.Errorf("expected Apple at key 1, got %s", result[1].item.Name)
	}
}

func TestToMapDuplicateKey(t *testing.T) {
	// 测试重复键的情况（后面的值会覆盖前面的）
	items := []testItem{
		{ID: "1", Name: "First"},
		{ID: "1", Name: "Second"}, // 相同的 ID
	}

	result := ToMap(items, func(item testItem) string {
		return item.ID
	})

	if len(result) != 1 {
		t.Errorf("expected 1 item (duplicate keys), got %d", len(result))
	}

	// 后面的值应该覆盖前面的
	if result["1"].Name != "Second" {
		t.Errorf("expected Second (last value wins), got %s", result["1"].Name)
	}
}
