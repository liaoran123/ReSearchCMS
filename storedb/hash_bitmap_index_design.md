# 哈希索引和位图索引设计方案

## 1. 哈希索引设计

### 1.1 概述

哈希索引是一种基于哈希表的索引结构，适合处理等值查询，具有O(1)的查询复杂度。哈希索引将键值通过哈希函数转换为哈希值，然后将哈希值映射到存储位置，实现快速查找。

### 1.2 数据模型

```go
// 哈希索引配置
type HashIndexOptions struct {
    FieldName   string            `json:"field_name"`   // 索引字段名
    HashFunc    string            `json:"hash_func"`    // 哈希函数类型，如"fnv1a", "murmur3", "crc32"
    BucketSize  int               `json:"bucket_size"`  // 桶大小
    Options     map[string]any    `json:"options"`      // 其他选项
}

// 哈希索引项
type HashIndexEntry struct {
    HashValue  []byte       `json:"hash_value"`  // 哈希值
    ObjectID   []byte       `json:"object_id"`   // 对象ID
    Value      any          `json:"value"`       // 字段值
}

// 哈希索引桶
type HashIndexBucket struct {
    Entries    []HashIndexEntry `json:"entries"`    // 桶中的索引项
    NextBucket []byte          `json:"next_bucket"` // 下一个桶的指针，用于解决哈希冲突
}
```

### 1.3 哈希函数实现

```go
// 哈希函数接口
type HashFunc interface {
    // 计算哈希值
    Hash(data []byte) []byte
    // 获取哈希函数名称
    Name() string
}

// FNV-1a哈希函数实现
type FNV1aHash struct{}

func (h *FNV1aHash) Hash(data []byte) []byte {
    // 实现FNV-1a哈希算法
    fnvPrime := uint64(1099511628211)
    offsetBasis := uint64(14695981039346656037)
    
    hash := offsetBasis
    for _, b := range data {
        hash ^= uint64(b)
        hash *= fnvPrime
    }
    
    return []byte(fmt.Sprintf("%016x", hash))
}

func (h *FNV1aHash) Name() string {
    return "fnv1a"
}
```

### 1.4 存储设计

在LevelDB中，哈希索引的存储格式：

```
// 索引元数据存储
键格式：[table_prefix]-hashidx-[field_name]-meta
值格式：序列化的HashIndexOptions

// 哈希桶存储
键格式：[table_prefix]-hashidx-[field_name]-[hash_value]-[bucket_id]
值格式：序列化的HashIndexBucket

// 示例
键：test_table-hashidx-username-meta
值：{"field_name":"username","hash_func":"fnv1a","bucket_size":100}

键：test_table-hashidx-username-5f4dcc3b5aa765d61d8327deb882cf99-0
值：{"entries":[{"hash_value":"5f4dcc3b5aa765d61d8327deb882cf99","object_id":"123","value":"admin"}],"next_bucket":null}
```

### 1.5 查询设计

```go
// 哈希查询参数
type HashQueryParams struct {
    FieldName  string   `json:"field_name"`  // 字段名
    Value      any      `json:"value"`       // 查询值
    Limit      int      `json:"limit"`       // 结果限制
    Skip       int      `json:"skip"`        // 跳过数量
}

// 哈希索引接口
type HashIndex interface {
    // 创建索引
    CreateIndex(table *Table, opts HashIndexOptions) error
    
    // 插入索引项
    Insert(objID []byte, value any) error
    
    // 删除索引项
    Delete(objID []byte, value any) error
    
    // 更新索引项
    Update(objID []byte, oldValue any, newValue any) error
    
    // 执行等值查询
    Query(params HashQueryParams) ([]HashIndexEntry, error)
    
    // 批量插入
    BatchInsert(entries []HashIndexEntry) error
    
    // 批量删除
    BatchDelete(entries []HashIndexEntry) error
    
    // 重建索引
    Rebuild() error
    
    // 获取索引统计信息
    Stats() (map[string]any, error)
}
```

