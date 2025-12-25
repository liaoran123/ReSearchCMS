// 设计原则是当前最简单快捷开发，不考虑通用性和将来扩展要求。
// 除了需要排序的主键和索引需要转换为[]byte外，其他所有字段值，皆转换为字符串存储
package resdb

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

const SPLIT = "-" //分隔符
/*
表结构，半结构泛型化支持
提供灵活的组合索引优化查询
*/
type Table struct {
	name string // 表名
	/*
		半结构，可以随意增加字段
		泛型化，字段可以存储任意类型，不限制某种类型。至于有什么作用，看自己发挥。
		底层支持泛型，业务上则由自己定义规则。
		默认固定一个id字段为自动增值，当值为nil时，使用counter自动增值。
	*/
	fields map[string]any // 字段映射，string为字段名，any为字段值
	/*
		组合主键时，所有索引对应的主键值都是经过转义的。
		这是为了提取主键中的某个字段值作为匹配索引。
		所以在用索引的主键值回表记录时，需要对索引值进行反转义。
	*/
	indexs *Indexs
	//primary  []string   // 主键字段。
	counter AutoInt // 自动增值计数器，使用自定义的AutoInt
	//index    [][]string // 索引数组，支持单索引和组合索引，以及单全文索引和组合全文索引
	//fullText []string   // 全文索引字段名，需要添加到索引中
	//ftlen    uint8      // 全文索引分词长度，默认是5.限制是3--11
	rsdb *rsdb // 数据库实例，不需要序列化
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
	//检查表名是否为系统保留名
	if slices.Contains(ReservedKeys, name) {
		fmt.Printf("表名 '%s' 是系统保留名，不能使用\n", name)
		return nil, fmt.Errorf("表名 '%s' 是系统保留名，不能使用", name)
	}
	return &Table{
		name:   name,
		fields: make(map[string]any),
		//primary: []string{"id"},
		//ftlen:   5,
		indexs: new(Indexs),
		rsdb:   RsDB,
	}, nil
}

// 获取自动增值的值
func (t *Table) GetAutoInc() int {
	if t.counter.Get() == 0 {
		t.InitAuto()
	}
	return int(t.counter.Increment())
}

// 初始化自动增值的值
func (t *Table) InitAuto() {
	maxValue := t.MaxAutoValue()
	// MaxAutoValue现在直接返回int64类型
	t.counter.Set(int(maxValue))
}

// 获取当前最大自动增值记录的主键值
func (t *Table) MaxAutoValue() int {
	key := []byte(t.name + SPLIT)
	iter := t.rsdb.GetIterator(key)
	defer iter.Release()
	var rkey []byte
	if iter.Last() {
		rkey = iter.Key()
	} else {
		return 0
	}
	rkey = rkey[len(key):]
	var target any
	// 如果主键字段为空，默认使用"id"
	if len(t.indexs.PrimaryFields()) == 0 || t.indexs.PrimaryFields()[0].Field() == "" {
		target = 0
	} else {
		target = t.fields[t.indexs.PrimaryFields()[0].Field()]
	}
	r := Bytes(rkey).ToAny(target)
	// 使用 reflect 包进行类型转换，更灵活地处理各种数值类型
	return AnyToInt(r)

}

// 必须先为表预设字段和类型
func (t *Table) SetFields(fields map[string]any) error {
	//检查fields，不能包含分隔符
	for field := range fields {
		if strings.Contains(field, SPLIT) {
			return fmt.Errorf("字段名 '%s' 不能包含分隔符 '%s'", field, SPLIT)
		}
	}
	t.fields = fields
	return nil
}

// 获取自动增值的值
func (t *Table) AutoValue() int {
	if t.counter.Get() == 0 {
		t.counter.Set(t.MaxAutoValue())
	}
	return t.counter.Increment()
}

// 主键前缀

// GetName 获取表名
func (t *Table) GetName() string {
	return t.name
}

