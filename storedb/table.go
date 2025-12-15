// 设计原则是当前最简单快捷开发，不考虑通用性和将来扩展要求。
// 除了需要排序的主键和索引需要转换为[]byte外，其他所有字段值，皆转换为字符串存储
package storedb

import (
	"bytes" //设置一个系统变量 GOEXPERIMENT=jsonv2 开启v2
	"fmt"
	"maps"
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
	index    [][]string   //单索引，组合索引，全文索引等的各种组合索引
	fullText []string     //全文索引字段名
	ftlen    uint8        //全文索引分词长度，默认是5.限制是3--11.
	rsdb     *rsdb
}

func (t *Table) Contains(v []byte) bool {
	panic("unimplemented")
}

// 新建一个表
// 表名不能包含分隔符SPLIT("-")，否则返回nil
// 每次项目启动都会重新创建，或通过json转换为Table结构体
func TableNew(name string) (*Table, error) {
	// 检查表名是否包含分隔符
	if strings.Contains(name, SPLIT) {
		fmt.Printf("表名 '%s' 不能包含分隔符 '%s'\n", name, SPLIT)
		return nil, fmt.Errorf("表名 '%s' 不能包含分隔符 '%s'", name, SPLIT)
	}
	if RsDB == nil {
		OpenDb("db")
	}
	return &Table{
		name:    name,
		fields:  make(map[string]any),
		primary: "id",
		ftlen:   5,
		rsdb:    RsDB,
	}, nil
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
	iter := t.rsdb.GetIterator(key)
	defer iter.Release()
	var rkey []byte
	if iter.Last() {
		rkey = iter.Key()
	} else {
		return 1
	}
	rkey = rkey[len(key):]
	var target any
	if t.fields[t.primary] == nil {
		target = int(1)
	} else {
		target = t.fields[t.primary]
	}
	return Bytes(rkey).ToAny(target)
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
		if t.counter.Load() == int64(0) {
			t.InitAuto()
		}
		t.fields[t.primary] = t.counter.Add(1)
	}
}

// 将表所有字段转换为字节数组
// 格式为: field1:value1-field2:value2,...
func (t *Table) GetFieldsValue() []byte {
	sr := ""
	for field, value := range t.fields {
		if value == nil {
			continue
		}
		sr += field + ":" + AnyToStr(value) + SPLIT
	}
	sr = sr[:len(sr)-len(SPLIT)]
	return []byte(sr)
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
	return AnyToBytes(t.fields[t.primary])
}

// 获取主键前缀
func (t *Table) GetPrimaryPrefixValue() []byte {
	pfx := []byte(t.GetPrimaryPrefix())
	val := t.GetPrimaryValue()
	return bytes.Join([][]byte{pfx, val}, []byte(SPLIT))
}

// 获取索引值
func (t *Table) GetIndexValue() [][]byte {
	indexValues := make([][]byte, 0, len(t.index))
	indexPrefix := []byte(t.GetIndexPrefix())
	var val []byte
	var exists bool
	var fieldValue any
	for _, index := range t.index {
		var idx bytes.Buffer
		idx.Write(indexPrefix)
		for _, field := range index {
			// 获取字段值
			fieldValue, exists = t.fields[field]
			// 如果字段不存在或值为nil，跳过该索引
			if !exists || fieldValue == nil {
				break
			}
			val = AnyToBytes(fieldValue)
			// 在每个索引字段前添加分隔符（包括第一个）
			idx.Write([]byte(SPLIT))
			idx.Write(val)
		}
		idxbyte := idx.Bytes()
		// 索引值不能超过255字节
		if len(idxbyte) >= 256 {
			idxbyte = idxbyte[:255]
		}
		indexValues = append(indexValues, idxbyte)
	}
	return indexValues
}