## 2. 位图索引设计

### 2.1 概述

位图索引是一种基于位向量的索引结构，适合处理低选择性的字段（如性别、状态等枚举类型）。位图索引为每个唯一值创建一个位向量，每个位代表一个记录是否包含该值，支持高效的按位运算。

### 2.2 数据模型

```go
// 位图索引配置
type BitmapIndexOptions struct {
    FieldName   string            `json:"field_name"`   // 索引字段名
    MaxBits     int               `json:"max_bits"`     // 位向量最大长度
    Compress    bool              `json:"compress"`     // 是否压缩位向量
    Options     map[string]any    `json:"options"`      // 其他选项
}

// 位图索引项
type BitmapIndexEntry struct {
    Value      any          `json:"value"`       // 字段值
    Bitmap     []byte       `json:"bitmap"`      // 位向量，使用二进制格式存储
    Cardinality int         `json:"cardinality"`  // 基数，即该值出现的次数
}

// 位操作类型
type BitOp int

const (
    BitOpAND BitOp = iota // 与操作
    BitOpOR              // 或操作
    BitOpNOT             // 非操作
    BitOpXOR             // 异或操作
)
```

### 2.3 位向量实现

```go
// 位向量接口
type BitVector interface {
    // 设置指定位
    Set(bitIndex int) error
    
    // 清除指定位
    Clear(bitIndex int) error
    
    // 获取指定位的值
    Get(bitIndex int) (bool, error)
    
    // 执行位操作
    BitOp(op BitOp, other BitVector) (BitVector, error)
    
    // 获取位向量长度
    Length() int
    
    // 获取位向量中的设置位数量
    Count() int
    
    // 序列化
    MarshalBinary() ([]byte, error)
    
    // 反序列化
    UnmarshalBinary(data []byte) error
    
    // 压缩位向量
    Compress() ([]byte, error)
    
    // 解压缩位向量
    Decompress(data []byte) error
}

// 简单位向量实现
type SimpleBitVector struct {
    bits   []byte  // 位向量存储
    length int     // 位向量长度
}

func (bv *SimpleBitVector) Set(bitIndex int) error {
    if bitIndex >= bv.length {
        // 扩展位向量
        newLength := ((bitIndex / 8) + 1) * 8
        newBits := make([]byte, newLength/8)
        copy(newBits, bv.bits)
        bv.bits = newBits
        bv.length = newLength
    }
    
    byteIndex := bitIndex / 8
    bitOffset := bitIndex % 8
    bv.bits[byteIndex] |= 1 << bitOffset
    return nil
}

// 其他方法实现...
```

### 2.4 存储设计

在LevelDB中，位图索引的存储格式：

```
// 索引元数据存储
键格式：[table_prefix]-bitmapidx-[field_name]-meta
值格式：序列化的BitmapIndexOptions

// 位值映射存储
键格式：[table_prefix]-bitmapidx-[field_name]-values
值格式：序列化的映射表，存储值到值ID的映射

// 位图数据存储
键格式：[table_prefix]-bitmapidx-[field_name]-[value_id]
值格式：序列化的BitmapIndexEntry

// 示例
键：test_table-bitmapidx-gender-meta
值：{"field_name":"gender","max_bits":1000,"compress":true}

键：test_table-bitmapidx-gender-values
值：{"male":0,"female":1,"other":2}

键：test_table-bitmapidx-gender-0
值：{"value":"male","bitmap":"AQAA...","cardinality":500}
```

### 2.5 查询设计