// GetPrimary 获取主键字段名
func (t *Table) GetPrimary() []string {
	primaryFields := make([]string, len(t.indexs.PrimaryFields()))
	for _, field := range t.indexs.PrimaryFields() {
		primaryFields = append(primaryFields, field.Field())
	}
	return primaryFields
}
func (t *Table) AddIndex(index ...*Index) {
	t.indexs.Add(index...)
}

/*
// 检测索引索引，并且返回类型枚举
func (t *Table) MatchIndex(indexs []string) *Indexs {
	idxlen := len(indexs)
	if idxlen == 0 {
		return nil
	}
	idxs := new(Indexs)
	idxs.fields = indexs
	indexCount := 0
	//判断是否符合主键索引
	for _, field := range t.IndexSet.PrimaryFields() {
		if slices.Contains(indexs, field.Field()) {
			indexCount++
		}
	}
	if indexCount == len(indexs) {
		idxs.kind = PrimaryKeyIndex
		return idxs
	}

	for _, idx := range t.IndexSet.Indexs() {
		indexCount = 0
		for _, field := range idx {
			if slices.Contains(indexs, field.Field()) {
				indexCount++
				if indexCount == len(indexs) {
					break
				}
			}
		}
		//判断是否符合普通索引
		if indexCount == len(indexs) {
			//判断是否符合全文索引
			indexCount = 0
			for _, field := range indexs {
				if slices.Contains(t.fullText, field) {
					indexCount++
					break
				}
			}
			if indexCount > 0 {
				idxs.kind = FullTextIndex
				return idxs
			} else {
				idxs.kind = NormalIndex
				return idxs
			}
		}
	}
	idxs.kind = NoIndex
	return idxs
}



// 检测索引索引，并且返回类型枚举
func (t *Table) MatchIndex(indexs []string) *Indexs {
	idxs := new(Indexs)
	idxs.fields = indexs
	indexCount := 0
	//判断是否是主键索引
	if len(indexs) == len(t.primary) {
		for _, field := range indexs {
			if slices.Contains(t.primary, field) {
				indexCount++
			}
		}
	}
	isPrimary := indexCount == len(indexs)
	if isPrimary {
		idxs.kind = PrimaryKeyIndex
		return idxs
	}

	indexCount = 0
	// 统计全文索引的个数，即indexs存在fulltext的字段数
	for _, field := range indexs {
		if slices.Contains(t.fullText, field) {
			indexCount++
		}
	}
	isFullText := indexCount > 0
	switch isFullText {
	case true:
		idxs.kind = FullTextIndex
	case false:
		idxs.kind = NormalIndex
	}
	return idxs
}



// GetAllIndexPrefix 获取所有索引前缀，包括普通索引和全文索引
func (t *Table) GetKeys(fieldsBytes *map[string][]byte, matchesfield ...string) (keys [][]byte) {
	for _, indexs := range t.index {
		if len(matchesfield) != 0 { //是更新情况下
			for _, field := range indexs {
				if !slices.Contains(matchesfield, field) {
					return nil
				}
			}
		}
		indtype := t.MatchIndex(indexs)
		switch indtype.kind {
		case NormalIndex: // 普通索引
			keys = append(keys, t.GetIndexkey(fieldsBytes, indexs))
		case FullTextIndex: // 全文索引
			keys = append(keys, t.GetFullTextKey(fieldsBytes, indexs)...)
		default: // 默认主键
			continue
		}
	}
	return keys
}

// 获取全文索引值
// 只支持字符串类型的全文索引，其他类型的字段值会被转换为字符串。
// 匹配字段matchesfield ，用于记录修改某个字段时，只需要更新该字段的全文索引即可。
func (t *Table) GetFullTextKey(fieldsBytes *map[string][]byte, fulltext []string) (r [][]byte) {
	fullTextPrefix := []byte(t.GetFullTextPrefix())
	var fieldValue []byte
	var exists bool

	var curpfx [][]byte
	//curpfx = append(curpfx, fullTextPrefix)
	var isnil bool
	for _, field := range fulltext {
		// 获取字段的实际值
		fieldValue, exists = (*fieldsBytes)[field]
		if !exists {
			continue // 跳过不存在的字段
		}
		//该字段是否是全文索引字段
		isFullTextField := slices.Contains(t.fullText, field)
		switch isFullTextField {
		case true: // 全文索引字段
			// 将字段值转换为字符串，全文索引只支持字符串
			fieldStr := string(fieldValue)
			// 对字段值进行分词
			if t.ftlen == 0 {
				t.ftlen = 5
			}
			tokens := t.GetFullTextToken(fieldStr, int(t.ftlen))
			for _, token := range tokens {
				if token == "" {
					continue
				}
				curpfx = append(curpfx, bytes.Join([][]byte{fullTextPrefix, []byte(token)}, []byte(SPLIT)))
			}
		case false: // 普通索引字段
			curpfx = append(curpfx, bytes.Join([][]byte{fullTextPrefix, fieldValue}, []byte(SPLIT)))
		}
		//返回结果是否为空
		isnil = r == nil
		switch isnil {
		case true:
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
		fullTextPrefix = fullTextPrefix[:0]
	}
	return
}

// 获取索引值
// 匹配字段matchesfield ，用于记录修改某个字段时，只需要更新该字段的索引即可。
func (t *Table) GetIndexkey(fieldsBytes *map[string][]byte, indexs []string) (r []byte) {
	indexPrefix := []byte(t.GetIndexPrefix())
	idx := t.GetFieldJoin(fieldsBytes, indexs)
	r = bytes.Join([][]byte{indexPrefix, idx}, []byte(SPLIT))
	return
}

func (t *Table) GetPrimarykey(fieldsBytes *map[string][]byte) (r []byte) {
	PrimaryPrefix := []byte(t.GetPrimaryPrefix())
	pk := t.GetFieldJoin(fieldsBytes, t.IndexSet.PrimaryFields())
	r = bytes.Join([][]byte{PrimaryPrefix, pk}, []byte(SPLIT))
	return
}

// 获取字段值拼接
// NeedEscape 是否需要转义，默认不转义。只为用于组合主键的情况。
func (t *Table) GetFieldJoin(fieldsBytes *map[string][]byte, fields []string, NeedEscape ...bool) (r []byte) {
	var exists bool
	var fieldValue []byte
	var isnil bool
	//是否符合转义规则，NeedEscape为真，且字段数大于0。
	isNeedEscape := len(NeedEscape) > 0 && NeedEscape[0] == true && len(fields) > 0
	for _, index := range fields {
		// 获取字段值
		fieldValue, exists = (*fieldsBytes)[index]
		// 如果字段不存在或值为nil，跳过该索引
		if !exists || fieldValue == nil {
			break
		}
		//是否需要转义
		if isNeedEscape {
			fieldValue = Bytes(fieldValue).Escape()
		}
		isnil = r == nil
		switch isnil {
		case true:
			r = append(r, fieldValue...)
		case false:
			r = bytes.Join([][]byte{r, fieldValue}, []byte(SPLIT))
		}
	}
	return r
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
func (t *Table) SetPrimary(primary []string) error {
	t.IndexSet.PrimaryFields = primary
	return nil
}

// 设置索引字段
func (t *Table) SetIndex(index [][]string) error {
	t.IndexSet.Index = index
	return nil
}

// 添加单个索引
func (t *Table) AddIndex(index []string) error {
	t.IndexSet.Index = append(t.IndexSet.Index, index)
	return nil
}

// 设置全文索引字段
func (t *Table) SetFullText(fullText []string) error {
	t.FullTextIndex = fullText
	return nil
}

// 添加单个全文索引字段
func (t *Table) SetFullTextField(field string) error {
	t.FullTextIndex = append(t.FullTextIndex, field)
	return nil
}



// 获取所有字段值，用于添加记录时，直接复制，无需自行创建。
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
func (t *Table) RecordKV(fieldsBytes *map[string][]byte, batch *leveldb.Batch, put bool) error {
	//pfx := []byte(t.GetPrimaryPrefix())
	//key := bytes.Join([][]byte{pfx, t.GetPrimarykey(fieldsBytes)}, []byte(SPLIT))
	key := t.GetPrimarykey(fieldsBytes)
	switch put {
	case true:
		var buf bytes.Buffer
		var value []byte
		//记录格式：field1:value1-field2:value2-...-fieldN:valueN-
		for field, val := range *fieldsBytes {
			value = Bytes([]byte(field)).Escape()
			buf.Write(value)
			buf.WriteString(":")
			value = Bytes(val).Escape()
			buf.Write(value)
			buf.WriteString(SPLIT)
		}
		batch.Put(key, buf.Bytes())
	case false:
		batch.Delete(key)
	}
	return nil
}

// put索引KV
// 匹配字段ufield ，用于记录修改某个字段时，只需要更新该字段的索引即可。
func (t *Table) IndexsKV(fieldsBytes *map[string][]byte, batch *leveldb.Batch, put bool, ufield ...string) {
	idxs := t.GetKeys(fieldsBytes, ufield...)
	//只有创建索引的时候，才需要转义主键值
	value := t.GetFieldJoin(fieldsBytes, t.primary, true)
	if len(value) == 0 {
		fmt.Printf("主键 '%s' 的值为空，无法创建索引\n", t.primary[0])
	}
	for _, idx := range idxs {
		switch put {
		case true:
			batch.Put(idx, value)
		case false:
			batch.Delete(idx)
		}
	}
}
*/
// 获取单个字段值
func (t *Table) GetField(field string) (any, bool) {
	if t.fields == nil {
		return nil, false
	}
	value, exists := t.fields[field]
	return value, exists
}

