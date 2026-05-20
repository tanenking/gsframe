package gsframe

type Random struct {
	rand int64
}

func CreateRandom(seed int64) *Random {
	r := &Random{rand: seed}
	return r
}

/**取得64位随机数 */
func (r *Random) Rand64() int64 {
	// Xorshift 算法（更快、随机性更好）
	r.rand ^= r.rand << 13
	r.rand ^= r.rand >> 17
	r.rand ^= r.rand << 5
	// 限制范围在 [0, 2^31)
	r.rand = r.rand & ((1 << 31) - 1)
	return r.rand
}

/**取得32位随机数 */
func (r *Random) Rand32() int32 {
	return int32(r.Rand64() & 0x7fffffff)
}

/**取得32位随机数 [0-n)*/
func (r *Random) Int32n(n int32) int32 {
	if n <= 0 {
		return 0
	}
	return r.Rand32() % n
}
