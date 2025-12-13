// 设计原则是当前最简单快捷开发，不考虑通用性和将来扩展要求。
package storedb

import (
	"bytes"
	"encoding/json/v2" //设置一个系统变量 GOEXPERIMENT=jsonv2 开启v2
	"fmt"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

const SPLIT = "-" //分隔符
/*
表结构，半结构泛型化支持
提供灵活的组合索引优化查询
*/
type Table struct {
	name string
	/*
		半结构，可以随意增加字段
		泛型化，字段可以存储任意类型，不限制某种类型。至于有什么作用，看自己发挥。
		底层支持泛型，业务上则由自己定义规则。
		默认固定一个id字段为自动增值，当值为nil时，使用counter自动增值。
	*/
	fields map[string]any //string为字段名，any为字段值。
	//下面的数组元素，对应于上面的fields的key字段名称
	primary  string       //只支持单个主键。
	counter  atomic.Int64 //自动增值计数器
	index    [][]string
	fullText []string
	ftlen    uint8 //全文索引分词长度，默认是5.限制是3--11.
	rsdb     *rsdb
}

// 新建一个表
// 表名不能包含分隔符SPLIT("-")，否则返回nil
// 每次项目启动都会重新创建，或通过json转换为Table结构体
func TableNew(name string) *Table {
	// 检查表名是否包含分隔符
	if strings.Contains(name, SPLIT) {
		fmt.Printf("表名 '%s' 不能包含分隔符 '%s'\n", name, SPLIT)
		return nil
	}
	if RsDB == nil {
		OpenDb("db")
	}
	t := &Table{
		name:    name,
		fields:  make(map[string]any),
		primary: "id",
		ftlen:   5,
		rsdb:    RsDB,
	}
	t.InitAuto()
	return t
}

// 初始化自动增值的值
func (t *Table) InitAuto() {
	// 检查主键是否已经显式设置
	if t.fields[t.primary] != nil {
		// 如果主键已经显式设置，则不更新计数器
		return
	}
	maxValue := t.MaxAutoValue()
	// 根据不同类型进行转换
	switch v := maxValue.(type) {
	case int:
		t.counter.Store(int64(v))
	case int64:
		t.counter.Store(v)
	case float64:
		// JSON解析数字默认是float64
		t.counter.Store(int64(v))
	default:
		// 默认值
		t.counter.Store(1)
	}
}

// 获取最大自动增值记录的主键值
func (t *Table) MaxAutoValue() any {
	key := []byte(t.GetPrimaryPrefix() + SPLIT)
	td := t.rsdb.GetIteratorData(key)
	defer td.Release()
	_, value := td.Last()
	if value != nil {
		var v map[string]any
		if err := json.Unmarshal(value, &v); err == nil {
			return v[t.primary]
		} else {
			fmt.Printf("初始化自动增值失败: %v\n", err)
			return 1
		}
	} else {
		return 1
	}
}

// 设置字段值,以第一次设置的数据类型为准
func (t *Table) SetField(field string, value any) error {
	// 字段名不能包含分隔符和标点符号
	if strings.ContainsAny(field, SPLIT+"!\"#$%&'()*+,./:;<=>?@[\\]^`{|}~") {
		return fmt.Errorf("字段名 '%s' 不能包含分隔符或标点符号", field)
	}
	if _, exists := t.fields[field]; !exists {
		t.fields[field] = value
	} else {
		//主键值赋予nil值，是为了后续自动增值。故此为例外的不需要判断类型相同。
		if field == t.primary && value == nil {
			t.fields[field] = nil
			return nil
		}
		// 已存在，判断类型是否相同
		if fmt.Sprintf("%T", t.fields[field]) == fmt.Sprintf("%T", value) {
			t.fields[field] = value
		} else {
			return fmt.Errorf("字段 '%s' 类型冲突，期望: %T, 实际: %T", field, t.fields[field], value)
		}
	}
	return nil
}
func (t *Table) SetFields(fields map[string]any) error {
	for field, value := range fields {
		if err := t.SetField(field, value); err != nil {
			return err
		}
	}
	return nil
}

// nil即表示使用自动增值，使用自动增值每次都需要将主键值设置为nil
func (t *Table) InitPrimary() {
	if t.fields[t.primary] == nil {
		t.fields[t.primary] = t.counter.Add(1)
	}
}

// 将表所有字段转换为JSON字节数组
func (t *Table) GetFields() []byte {
	b, err := json.Marshal(t.fields)
	if err != nil {
		fmt.Printf("字段序列化失败: %v\n", err)
		return nil
	}
	return b
}

// 主键前缀
func (t *Table) GetPrimaryPrefix() string {
	return t.name + SPLIT + "pk" //+ SPLIT
}

// 索引前缀
func (t *Table) GetIndexPrefix() string {
	return t.name + SPLIT + "idx" //+ SPLIT
}

// 全文索引前缀
func (t *Table) GetFullTextPrefix() string {
	return t.name + SPLIT + "ft" //+ SPLIT
}

// 获取主键值
func (t *Table) GetPrimaryValue() []byte {
	t.InitPrimary()
	val, err := json.Marshal(t.fields[t.primary])
	if err != nil {
		fmt.Printf("主键序列化失败: %v\n", err)
		return nil
	}
	return val
}

// 获取主键前缀
func (t *Table) GetPrimaryPrefixValue() []byte {
	return bytes.Join([][]byte{[]byte(t.GetPrimaryPrefix()), t.GetPrimaryValue()}, []byte(SPLIT))
}

// 获取索引值
func (t *Table) GetIndexValue() [][]byte {
	indexValues := make([][]byte, 0, len(t.index))
	indexPrefix := []byte(t.GetIndexPrefix())
	for _, index := range t.index {
		var idx bytes.Buffer
		idx.Write(indexPrefix)
		for _, field := range index {
			// 获取字段值
			fieldValue, exists := t.fields[field]
			if !exists {
				// 如果字段不存在，跳过该索引
				break
			}
			// 序列化字段值
			val, err := json.Marshal(fieldValue)
			if err != nil {
				fmt.Printf("索引字段序列化失败: %v\n", err)
				return nil
			}

			// 在每个索引字段前添加分隔符（包括第一个）
			idx.Write([]byte(SPLIT))
			idx.Write(val)
		}
		indexValues = append(indexValues, idx.Bytes())
	}
	return indexValues
}

// 获取全文索引值
func (t *Table) GetFullTextValue() [][]byte {
	fullTextValues := make([][]byte, 0, len(t.fullText))
	if t.ftlen == 0 {
		t.ftlen = 5
	}
	fullTextPrefix := []byte(t.GetFullTextPrefix())
	for _, field := range t.fullText {
		// 获取字段的实际值
		fieldValue, exists := t.fields[field]
		if !exists {
			continue // 跳过不存在的字段
		}
		// 将字段值转换为字符串
		var fieldStr string
		switch v := fieldValue.(type) {
		case string:
			fieldStr = v
		default:
			// 将其他类型转换为字符串
			val, err := json.Marshal(v)
			if err != nil {
				fmt.Printf("字段值转换为字符串失败: %v\n", err)
				continue
			}
			fieldStr = string(val)
		}
		// 对字段值进行分词
		tokens := t.GetFullTextToken(fieldStr, int(t.ftlen))
		for _, token := range tokens {
			var idx bytes.Buffer
			idx.Write(fullTextPrefix)
			idx.Write([]byte(SPLIT))

			val, err := json.Marshal(token)
			if err != nil {
				fmt.Printf("全文索引词序列化失败: %v\n", err)
				return nil
			}
			idx.Write(val)

			fullTextValues = append(fullTextValues, idx.Bytes())
			//fmt.Printf("fullTextValues: %s\n", idx.String())
		}
	}

	return fullTextValues
}

// 全文索引遍历分词法
func (t *Table) GetFullTextToken(nr string, ftlen int) (tokens []string) {
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

// 设置主键字段
func (t *Table) SetPrimaryValue(primary string) error {
	t.primary = primary
	return nil
}

// 设置索引字段
func (t *Table) SetIndexValue(index [][]string) error {
	t.index = index
	return nil
}

// 添加单个索引
func (t *Table) AddIndex(index []string) error {
	t.index = append(t.index, index)
	return nil
}

// 设置全文索引字段
func (t *Table) SetFullTextValue(fullText []string) error {
	t.fullText = fullText
	return nil
}

// 添加单个全文索引字段
func (t *Table) AddFullTextField(field string) error {
	t.fullText = append(t.fullText, field)
	return nil
}

// 获取单个字段值
func (t *Table) GetField(field string) (any, bool) {
	if t.fields == nil {
		return nil, false
	}
	value, exists := t.fields[field]
	return value, exists
}

// 获取所有字段
func (t *Table) GetAllFields() map[string]any {
	// 返回字段的副本，避免直接修改内部状态
	result := make(map[string]any)
	for k, v := range t.fields {
		result[k] = v
	}
	return result
}

// 设置全文索引分词长度
func (t *Table) SetFullTextLen(ftlen uint8) error {
	if ftlen >= 3 && ftlen <= 11 {
		t.ftlen = ftlen
		return nil
	}
	return fmt.Errorf("全文索引分词长度 '%d' 无效，必须在 3-11 之间", ftlen)
}

// 添加主键记录
func (t *Table) RecordKV(batch *leveldb.Batch, put bool) error {
	pv := t.GetPrimaryPrefixValue()
	pvFields := t.GetFields()
	if put {
		batch.Put(pv, pvFields)
	} else {
		batch.Delete(pv)
	}
	return nil
}

// put索引KV
func (t *Table) IndexKV(batch *leveldb.Batch, put bool) {
	idxs := t.GetIndexValue()
	value := t.GetPrimaryValue()
	for _, idx := range idxs {
		if put {
			batch.Put(idx, value)
		} else {
			batch.Delete(idx)
		}
	}
}

// put全文索引KV
func (t *Table) FullTextKV(batch *leveldb.Batch, put bool) {
	idxs := t.GetFullTextValue()
	value := t.GetPrimaryValue()
	for _, idx := range idxs {
		if put {
			batch.Put(idx, value)
		} else {
			batch.Delete(idx)
		}
	}
}

// 插入记录
func (t *Table) Insert() error {
	batch := Batch.Get().(*leveldb.Batch)
	defer Batch.Put(batch)
	t.RecordKV(batch, true)
	t.IndexKV(batch, true)
	t.FullTextKV(batch, true)
	err := t.rsdb.Db.Write(batch, nil)
	if err != nil {
		return err
	}
	return nil
}

// 删除记录
func (t *Table) Delete() error {
	batch := Batch.Get().(*leveldb.Batch)
	defer Batch.Put(batch)
	t.RecordKV(batch, false)
	t.IndexKV(batch, false)
	t.FullTextKV(batch, false)
	err := t.rsdb.Db.Write(batch, nil)
	if err != nil {
		return err
	}
	return nil
}

// 更新记录，由于项目基本没有更新操作，所以并不考虑性能和一致性。
func (t *Table) Update(fields map[string]any) error {
	// 检查是否提供了主键字段
	primaryValue, ok := fields[t.primary]
	if !ok {
		return fmt.Errorf("更新操作必须提供主键字段 '%s'", t.primary)
	}

	// 读取旧记录
	oldFields := t.Read(primaryValue)
	if oldFields == nil {
		return fmt.Errorf("主键值 '%v' 的记录不存在", primaryValue)
	}

	// 创建新记录
	newTable := TableNew(t.name)
	newTable.primary = t.primary

	// 合并字段值
	for k, v := range oldFields {
		newTable.fields[k] = v
	}
	for k, v := range fields {
		newTable.fields[k] = v
	}

	// 删除旧记录并插入新记录
	if err := t.Delete(); err != nil {
		return err
	}

	return newTable.Insert()
}

// 从按主键数据库读取记录
func (t *Table) Read(primary any) map[string]any {
	pb, err := json.Marshal(primary)
	if err != nil {
		fmt.Printf("主键序列化失败: %v\n", err)
		return nil
	}

	// 优化字符串拼接
	var key bytes.Buffer
	key.WriteString(t.GetPrimaryPrefix())
	key.WriteString(SPLIT)
	key.Write(pb)

	v, err := t.rsdb.Db.Get(key.Bytes(), nil)
	if err != nil {
		// leveldb.ErrNotFound 是正常的未找到错误，不需要打印
		if err != leveldb.ErrNotFound {
			fmt.Printf("读取记录失败: %v\n", err)
		}
		return nil
	}

	fields := make(map[string]any)
	if err := json.Unmarshal(v, &fields); err != nil {
		fmt.Printf("记录反序列化失败: %v\n", err)
		return nil
	}

	return fields
}

// 根据字段名获取对应的主键，索引，全文索引前缀
func (t *Table) GetPrefix(field string) string {
	if field == t.primary {
		return t.GetPrimaryPrefix()
	}
	for _, idx := range t.index {
		if idx[0] == field { //必须是第一个索引字段
			return t.GetIndexPrefix()
		}
	}
	if slices.Contains(t.fullText, field) {
		return t.GetFullTextPrefix()
	}
	return ""
}

// 遍历表所有kv，复制表用
func (t *Table) For() iterator.Iterator {
	pfx := t.name + SPLIT
	return t.rsdb.GetIterator([]byte(pfx))
}

// 遍历表所有数据
func (t *Table) ForData() *TableData {
	pfx := t.GetPrimaryPrefix()
	return t.rsdb.GetIteratorData([]byte(pfx))
}

// 根据字段名和值搜索返回迭代器
func (t *Table) Search(field string, value ...any) iterator.Iterator {
	pfx := t.GetPrefix(field)
	keys := [][]byte{}
	if len(value) == 0 {
		return nil
	}
	// 使用第一个值进行搜索
	for _, v := range value {
		jsonData, err := json.Marshal(v)
		if err != nil {
			fmt.Printf("值序列化失败: %v\n", err)
			return nil
		}
		//示例值["Bob"]=>"Bob
		jsonData = jsonData[1 : len(jsonData)-1] // 去掉首尾的双中括号[]="Bob"
		if jsonData[len(jsonData)-1] == '"' {    //如果是字符串，去掉结尾的双引号
			jsonData = jsonData[:len(jsonData)-1] // 去掉结尾的双引号="Bob
		}
		key := bytes.Join([][]byte{[]byte(pfx), []byte(jsonData)}, []byte(SPLIT))
		keys = append(keys, key)
	}
	return t.rsdb.GetIterator(keys...)
}

// 根据字段名和值搜索返回数据迭代器
func (t *Table) SearchData(field string, value ...any) *TableData {
	return TableDataNew(t.Search(field, value))
}