// 检查类型是否匹配
func (t *Table) CheckType(fields *map[string]any) error {
	for field, value := range *fields {
		fieldValue, exists := t.GetField(field)
		// 只检查已经存在于表中的字段的类型
		if exists {
			// 检查类型是否匹配
			//主键可以是nil，其他字段不能是nil
			if field == t.indexs.PrimaryFields()[0].Field() && t.indexs.PrimaryFields()[0].Field() == "id" && value == nil {
				continue
			}
			if fieldValue != nil && reflect.TypeOf(fieldValue) != reflect.TypeOf(value) {
				return fmt.Errorf("字段 '%s' 的类型 '%T' 与提供的值 '%T' 类型不匹配", field, fieldValue, value)
			}
		} else {
			return fmt.Errorf("字段 '%s' 不存在于表中", field)
		}
	}
	return nil
}

// 将数据转换为字节数组
func (t *Table) FieldsToBytes(fields *map[string]any) map[string][]byte {
	result := make(map[string][]byte, len(*fields))
	for k, v := range *fields {
		result[k] = AnyToBytes(v)
	}
	return result
}

// 插入记录
func (t *Table) Insert(fields *map[string]any, batchs ...*leveldb.Batch) (currentID int, err error) {
	if t.fields == nil {
		return 0, fmt.Errorf("表 '%s' 未设置字段和类型", t.name)
	}
	//当前自动增值的值
	currentID = -1
	//是否支持默认自动增值主键
	supportDefault := len(t.indexs.PrimaryFields()) == 1 && t.indexs.PrimaryFields()[0].Field() == "id"
	if supportDefault {
		// 检查是否提供了主键字段
		//使用默认自动增值主键时，不需要提供主键字段，系统自动生成
		_, ok := (*fields)[t.indexs.PrimaryFields()[0].Field()]
		if !ok { //未提供主键字段，自动生成主键值
			currentID = t.GetAutoInc()
			(*fields)[t.indexs.PrimaryFields()[0].Field()] = currentID
		} else { //提供了主键字段，但是值为nil，自动生成主键值
			if (*fields)[t.indexs.PrimaryFields()[0].Field()] == nil {
				currentID = t.GetAutoInc()
				(*fields)[t.indexs.PrimaryFields()[0].Field()] = currentID
			} else { //提供了主键字段，且值不为nil，转换为int类型
				currentID = AnyToInt((*fields)[t.indexs.PrimaryFields()[0].Field()])
			}
		}
	}
	// 检查字段类型是否匹配
	if err = t.CheckType(fields); err != nil {
		return 0, err
	}
	// 转换字段为字节数组
	fieldsBytes := t.FieldsToBytes(fields)
	var batch *leveldb.Batch
	//是否用户手动控制事务
	useBatch := len(batchs) > 0
	if useBatch { //用户手动控制事务
		batch = batchs[0]
	} else {
		batch = GlobalBatchPool.Get()
		defer func() {
			GlobalBatchPool.Put(batch)
		}()
	}
	t.indexs.Joins(&fieldsBytes)
	return currentID, nil
}

