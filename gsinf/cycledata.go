package gsinf

type ICycleData interface {
	GetData() interface{}
	Trigger() <-chan struct{}
}

type ICycle interface {
	CurrSeq() int32
	MaxCount() int32
	GetCycleData(seq int32) ICycleData
	PublishData(data interface{})
}
