package resdb

// IndexKind 定义索引类型枚举
type IndexKind int

const (
	PrimaryKeyIndex IndexKind = iota // 主键
	NormalIndex                      // 普通索引
	FullTextIndex                    // 全文索引
)

// 主键结构体
type Indexs struct {
	kind   IndexKind // 索引类型
	fields []string  // 主键字段列表，单主键长度为1，组合主键长度>1
}

// 实现 IndexType 接口
func (pk *Indexs) Name() string {
	r := ""
	for _, field := range pk.fields {
		r += field + ","
	}
	return r
}
func (pk *Indexs) Kind() IndexKind {
	return pk.kind
}
func (pk *Indexs) Fields() []string {
	return pk.fields
}

func (pk *Indexs) IsComposite() bool {
	return len(pk.fields) > 1
}