// 获取全文索引值
// 只支持字符串类型的全文索引，其他类型的字段值会被转换为字符串。
func (t *Table) GetFullTextValue() [][]byte {
	fullTextValues := make([][]byte, 0, len(t.fullText))
	if t.ftlen == 0 {
		t.ftlen = 5
	}
	fullTextPrefix := []byte(t.GetFullTextPrefix())
	var fieldValue any
	var exists bool
	for _, field := range t.fullText {
		// 获取字段的实际值
		fieldValue, exists = t.fields[field]
		if !exists {
			continue // 跳过不存在的字段
		}
		// 将字段值转换为字符串，全文索引只支持字符串
		fieldStr := AnyToStr(fieldValue)
		// 对字段值进行分词
		tokens := t.GetFullTextToken(fieldStr, int(t.ftlen))

		for _, token := range tokens {
			var idx bytes.Buffer
			idx.Write(fullTextPrefix)
			idx.Write([]byte(SPLIT))
			idx.Write([]byte(token))
			fullTextValues = append(fullTextValues, idx.Bytes())
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
	maps.Copy(result, t.fields)
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
	pvFields := t.GetFieldsValue()
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
	record := t.Read(primaryValue)
	oldFields := t.ParseValue(record)
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
func (t *Table) Read(primary any) []byte {
	pb := AnyToBytes(primary)
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

	return v
}

// 将记录转换为map
func (t *Table) ParseValue(record []byte) map[string]any {
	if record == nil {
		return nil
	}
	bs := Bytes(record).Split()
	fields := make(map[string]any, len(bs))
	var kv [][]byte
	file := ""
	for _, b := range bs {
		kv = Bytes(b).Split(':')
		if len(kv) == 2 {
			file = string(kv[0])
			fields[file] = Bytes(kv[1]).ToAny(t.fields[file])
		}
	}
	return fields
}

// 根据字段名获取对应的主键，索引，全文索引前缀
func (t *Table) GetPrefix(field string) (string, int) {
	if field == t.primary {
		return t.GetPrimaryPrefix(), 0
	}
	for _, idx := range t.index {
		if idx[0] == field { //必须是第一个索引字段
			return t.GetIndexPrefix(), 1
		}
	}
	if slices.Contains(t.fullText, field) {
		return t.GetFullTextPrefix(), 2
	}
	return "", 0
}

// 遍历表所有kv，复制表用
func (t *Table) For() iterator.Iterator {
	pfx := t.name + SPLIT
	return t.rsdb.GetIterator([]byte(pfx))
}

// 遍历表所有数据
func (t *Table) ForData() *TableData {
	pfx := t.GetPrimaryPrefix()
	return TableDataNew(t.rsdb.GetIterator([]byte(pfx)), t)
}

// 根据索引进行搜索返回迭代器
func (t *Table) Search(field ...string) iterator.Iterator {
	if len(field) == 0 {
		return nil
	}
	idx := t.MatchIndex(field...)
	if idx == nil {
		return nil
	}
	pfx, idxType := t.GetPrefix(idx[0])
	keys := []byte(pfx)
	var val any
	var bval []byte
	var fval string
	// 使用第一个值进行搜索
	for _, v := range field {
		val = t.fields[v]
		if idxType != 2 {
			bval = AnyToBytes(val)
			keys = append(keys, bval...)
		} else { //全文索引
			fval = AnyToStr(val)
			if len([]rune(fval)) > int(t.ftlen) {
				fval = string([]rune(fval)[:t.ftlen])
			}
			keys = append(keys, []byte(fval)...)
		}
	}
	return t.rsdb.GetIterator(keys)
}

// 匹配对应的索引字段和索引类型
func (t *Table) MatchIndex(field ...string) ([]string, int) {
	if len(field) == 1 {
		if field[0] == t.primary {
			return []string{t.GetPrimaryPrefix()}, 0
		}
		if slices.Contains(t.fullText, field[0]) {
			return []string{t.GetFullTextPrefix()}, 2
		}
	}

	maxCnt := 0
	var result []string
	for _, idx := range t.index {
		cnt := 0
		for _, f := range field {
			if slices.Contains(idx, f) {
				cnt++
			}
		}
		if cnt > maxCnt {
			maxCnt = cnt
			result = idx
		}
	}
	return result, 1
}

// 根据字段名和值搜索返回数据迭代器
// 缓存迭代器，避免每次for都重新创建迭代器
func (t *Table) SearchData(field ...string) *TableData {
	str := strings.Join(field, ":")
	str = t.name + "." + str
	td := TDCache.Load(str)
	if td == nil {
		iter := t.Search(field...)
		td = TableDataNew(iter, t)
		TDCache.Store(str, td)
	}
	return td
}
