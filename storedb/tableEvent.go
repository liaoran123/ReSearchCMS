package storedb

//### 事件机制
//添加事件机制，便于在表结构变化时触发相应的处理逻辑
type TableEvent int

const (
	EventCreate TableEvent = iota
	EventUpdate
	EventDelete
	EventQuery
	// ...
)

type EventHandler func(event TableEvent, data any) error

/*
type Table struct {
	// ... 现有字段
	eventHandlers map[TableEvent][]EventHandler
}
*/
