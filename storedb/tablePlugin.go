package storedb

//添加插件机制，便于扩展功能
type TablePlugin interface {
	Name() string
	Initialize(table *Table) error
	OnCreate() error
	OnUpdate() error
	OnDelete() error
	OnQuery() error
	// ...
}

/*
type Table struct {
    // ... 现有字段
    plugins []TablePlugin
}
*/
