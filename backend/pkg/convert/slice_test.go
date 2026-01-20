package convert

import (
	"testing"
)

type testUser struct {
	ID   string
	Name string
	Age  int
}

func TestSliceToMap(t *testing.T) {
	users := []testUser{
		{ID: "1", Name: "Alice", Age: 25},
		{ID: "2", Name: "Bob", Age: 30},
		{ID: "3", Name: "Charlie", Age: 35},
	}

	result := SliceToMap(users, func(u testUser) string {
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

func TestSliceToMapPtr(t *testing.T) {
	users := []testUser{
		{ID: "1", Name: "Alice", Age: 25},
		{ID: "2", Name: "Bob", Age: 30},
	}

	result := SliceToMapPtr(users, func(u testUser) string {
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

func TestMapToSlice(t *testing.T) {
	m := map[string]testUser{
		"1": {ID: "1", Name: "Alice", Age: 25},
		"2": {ID: "2", Name: "Bob", Age: 30},
	}

	result := MapToSlice(m)

	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
}

func TestMapKeys(t *testing.T) {
	m := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	keys := MapKeys(m)

	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}

func TestMapValues(t *testing.T) {
	m := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	values := MapValues(m)

	if len(values) != 3 {
		t.Errorf("expected 3 values, got %d", len(values))
	}
}

func TestSliceToUnique(t *testing.T) {
	input := []int{1, 2, 2, 3, 3, 3, 4}
	result := SliceToUnique(input)

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

func TestSliceFilter(t *testing.T) {
	users := []testUser{
		{ID: "1", Name: "Alice", Age: 25},
		{ID: "2", Name: "Bob", Age: 30},
		{ID: "3", Name: "Charlie", Age: 35},
	}

	result := SliceFilter(users, func(u testUser) bool {
		return u.Age >= 30
	})

	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
}

func TestSliceMap(t *testing.T) {
	users := []testUser{
		{ID: "1", Name: "Alice", Age: 25},
		{ID: "2", Name: "Bob", Age: 30},
	}

	result := SliceMap(users, func(u testUser) string {
		return u.Name
	})

	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}

	if result[0] != "Alice" {
		t.Errorf("expected Alice, got %s", result[0])
	}
}

func TestSliceReduce(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}

	sum := SliceReduce(numbers, 0, func(acc, curr int) int {
		return acc + curr
	})

	if sum != 15 {
		t.Errorf("expected sum 15, got %d", sum)
	}

	product := SliceReduce(numbers, 1, func(acc, curr int) int {
		return acc * curr
	})

	if product != 120 {
		t.Errorf("expected product 120, got %d", product)
	}
}

func TestSliceContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	if !SliceContains(slice, "banana") {
		t.Error("expected to contain banana")
	}

	if SliceContains(slice, "orange") {
		t.Error("expected not to contain orange")
	}
}

func TestSliceGroupBy(t *testing.T) {
	users := []testUser{
		{ID: "1", Name: "Alice", Age: 25},
		{ID: "2", Name: "Bob", Age: 30},
		{ID: "3", Name: "Charlie", Age: 25},
	}

	grouped := SliceGroupBy(users, func(u testUser) int {
		return u.Age
	})

	if len(grouped) != 2 {
		t.Errorf("expected 2 groups, got %d", len(grouped))
	}

	if len(grouped[25]) != 2 {
		t.Errorf("expected 2 users with age 25, got %d", len(grouped[25]))
	}
}

func TestSliceChunk(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	chunks := SliceChunk(slice, 3)

	if len(chunks) != 3 {
		t.Errorf("expected 3 chunks, got %d", len(chunks))
	}

	if len(chunks[0]) != 3 {
		t.Errorf("expected first chunk size 3, got %d", len(chunks[0]))
	}

	if len(chunks[2]) != 3 {
		t.Errorf("expected last chunk size 3, got %d", len(chunks[2]))
	}
}

func TestSliceChunkUneven(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	chunks := SliceChunk(slice, 3)

	if len(chunks) != 4 {
		t.Errorf("expected 4 chunks, got %d", len(chunks))
	}

	if len(chunks[3]) != 1 {
		t.Errorf("expected last chunk size 1, got %d", len(chunks[3]))
	}
}

func TestSliceReverse(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}

	reversed := SliceReverse(slice)

	if len(reversed) != 5 {
		t.Errorf("expected length 5, got %d", len(reversed))
	}

	if reversed[0] != 5 || reversed[4] != 1 {
		t.Error("slice was not reversed correctly")
	}
}

func TestSliceMax(t *testing.T) {
	numbers := []int{3, 1, 4, 1, 5, 9, 2, 6}

	max := SliceMax(numbers)

	if max != 9 {
		t.Errorf("expected max 9, got %d", max)
	}
}

func TestSliceMaxEmpty(t *testing.T) {
	numbers := []int{}

	max := SliceMax(numbers)

	if max != 0 {
		t.Errorf("expected zero value for empty slice, got %d", max)
	}
}

func TestSliceMin(t *testing.T) {
	numbers := []int{3, 1, 4, 1, 5, 9, 2, 6}

	min := SliceMin(numbers)

	if min != 1 {
		t.Errorf("expected min 1, got %d", min)
	}
}

func TestSliceMinEmpty(t *testing.T) {
	numbers := []int{}

	min := SliceMin(numbers)

	if min != 0 {
		t.Errorf("expected zero value for empty slice, got %d", min)
	}
}

func TestEmptySlices(t *testing.T) {
	t.Run("SliceToMap with empty slice", func(t *testing.T) {
		result := SliceToMap([]int{}, func(i int) int { return i })
		if len(result) != 0 {
			t.Errorf("expected empty map, got %d items", len(result))
		}
	})

	t.Run("SliceFilter with empty slice", func(t *testing.T) {
		result := SliceFilter([]int{}, func(i int) bool { return true })
		if len(result) != 0 {
			t.Errorf("expected empty slice, got %d items", len(result))
		}
	})

	t.Run("SliceMap with empty slice", func(t *testing.T) {
		result := SliceMap([]int{}, func(i int) int { return i * 2 })
		if len(result) != 0 {
			t.Errorf("expected empty slice, got %d items", len(result))
		}
	})
}