```go
// 位图查询条件
type BitmapQueryCondition struct {
    Value      any      `json:"value"`       // 字段值
    Op         BitOp    `json:"op"`          // 位操作类型
}

// 位图查询参数
type BitmapQueryParams struct {
    FieldName   string                  `json:"field_name"`   // 字段名
    Conditions  []BitmapQueryCondition   `json:"conditions"`    // 查询条件
    Limit       int                     `json:"limit"`        // 结果限制
    Skip        int                     `json:"skip"`         // 跳过数量
}

// 位图索引接口
type BitmapIndex interface {
    // 创建索引
    CreateIndex(table *Table, opts BitmapIndexOptions) error
    
    // 插入索引项
    Insert(objID []byte, value any) error
    
    // 删除索引项
    Delete(objID []byte, value any) error
    
    // 更新索引项
    Update(objID []byte, oldValue any, newValue any) error
    
    // 执行位图查询
    Query(params BitmapQueryParams) ([]BitmapIndexEntry, error)
    
    // 批量插入
    BatchInsert(entries []BitmapIndexEntry) error
    
    // 批量删除
    BatchDelete(entries []BitmapIndexEntry) error
    
    // 重建索引
    Rebuild() error
    
    // 获取索引统计信息
    Stats() (map[string]any, error)
}
```

## 3. 与现有系统集成

### 3.1 扩展Table结构

```go
// 扩展Table结构，添加哈希索引和位图索引支持
type Table struct {
    // 现有字段...
    hashIndexes   map[string]HashIndex    // 哈希索引映射，key为字段名
    bitmapIndexes map[string]BitmapIndex  // 位图索引映射，key为字段名
}

// 添加哈希索引方法
func (t *Table) AddHashIndex(opts HashIndexOptions) error {
    // 根据配置创建哈希索引实例
    // 将索引添加到table.hashIndexes映射中
    return nil
}

// 添加位图索引方法
func (t *Table) AddBitmapIndex(opts BitmapIndexOptions) error {
    // 根据配置创建位图索引实例
    // 将索引添加到table.bitmapIndexes映射中
    return nil
}

// 哈希查询方法
func (t *Table) HashQuery(fieldName string, params HashQueryParams) ([]Record, error) {
    // 从hashIndexes中获取对应字段的哈希索引
    // 执行哈希查询
    // 根据查询结果获取完整记录
    return records, nil
}

// 位图查询方法
func (t *Table) BitmapQuery(fieldName string, params BitmapQueryParams) ([]Record, error) {
    // 从bitmapIndexes中获取对应字段的位图索引
    // 执行位图查询
    // 根据查询结果获取完整记录
    return records, nil
}
```

### 3.2 集成现有索引机制

```go
// 扩展Index结构，添加哈希索引和位图索引标识
type Index struct {
    // 现有字段...
    IsHash     bool             `json:"is_hash"`
    IsBitmap   bool             `json:"is_bitmap"`
    HashOptions HashIndexOptions  `json:"hash_options"`
    BitmapOptions BitmapIndexOptions `json:"bitmap_options"`
}
```

## 4. 实现步骤

### 4.1 哈希索引实现步骤

1. 实现哈希函数接口和具体哈希算法
2. 实现HashIndexEntry和HashIndexBucket结构体
3. 实现HashIndex接口，包括索引创建、插入、删除、更新和查询
4. 实现批量操作和重建索引功能
5. 集成到Table结构
6. 添加API接口
7. 测试和优化

### 4.2 位图索引实现步骤

1. 实现BitVector接口和具体位向量算法
2. 实现BitmapIndexEntry结构体
3. 实现BitmapIndex接口，包括索引创建、插入、删除、更新和查询
4. 实现位操作功能
5. 实现批量操作和重建索引功能
6. 集成到Table结构
7. 添加API接口
8. 测试和优化

## 5. 性能优化

### 5.1 哈希索引性能优化

1. **选择合适的哈希函数**：根据数据特征选择合适的哈希函数，减少哈希冲突
2. **动态扩容**：根据数据量动态调整桶大小，保持哈希表的负载因子在合理范围内
3. **哈希冲突处理**：使用链地址法或开放寻址法处理哈希冲突
4. **缓存热点数据**：缓存常用的哈希桶，减少磁盘访问
5. **批量操作**：支持批量插入、删除和更新，减少I/O次数

### 5.2 位图索引性能优化

