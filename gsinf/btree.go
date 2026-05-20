package gsinf

type IBSTData interface {
	Key() string
	Compare(IBSTData) bool //与目标比较大小, false 小, true 大
}

type IBSTree interface {
	Set(value IBSTData) int32
	Delete(key string)
	GetRankData(key string) (int32, IBSTData)          //返回排行名次和数据, 不存在返回 0
	GetDataFromRank(rank int32) IBSTData               //通过名次返回数据
	GetRankList(count uint32) []IBSTData               //获取所有数据, 已排序, 0代表取所有数据
	EnumRankList(func(nRank int32, val IBSTData) bool) //枚举排行榜 return false 中断枚举
	Size() uint32
	ClearBSTree()
}
