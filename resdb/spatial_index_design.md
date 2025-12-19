# 空间索引设计方案

## 1. 空间数据模型

### 1.1 基础几何类型

```go
// 点类型
type Point struct {
    Longitude float64 `json:"longitude"` // 经度
    Latitude  float64 `json:"latitude"`  // 纬度
}

// 线类型
type LineString struct {
    Points []Point `json:"points"` // 点集合
}

// 面类型
type Polygon struct {
    Rings []LineString `json:"rings"` // 环集合，第一个环为外环，其余为内环
}

// 最小边界矩形
type MBR struct {
    MinX float64 `json:"min_x"` // 最小X坐标
    MinY float64 `json:"min_y"` // 最小Y坐标
    MaxX float64 `json:"max_x"` // 最大X坐标
    MaxY float64 `json:"max_y"` // 最大Y坐标
}
```

### 1.2 空间对象接口

```go
// 空间对象接口
type SpatialObject interface {
    // 获取最小边界矩形
    GetMBR() MBR
    // 序列化
    MarshalBinary() ([]byte, error)
    // 反序列化
    UnmarshalBinary(data []byte) error
    // 转换为GeoJSON
    ToGeoJSON() ([]byte, error)
    // 检查是否包含点
    Contains(p Point) bool
    // 检查是否与其他空间对象相交
    Intersects(other SpatialObject) bool
    // 计算与其他空间对象的距离
    Distance(other SpatialObject) float64
}
```

## 2. 空间索引算法

### 2.1 基于Z-Order曲线的空间索引

Z-Order曲线（Z曲线）将多维空间坐标转换为一维值，适合在LevelDB的有序键空间中存储。

```go
// Z-Order曲线编码
type ZOrder struct {
    // 精度，默认为26位，可根据需求调整
    precision uint
}

// 将二维坐标编码为Z-Order值
func (z *ZOrder) Encode(x, y float64) uint64 {
    // 实现Z-Order编码逻辑
    // 1. 将坐标归一化到[0, 1]范围
    // 2. 将归一化坐标转换为二进制
    // 3. 交错合并二进制位
    // 4. 返回合并后的uint64值
    return zValue
}

// 将Z-Order值解码为二维坐标
func (z *ZOrder) Decode(zValue uint64) (float64, float64) {
    // 实现Z-Order解码逻辑
    return x, y
}
```

### 2.2 基于R树的空间索引

R树是一种多级平衡树，适合处理高维空间数据，支持高效的空间查询。

```go
// R树节点
type RTreeNode struct {
    MBR       MBR             // 节点的最小边界矩形
    IsLeaf    bool            // 是否为叶节点
    Children  []*RTreeNode    // 子节点（非叶节点）
    Entries   []RTreeEntry    // 索引项（叶节点）
    MaxEntries int           // 最大子节点数
}

// R树索引项
type RTreeEntry struct {
    MBR      MBR            // 空间对象的最小边界矩形
    ObjectID []byte         // 空间对象的唯一标识
    Object   SpatialObject  // 空间对象（可选，叶节点可只存储引用）
}
```

## 3. 存储设计

### 3.1 空间数据存储

在LevelDB中，空间数据的存储格式：

```
键格式：[table_prefix]-spatial-[field_name]-[object_id]
值格式：序列化的空间对象（如Point, LineString, Polygon）

示例：
test_table-spatial-location-123 → {"longitude":116.397, "latitude":39.907}
```

### 3.2 空间索引存储

#### 3.2.1 Z-Order索引存储

```
键格式：[table_prefix]-sidx-[field_name]-[z_order_value]-[object_id]
值格式：序列化的MBR和对象引用

示例：
test_table-sidx-location-0x123456-123 → {"min_x":116.39, "min_y":39.90, "max_x":116.40, "max_y":39.91}
```

#### 3.2.2 R树索引存储

```
键格式：[table_prefix]-rtree-[field_name]-[node_id]
值格式：序列化的R树节点

示例：
test_table-rtree-location-root → {"mbr":{...}, "children":[...], "entries":[...]}
```

## 4. 查询设计

### 4.1 空间查询类型

```go
// 空间查询类型
type SpatialQueryType int

const (
    QueryContains     SpatialQueryType = iota // 包含查询
    QueryIntersects                          // 相交查询
    QueryWithin                              // 在内部查询
    QueryDistance                            // 距离查询
    QueryNearest                             // 最近邻查询
)

// 空间查询参数
type SpatialQueryParams struct {
    QueryType   SpatialQueryType // 查询类型
    Geometry    SpatialObject     // 查询几何对象
    Distance    float64           // 距离阈值（用于距离查询）
    Limit       int               // 返回结果数量限制
    Skip        int               // 跳过结果数量
}
```

### 4.2 查询API设计

```go
// 空间查询接口
type SpatialIndex interface {
    // 创建空间索引
    CreateIndex(table *Table, fieldName string) error
    
    // 插入空间对象
    Insert(objID []byte, obj SpatialObject) error
    
    // 删除空间对象
    Delete(objID []byte) error
    
    // 更新空间对象
    Update(objID []byte, obj SpatialObject) error
    
    // 执行空间查询
    Query(params SpatialQueryParams) ([]*RTreeEntry, error)
    
    // 批量插入
    BatchInsert(entries []*RTreeEntry) error
    
    // 批量删除
    BatchDelete(objIDs [][]byte) error
    
    // 重建索引
    Rebuild() error
    
    // 获取索引统计信息
    Stats() (map[string]any, error)
}
```

