package resdb

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

// ------------------------------------------
// 定义各个索引的前缀
type Prefix func(tbname string, indexkind string) string

func DefaultPrefix(tbname string, indexkind string) string {
	return tbname + SPLIT + indexkind + SPLIT
}

// ------------------------------------------

// 定义索引类型接口
type IndexKind interface {
	Name() string
	Prefix(prefix Prefix, tbname string) string
	//JoinIndexValue(fieldsBytes *map[string][]byte, fields ...string) any
}

// 没有索引
type NoIndex struct{}

func (n *NoIndex) Name() string {
	return "NoIndex"
}
func (n *NoIndex) Prefix(prefix Prefix, tbname string) string {
	return prefix(tbname, n.Name())
}
func (n *NoIndex) DefaultPrefix(tbname string) string {
	return DefaultPrefix(tbname, n.Name())
}

// 主键索引
type Primary struct{}

func (p *Primary) Name() string {
	return "Primary"
}
func (p *Primary) Prefix(prefix Prefix, tbname string) string {
	return prefix(tbname, p.Name())
}
func (p *Primary) DefaultPrefix(tbname string) string {
	return DefaultPrefix(tbname, p.Name())
}

// 普通索引
type NormalIndex struct{}

func (i *NormalIndex) Name() string {
	return "Normal"
}
func (i *NormalIndex) Prefix(prefix Prefix, tbname string) string {
	return prefix(tbname, i.Name())
}
func (i *NormalIndex) DefaultPrefix(tbname string) string {
	return DefaultPrefix(tbname, i.Name())
}

// 全文索引
type FullTextIndex struct {
	ftlen uint8 // 全文索引分词长度限制3--11之间
}

func FullTextIndexNew() *FullTextIndex {
	return &FullTextIndex{ftlen: uint8(5)}
}
func (i *FullTextIndex) Set(ftlen int) {
	if ftlen < 3 || ftlen > 11 {
		ftlen = 5
	}
	i.ftlen = uint8(ftlen)
}
func (i *FullTextIndex) GetLen() int {
	return int(i.ftlen)
}
func (i *FullTextIndex) Name() string {
	return "FullText"
}
func (i *FullTextIndex) DefaultPrefix(tbname string) string {
	return DefaultPrefix(tbname, i.Name())
}
func (i *FullTextIndex) Prefix(prefix Prefix, tbname string) string {
	return prefix(tbname, i.Name())
}

// ------------------------------------------

// 字段索引类型
type FieldIndexKind struct {
	field     string
	indexKind IndexKind
}

// 设置值
func (fit *FieldIndexKind) Set(field string, kind IndexKind) {
	fit.field = field
	fit.indexKind = kind
}

// 获取索引类型
func (fit *FieldIndexKind) IndexKind() IndexKind {
	return fit.indexKind
}

// 获取索引字段
func (fit *FieldIndexKind) Field() string {
	return fit.field
}

// ------------------------------------------
// 断言索引类型函数
type AssertIndexKind = func(fieldIndexKind *[]FieldIndexKind) IndexKind

// 默认断言索引类型函数
// 判断索引是何种类型,主键，普通索引，全文索引
func DefaultAssertIndexKind(fieldIndexKind *[]FieldIndexKind) IndexKind {
	PrimaryKeyIndexCount := 0
	NormalIndexCount := 0
	FullTextIndexCount := 0
	for _, fit := range *fieldIndexKind {
		switch fit.indexKind.Name() {
		case "Primary":
			PrimaryKeyIndexCount++
		case "Normal":
			NormalIndexCount++
		case "FullText":
			FullTextIndexCount++
		}
	}
	//优先1
	//规则是：有一个主键，则认为是主键索引
	if PrimaryKeyIndexCount > 0 {
		return &Primary{}
	}
	//优先2
	//规则是：有一个全文索引，则认为是全文索引
	if FullTextIndexCount > 0 {
		return &FullTextIndex{}
	}
	//优先3
	//规则是：剩下的是普通索引
	if NormalIndexCount > 0 {
		return &NormalIndex{}
	}
	return nil
}

