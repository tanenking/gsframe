package gsframe

import "math/rand"

func GetRandomBetweenI64(min int64, max int64) int64 {
	if min == max {
		return min
	}
	_min := min
	_max := max
	if _min > _max {
		_min = max
		_max = min
	}
	v := _max - _min
	if v <= 0 {
		return _min
	}
	v++

	return min + rand.Int63n(v)
}
func GetRandomBetween[T int | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64](min T, max T) T {
	v := GetRandomBetweenI64(int64(min), int64(max))
	return T(v)
}

func GetProbResult[T int | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64](prob T) bool {
	//万分比
	g := T(rand.Intn(10000) + 1)

	return g <= prob
}

// 二维数组 [[id,weight]]
func GetWeightFromJsonArray[T int | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64](weights [][]T) T {
	if len(weights) <= 0 {
		return 0
	}
	type wd_t struct {
		idx    T
		weight T
	}

	wdsum := T(0)
	wds := []wd_t{}
	for _, v := range weights {
		if len(v) < 2 {
			continue
		}
		wdsum += v[1]
		wds = append(wds, wd_t{
			idx:    v[0],
			weight: wdsum,
		})
	}
	r := GetRandomBetween(0, int32(wdsum)-1)

	for i := 0; i < len(wds); i++ {
		if r < int32(wds[i].weight) {
			return wds[i].idx
		}
	}
	return 0
}
func GetWeightFromMaps[T int | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64](weights map[T]T) T {
	if len(weights) <= 0 {
		return 0
	}
	type wd_t struct {
		idx    T
		weight T
	}

	wdsum := T(0)
	wds := []wd_t{}
	for k, v := range weights {
		if v > 0 {
			wdsum += v
			wds = append(wds, wd_t{
				idx:    k,
				weight: wdsum,
			})
		}
	}
	r := GetRandomBetween(0, int32(wdsum)-1)

	for i := 0; i < len(wds); i++ {
		if r < int32(wds[i].weight) {
			return wds[i].idx
		}
	}
	return 0
}
func GetWeightFromProbs[T int | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64](probs []T) int {
	if len(probs) <= 0 {
		return 0
	}
	type wd_t struct {
		idx    int
		weight T
	}

	wdsum := T(0)
	wds := []wd_t{}
	for k, v := range probs {
		if v > 0 {
			wdsum += v
			wds = append(wds, wd_t{
				idx:    k,
				weight: wdsum,
			})
		}
	}
	r := GetRandomBetween(0, wdsum-1)

	for i := 0; i < len(wds); i++ {
		if r < wds[i].weight {
			return wds[i].idx
		}
	}
	return 0
}