## 5. 集成现有系统

### 5.1 扩展Table结构

```go
// 扩展Table结构，添加空间索引支持
type Table struct {
    // 现有字段...
    spatialIndexes map[string]SpatialIndex // 空间索引映射，key为字段名
}

// 添加空间索引
type SpatialIndexOptions struct {
    FieldName   string            `json:"field_name"`
    IndexType   string            `json:"index_type"`   // zorder, rtree
    Precision   uint              `json:"precision"`   // 仅Z-Order索引使用
    MaxEntries  int               `json:"max_entries"` // 仅R树索引使用
    Options     map[string]any    `json:"options"`
}

// 添加空间索引方法
func (t *Table) AddSpatialIndex(opt SpatialIndexOptions) error {
    // 根据索引类型创建不同的空间索引实例
    // 将索引添加到table.spatialIndexes映射中
    return nil
}

// 空间查询方法
func (t *Table) SpatialQuery(fieldName string, params SpatialQueryParams) ([]Record, error) {
    // 从spatialIndexes中获取对应字段的空间索引
    // 执行空间查询
    // 根据查询结果获取完整记录
    return records, nil
}
```

### 5.2 集成现有索引机制

```go
// 在现有索引结构中添加空间索引标识
type Index struct {
    // 现有字段...
    IsSpatial bool `json:"is_spatial"`
    SpatialOptions SpatialIndexOptions `json:"spatial_options"`
}
```

## 6. 实现步骤

1. **实现空间数据模型**：定义Point, LineString, Polygon等几何类型，实现SpatialObject接口
2. **实现MBR计算**：为每个几何类型实现最小边界矩形计算
3. **选择索引算法**：根据需求选择Z-Order或R树索引
4. **实现索引构建**：实现索引的插入、删除、更新操作
5. **实现查询功能**：实现各种空间查询（包含、相交、距离等）
6. **集成到Table**：扩展Table结构，添加空间索引支持
7. **添加API接口**：为用户提供空间查询的API
8. **测试和优化**：测试空间查询性能，根据需要进行优化

## 7. 性能优化

1. **批量操作**：支持批量插入、删除、更新，减少I/O次数
2. **缓存机制**：缓存常用的索引节点，减少磁盘访问
3. **并行查询**：对于复杂查询，考虑并行处理
4. **索引压缩**：对索引数据进行压缩，减少存储开销
5. **查询优化**：根据查询类型选择最优的索引遍历方式
6. **自适应索引**：根据数据分布动态调整索引参数

## 8. 应用场景

1. **地理位置搜索**：基于位置的服务（LBS）
2. **地理围栏**：检测点是否在指定区域内
3. **路径规划**：查找最优路径
4. **空间分析**：分析空间数据分布
5. **地图可视化**：高效渲染地图数据
6. **物流配送**：优化配送路线

## 9. 与现有插件机制集成

```go
// 空间索引插件
type SpatialIndexPlugin struct {
    // 插件实现
}

// 实现TablePlugin接口
func (p *SpatialIndexPlugin) Name() string {
    return "spatial_index"
}

func (p *SpatialIndexPlugin) Initialize(table *Table) error {
    // 初始化空间索引
    return nil
}

func (p *SpatialIndexPlugin) OnCreate() error {
    // 表创建时的处理
    return nil
}

func (p *SpatialIndexPlugin) OnUpdate() error {
    // 表更新时的处理
    return nil
}

func (p *SpatialIndexPlugin) OnDelete() error {
    // 表删除时的处理
    return nil
}
```

## 10. 注意事项

1. **坐标系统**：统一使用WGS84坐标系统（经度范围-180~180，纬度范围-90~90）
2. **精度问题**：浮点数比较需要考虑精度误差
3. **边界情况**：处理好空间对象的边界情况
4. **并发安全**：确保空间索引的并发安全
5. **事务支持**：考虑如何与LevelDB的事务机制集成
6. **数据一致性**：确保空间数据与索引的一致性
7. **性能测试**：针对不同数据量进行性能测试，调整索引参数

## 11. 未来扩展

1. **支持更多几何类型**：如MultiPoint, MultiLineString, MultiPolygon等
2. **支持三维空间**：扩展到三维空间索引
3. **支持时间维度**：实现时空索引，支持时空查询
4. **分布式支持**：支持分布式空间索引
5. **GPU加速**：考虑使用GPU加速空间计算
6. **与现有GIS系统集成**：支持导入导出常见GIS格式（如Shapefile, GeoJSON等）

## 12. 总结

空间索引设计需要综合考虑数据模型、索引算法、存储方式和查询接口等多个方面。基于LevelDB的特性，推荐使用Z-Order曲线或R树作为空间索引算法，Z-Order实现简单，适合点数据和简单查询；R树功能强大，适合复杂空间查询。

通过扩展现有Table结构，添加空间索引支持，可以实现高效的空间查询，满足地理位置搜索、地理围栏、路径规划等应用场景的需求。