1. **位向量压缩**：使用Run-Length Encoding（RLE）或其他压缩算法压缩位向量
2. **分段存储**：将大的位向量分段存储，提高查询效率
3. **预计算常用组合**：预计算常用查询条件的组合结果，提高查询效率
4. **并行计算**：支持并行执行位操作，提高查询速度
5. **位向量合并优化**：优化位向量的合并算法，减少计算时间

## 6. 应用场景

### 6.1 哈希索引应用场景

1. **等值查询**：如`SELECT * FROM users WHERE username = 'admin'`
2. **高选择性字段**：如主键、唯一标识符等
3. **频繁更新的字段**：如计数器、状态等
4. **点查询**：如用户登录、订单查询等
5. **缓存层**：作为其他索引的缓存，提高查询效率

### 6.2 位图索引应用场景

1. **低选择性字段**：如性别、状态、类型等枚举类型
2. **多条件组合查询**：如`SELECT * FROM users WHERE gender = 'male' AND status = 'active'`
3. **聚合查询**：如`SELECT COUNT(*) FROM users WHERE gender = 'female'`
4. **范围查询**：如`SELECT * FROM orders WHERE status IN (1, 2, 3)`
5. **分析查询**：如数据分析、数据挖掘等

## 7. 与其他索引的比较

| 索引类型 | 查询类型 | 时间复杂度 | 空间复杂度 | 适合字段 | 优点 | 缺点 |
|---------|---------|-----------|-----------|---------|------|------|
| 哈希索引 | 等值查询 | O(1) | O(n) | 高选择性字段 | 查询速度快，插入效率高 | 不支持范围查询，哈希冲突处理复杂 |
| 位图索引 | 等值查询、组合查询 | O(n/w)，w为机器字长 | O(n * v)，v为唯一值数量 | 低选择性字段 | 支持高效的组合查询，存储效率高 | 不适合高选择性字段，更新代价高 |
| 空间索引 | 空间查询 | O(log n) | O(n) | 空间数据 | 支持复杂的空间查询 | 实现复杂，查询代价较高 |

## 8. 注意事项

### 8.1 哈希索引注意事项

1. **哈希冲突**：选择合适的哈希函数和桶大小，减少哈希冲突
2. **哈希函数选择**：根据数据特征选择合适的哈希函数
3. **动态扩容**：实现动态扩容机制，避免哈希表负载过高
4. **并发安全**：确保多线程环境下的安全访问
5. **数据一致性**：确保索引与数据的一致性

### 8.2 位图索引注意事项

1. **字段选择性**：仅对低选择性字段使用位图索引
2. **更新代价**：位图索引的更新代价较高，适合读多写少的场景
3. **压缩策略**：选择合适的压缩策略，平衡存储效率和查询性能
4. **位向量长度**：合理设置位向量长度，避免浪费空间
5. **位操作优化**：优化位操作算法，提高查询效率

## 9. 未来扩展

1. **支持复合哈希索引**：支持多个字段的组合哈希索引
2. **支持自适应哈希索引**：根据查询模式自动调整哈希函数和桶大小
3. **支持布隆过滤器**：结合布隆过滤器，提高查询效率
4. **支持压缩位图索引**：实现更高效的位图压缩算法
5. **支持列存优化**：结合列存技术，提高查询效率
6. **支持分布式哈希索引**：实现分布式环境下的哈希索引
7. **支持分布式位图索引**：实现分布式环境下的位图索引

## 10. 总结

哈希索引和位图索引是两种重要的索引结构，分别适合不同的应用场景：

- **哈希索引**：适合处理等值查询，具有O(1)的查询复杂度，适合高选择性字段和频繁更新的场景
- **位图索引**：适合处理低选择性字段和多条件组合查询，支持高效的按位运算，适合分析查询和数据挖掘场景

通过实现这两种索引，可以丰富系统的索引类型，提高查询效率，满足不同场景的需求。与现有系统的集成设计保持了一致性，便于后续扩展和维护。