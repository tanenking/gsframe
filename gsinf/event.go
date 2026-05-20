package gsinf

// ///////////////////////////////////////////////////////////////////////////////////////////////////
type IEventManager interface {
	AddEventListener(cb func(args ...interface{})) int
	RemoveEventLister(cb_id int)
	DispatchEvent(params ...interface{})
	Update()
}