// ------------------------------------------
// 全文索引遍历分词法
func GetFullTextToken(nr string, ftlen int) (tokens []string) {
	var knr string //, fid
	var ml, cl int
	var r, idxstr []rune
	r = []rune(nr)
	cl = len([]rune(nr))
	for cl > 0 {
		ml = min(cl, ftlen)
		idxstr = r[:ml]
		knr = string(idxstr)
		tokens = append(tokens, knr)
		r = r[1:]
		cl = len(r)
	}
	return
}

// ------------------------------------------
// 组织索引值
// 组合索引值，主键和索引是一样的，但是全文索引不同。
type JoinIndexValue func(fieldsBytes *map[string][]byte, tbname string, fieldIndexKind *[]FieldIndexKind, fields ...string) any

// 将字段对应的字节值拼接起来，用SPLIT分隔
func DefaultJoinIndexValue(fieldsBytes *map[string][]byte, tbname string, fieldIndexKind *[]FieldIndexKind, fields ...string) any {
	var indexValue []byte
	for i, field := range fields {
		if v, ok := (*fieldsBytes)[field]; !ok {
			return nil
		} else {
			indexValue = append(indexValue, v...)
			if i < len(fields)-1 {
				indexValue = append(indexValue, SPLIT...)
			}
		}
	}
	pfx := (*fieldIndexKind)[0].indexKind.Prefix(DefaultPrefix, tbname)
	indexValue = bytes.Join([][]byte{[]byte(pfx), indexValue}, []byte(SPLIT))
	return indexValue
}

// 将字段对应的字节值拼接起来，用SPLIT分隔
func JoinFullTextIndexValue(fieldsBytes *map[string][]byte, tbname string, fieldIndexKind *[]FieldIndexKind, fields ...string) any {
	var r [][]byte
	var ftpxf []byte
	var fieldValue []byte
	var exists bool
	var curpfx [][]byte
	//curpfx = append(curpfx, fullTextPrefix)
	var isnil bool
	for _, field := range fields {
		// 获取字段的实际值
		fieldValue, exists = (*fieldsBytes)[field]
		if !exists {
			continue // 跳过不存在的字段
		}
		var isFullTextField bool
		ftlen := 5

		for _, fit := range *fieldIndexKind {
			if fit.field == field {
				isFullTextField = fit.indexKind.Name() == "FullText"
				ftlen = fit.indexKind.(*FullTextIndex).GetLen()
				if isFullTextField {
					ftpxf = []byte(fit.indexKind.Prefix(DefaultPrefix, tbname))
					break
				}
			}
		}
		//该字段是否是全文索引字段
		switch isFullTextField {
		case true: // 全文索引字段
			// 将字段值转换为字符串，全文索引只支持字符串
			fieldStr := string(fieldValue)
			// 对字段值进行分词
			tokens := GetFullTextToken(fieldStr, ftlen)
			for _, token := range tokens {
				if token == "" {
					continue
				}
				curpfx = append(curpfx, bytes.Join([][]byte{ftpxf, []byte(token)}, []byte(SPLIT)))
			}
		case false: // 普通索引字段
			curpfx = append(curpfx, bytes.Join([][]byte{ftpxf, fieldValue}, []byte(SPLIT)))
		}
		//返回结果是否为空
		isnil = r == nil
		switch isnil {
		case true:
			//r = append(r, append([][]byte{}, ftpxf...)...)
			r = append(r, append([][]byte{}, curpfx...)...)
		case false: // 拼接当前索引前缀到结果中,一对多和多对一关系，多对多已经过滤，属于非法。
			for i, idx := range r {
				for _, pfx := range curpfx {
					var tidx []byte
					tidx = append(tidx, append([]byte{}, idx...)...)
					var tpfx []byte
					tpfx = append(tpfx, append([]byte{}, pfx...)...)
					tidx = bytes.Join([][]byte{tidx, tpfx}, []byte{})
					r[i] = tidx
				}
			}
		}
		curpfx = curpfx[:0]
		ftpxf = ftpxf[:0]
	}
	return r
}