/*
// 删除记录
func (t *Table) Delete(fields *map[string]any, batchs ...*leveldb.Batch) error {
	if t.fields == nil {
		return fmt.Errorf("表 '%s' 未设置字段和类型", t.name)
	}
	fieldsbyte := t.FieldsToBytes(fields)
	primaryValue := t.GetFieldJoin(&fieldsbyte, t.indexSet.PrimaryFields())
	if primaryValue == nil {
		return fmt.Errorf("更新操作必须提供主键字段 '%s'", t.indexSet.PrimaryFields()[0])
	}
	if err := t.CheckType(fields); err != nil {
		return err
	}
	// 转换字段为字节数组
	fieldsBytes := t.FieldsToBytes(fields)
	var batch *leveldb.Batch
	//是否用户手动控制事务
	useBatch := len(batchs) > 0
	if useBatch { //用户手动控制事务
		batch = batchs[0]
	} else {
		batch = GlobalBatchPool.Get()
		defer func() {
			GlobalBatchPool.Put(batch) // 放回对象池
		}()
	}
	t.RecordKV(&fieldsBytes, batch, false)
	t.IndexsKV(&fieldsBytes, batch, false)
	if !useBatch { //用户未手动控制事务，自动提交
		err := t.rsdb.Db.Write(batch, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// 更新记录
func (t *Table) Update(fields *map[string]any, batchs ...*leveldb.Batch) error {
	if t.fields == nil {
		return fmt.Errorf("表 '%s' 未设置字段和类型", t.name)
	}
	fieldsbyte := t.FieldsToBytes(fields)
	primaryValue := t.GetFieldJoin(&fieldsbyte, t.primary)
	if primaryValue == nil {
		return fmt.Errorf("更新操作必须提供主键字段 '%s'", t.primary[0])
	}

	//获取要修改的字段，用于更新索引和全文索引
	var matchesfield []string
	for field := range *fields {
		if field != t.primary[0] {
			matchesfield = append(matchesfield, field)
		}
	}
	// 读取旧记录
	record := t.Read(primaryValue)
	currentFields := t.ParseValue(record)
	if currentFields == nil {
		return fmt.Errorf("主键值 '%v' 的记录不存在", primaryValue)
	}
	var batch *leveldb.Batch
	//是否用户手动控制事务
	useBatch := len(batchs) > 0
	if useBatch { //用户手动控制事务
		batch = batchs[0]
	} else {
		batch = GlobalBatchPool.Get()
		defer func() {
			GlobalBatchPool.Put(batch) // 清空并放回对象池
		}()
	}
	oldfieldsBytes := t.FieldsToBytes(&currentFields)
	//获取要删除的旧索引和全文索引
	t.IndexsKV(&oldfieldsBytes, batch, false, matchesfield...)
	//更新字段值
	// 合并旧记录和新记录的字段值
	maps.Copy(currentFields, *fields)
	newfieldsBytes := t.FieldsToBytes(&currentFields)
	t.RecordKV(&newfieldsBytes, batch, true)
	t.IndexsKV(&newfieldsBytes, batch, true, matchesfield...)
	if !useBatch { //用户未手动控制事务，自动提交
		err := t.rsdb.Db.Write(batch, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// 从按主键数据库读取记录
func (t *Table) Read(primary any) []byte {
	if t.fields == nil {
		return nil
	}
	pb := AnyToBytes(primary)
	// 优化字符串拼接
	var key bytes.Buffer
	key.WriteString(t.GetPrimaryPrefix())
	key.WriteString(SPLIT)
	key.Write(pb)
	/*
		v, err := t.rsdb.Db.Get(key.Bytes(), nil)
		if err != nil {
			// leveldb.ErrNotFound 是正常的未找到错误，不需要打印
			if err != leveldb.ErrNotFound {
				fmt.Printf("读取记录失败: %v\n", err)
			}
			return nil
		}
	///
	return t.ReadByBytes(key.Bytes())
}

// 从按主键数据库读取记录
func (t *Table) ReadByBytes(key []byte) []byte {
	key = nil // bytes.Join([][]byte{[]byte(t.GetPrimaryPrefix()), key}, []byte(SPLIT))
	v, err := t.rsdb.Db.Get(key, nil)
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
	if record == nil || t.fields == nil {
		return nil
	}
	bs := Bytes(record).Split()
	fields := make(map[string]any, len(bs))
	field := ""
	for _, b := range bs {
		//以第一个':'为分隔符，将值分为两部分
		before, _, ok := bytes.Cut(b, []byte{':'})
		if ok {
			field = string(before)
			fields[field] = Bytes(b[len(field)+1:]).ToAny(t.fields[field])
		}
	}
	return fields
}

// 遍历表所有kv，复制表用
func (t *Table) For() iterator.Iterator {
	pfx := t.name + SPLIT
	return t.rsdb.GetIterator([]byte(pfx))
}

// 遍历表所有数据
func (t *Table) ForData() iterator.Iterator {
	pfx := t.name + SPLIT
	return t.rsdb.GetIterator([]byte(pfx))
}

/*
// 根据字段名和值搜索返回数据迭代器
// 缓存迭代器，避免每次for都重新创建迭代器
func (t *Table) Search(fields *map[string]any) (iterator.Iterator, *Indexs, error) {
	var field []string
	for k := range *fields {
		//判断字段是否在表中
		if _, ok := t.fields[k]; !ok {
			return nil, nil, fmt.Errorf("字段 '%s' 不存在于表 '%s'", k, t.name)
		}
		field = append(field, k)
	}
	idxType := t.MatchIndex(field)
	pfx := ""
	switch idxType.kind {
	case PrimaryKeyIndex: //主键索引
		pfx = t.GetPrimaryPrefix()
	case NormalIndex: //普通索引
		pfx = t.GetIndexPrefix()
	case FullTextIndex: //全文索引
		pfx = t.GetFullTextPrefix()
	case NoIndex: //没有索引则全表扫描
		pfx = t.name + SPLIT
	}
	keys := []byte(pfx) //[]byte(pfx + SPLIT)
	var val any
	var bval []byte
	var fval string
	for _, v := range idxType.fields {
		val = (*fields)[v]
		switch idxType.kind {
		case PrimaryKeyIndex:
			bval = AnyToBytes(val)
		case NormalIndex:
			bval = AnyToBytes(val)
		case FullTextIndex:
			fval = AnyToStr(val)
			// 全文索引只取前ftlen个字符
			if len([]rune(fval)) > int(t.ftlen) {
				fval = string([]rune(fval)[:t.ftlen])
			}
			bval = []byte(fval)
		}
		if len(bval) > 0 {
			keys = bytes.Join([][]byte{[]byte(keys), bval}, []byte(SPLIT))
		}
	}
	return t.rsdb.GetIterator(keys), idxType, nil
}

func (t *Table) SearchToDataIter(fields *map[string]any) *TableIter {
	iter, indexs, err := t.Search(fields)
	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return nil
	}
	return TableIterNew(t, iter, indexs)
}
*/
