package convert

import (
	"testing"
)

type testUser struct {
	ID   string
	Name string
	Age  int
}

func TestToMap(t *testing.T) {
	users := []testUser{
		{ID: "1", Name: "Alice", Age: 25},
		{ID: "2", Name: "Bob", Age: 30},
		{ID: "3", Name: "Charlie", Age: 35},
	}

	result := ToMap(users, func(u testUser) string {
		return u.ID
	})

	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}

	if result["1"].Name != "Alice" {
		t.Errorf("expected Alice, got %s", result["1"].Name)
	}

	if result["2"].Age != 30 {
		t.Errorf("expected age 30, got %d", result["2"].Age)
	}
}

func TestToMapDuplicateKey(t *testing.T) {
	// 键重复时后者覆盖前者
	users := []testUser{
		{ID: "1", Name: "First", Age: 25},
		{ID: "1", Name: "Second", Age: 30},
	}

	result := ToMap(users, func(u testUser) string {
		return u.ID
	})

	if len(result) != 1 {
		t.Errorf("expected 1 item (duplicate keys), got %d", len(result))
	}

	if result["1"].Name != "Second" {
		t.Errorf("expected Second (last value wins), got %s", result["1"].Name)
	}
}

func TestToMapPtr(t *testing.T) {
	users := []testUser{
		{ID: "1", Name: "Alice", Age: 25},
		{ID: "2", Name: "Bob", Age: 30},
	}

	result := ToMapPtr(users, func(u testUser) string {
		return u.ID
	})

	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}

	if result["1"].Name != "Alice" {
		t.Errorf("expected Alice, got %s", result["1"].Name)
	}

	// 修改通过指针获取的值应该影响原始切片
	result["1"].Age = 26
	if users[0].Age != 26 {
		t.Error("pointer modification should affect original slice")
	}
}

func TestMap(t *testing.T) {
	users := []testUser{
		{ID: "1", Name: "Alice", Age: 25},
		{ID: "2", Name: "Bob", Age: 30},
	}

	result := Map(users, func(u testUser) string {
		return u.Name
	})

	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}

	if result[0] != "Alice" {
		t.Errorf("expected Alice, got %s", result[0])
	}
}

func TestGroupBy(t *testing.T) {
	users := []testUser{
		{ID: "1", Name: "Alice", Age: 25},
		{ID: "2", Name: "Bob", Age: 30},
		{ID: "3", Name: "Charlie", Age: 25},
	}

	grouped := GroupBy(users, func(u testUser) int {
		return u.Age
	})

	if len(grouped) != 2 {
		t.Errorf("expected 2 groups, got %d", len(grouped))
	}

	if len(grouped[25]) != 2 {
		t.Errorf("expected 2 users with age 25, got %d", len(grouped[25]))
	}
}

func TestUnique(t *testing.T) {
	input := []int{1, 2, 2, 3, 3, 3, 4}
	result := Unique(input)

	if len(result) != 4 {
		t.Errorf("expected 4 unique items, got %d", len(result))
	}

	expected := []int{1, 2, 3, 4}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("expected %d at index %d, got %d", expected[i], i, v)
		}
	}
}

func TestUniqueNonAdjacent(t *testing.T) {
	// 非相邻重复也能去除（区别于 slices.Compact）
	input := []int{1, 2, 1, 3, 2}
	result := Unique(input)

	if len(result) != 3 {
		t.Errorf("expected 3 unique items, got %d", len(result))
	}

	expected := []int{1, 2, 3}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("expected %d at index %d, got %d", expected[i], i, v)
		}
	}
}

func TestFilter(t *testing.T) {
	users := []testUser{
		{ID: "1", Name: "Alice", Age: 25},
		{ID: "2", Name: "Bob", Age: 30},
		{ID: "3", Name: "Charlie", Age: 35},
	}

	result := Filter(users, func(u testUser) bool {
		return u.Age >= 30
	})

	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
}

func TestReduce(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}

	sum := Reduce(numbers, 0, func(acc, curr int) int {
		return acc + curr
	})

	if sum != 15 {
		t.Errorf("expected sum 15, got %d", sum)
	}

	product := Reduce(numbers, 1, func(acc, curr int) int {
		return acc * curr
	})

	if product != 120 {
		t.Errorf("expected product 120, got %d", product)
	}
}

func TestEmptySlices(t *testing.T) {
	t.Run("ToMap with empty slice", func(t *testing.T) {
		result := ToMap([]int{}, func(i int) int { return i })
		if len(result) != 0 {
			t.Errorf("expected empty map, got %d items", len(result))
		}
	})

	t.Run("Filter with empty slice", func(t *testing.T) {
		result := Filter([]int{}, func(i int) bool { return true })
		if len(result) != 0 {
			t.Errorf("expected empty slice, got %d items", len(result))
		}
	})

	t.Run("Map with empty slice", func(t *testing.T) {
		result := Map([]int{}, func(i int) int { return i * 2 })
		if len(result) != 0 {
			t.Errorf("expected empty slice, got %d items", len(result))
		}
	})
}