// ------------------------------------------
// 索引的结构体
type Index struct {
	set             []FieldIndexKind // 主键字段列表，单主键长度为1，组合主键长度>1
	assertIndexKind AssertIndexKind
	indexKind       IndexKind //可以手动设置索引类型，如果没有手动设置，则根据断言索引类型函数自动设置
}

func IndexNew(fieldIndexKind []FieldIndexKind, assertIndexKind AssertIndexKind, indexKind IndexKind) *Index {
	return &Index{
		set:             fieldIndexKind,
		assertIndexKind: assertIndexKind,
		indexKind:       indexKind,
	}
}
func IndexNewDefault(fieldIndexKind []FieldIndexKind) *Index {
	return IndexNew(fieldIndexKind, DefaultAssertIndexKind, nil)
}

// 判断索引是何种类型,主键，普通索引，全文索引
func (idx *Index) AssertIndexKind(fieldIndexKind *[]FieldIndexKind) *IndexKind {
	if idx.indexKind != nil {
		return &idx.indexKind
	}
	idx.indexKind = idx.assertIndexKind(fieldIndexKind)
	return &idx.indexKind
}
func (idx *Index) GetName() string {
	fields := make([]string, 0, len(idx.set))
	for _, fik := range idx.set {
		fields = append(fields, fik.Field())
	}
	return strings.Join(fields, "-")
}

// 普通索引值拼接
func (idx *Index) JoinIndexValue(joinIndexValue JoinIndexValue, fieldsBytes *map[string][]byte, tbname string, fields ...string) any {
	return joinIndexValue(fieldsBytes, tbname, &idx.set, fields...)
}
func (idx *Index) DefaultJoinIndexValue(fieldsBytes *map[string][]byte, tbname string, fields ...string) any {
	return DefaultJoinIndexValue(fieldsBytes, tbname, &idx.set, fields...)
}
func (idx *Index) FullTextJoinIndexValue(fieldsBytes *map[string][]byte, tbname string, fields ...string) any {
	return JoinFullTextIndexValue(fieldsBytes, tbname, &idx.set, fields...)
}

// ------------------------------------------

// 索引集的结构体
type IndexSet struct {
	set []Index // 主键字段列表，单主键长度为1，组合主键长度>1
}

// 主键必须先添加，才能优先匹配主键
func (is *IndexSet) Add(index ...Index) {
	is.set = append(is.set, index...)
}

// 匹配索引
// 当用户提交查询时，根据查询的字段，匹配索引集中的对应索引
func (is *IndexSet) Match(fields []string) []FieldIndexKind {
	Count := 0
	//主键必须先添加，才能优先匹配主键，优先级由用户手动添加的顺序决定
	for _, idx := range is.set {
		for _, field := range idx.set {
			if slices.Contains(fields, field.field) {
				Count++
			}
		}
		if Count == len(idx.set) {
			return idx.set
		}
	}
	return nil
}

// 数据库的添加，删除，修改等操作
// 匹配字段ufield ，用于记录修改某个字段时，只需要处理这些字段的索引即可。
// ins 表示是否是插入操作，true 表示put操作，false 表示delete操作。
func (is *IndexSet) Operation(tbname string, fieldsBytes *map[string][]byte, ins bool, batch *leveldb.Batch, ufield ...string) {

	// 遍历索引集
	for _, idx := range is.set {
		// 遍历索引中的每个字段
		for _, fik := range idx.set {
			// 如果字段在ufield中，或者ufield为空（表示所有字段都需要更新）
			fmt.Printf("fik: %v\n", fik)
		}
	}
}
