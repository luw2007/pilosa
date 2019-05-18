+++
title = "数据模型"
weight = 5
nav = [
    "简介",
    "Index",
    "Column",
    "Row",
    "Field",
    "Time Quantum",
    "Attribute",
    "Shard",
]
+++

## 数据模型

### 简介

Pilosa 数据模型的核心部分是布尔矩阵（boolean matrix）。矩阵中的每个单元都是单个位（bit）; 如果该位设置为1，则表示该特定行和列之间存在关系。

行和列可以表示任何内容（它们甚至可以表示与[bigraph](https://en.wikipedia.org/wiki/Bigraph)中相同的一组内容）。Pilosa 可以将任意键/值对（key/value 称为属性）与行和列相关联，但查询和存储围绕核心矩阵进行优化。

Pilosa 首先在行中布局数据，因此查询获得一行或多行中的所有设置位，或计算组合操作（例如：交集或多行的并集）是最快的。Pilosa 将行分类到不同的**字段**中，并快速检索字段中的顶行，这些行按每行中设置的列数排序。

请注意，当行和列 ID 从0开始按顺序排列时，Pilosa 的性能最高。您可以在某种程度上偏离这一点，但是在单节点集群上设置一个列ID为2<sup>63</sup>的位将由于内存限制导致无法正常工作。

![基本数据模型图](https://www.pilosa.com/img/docs/data-model.png)
*基本数据模型图*

### 索引

索引的目的是表示数据命名空间。您无法执行跨索引查询。

### 列

列ID是顺序的，增加整数，它们对索引中的所有字段都是通用的。单个列通常对应于关系表中的记录（尽管其他配置也是可能的）这样比较方便。

### 行

行ID是顺序的，增加了名称空间到索引中每个字段的整数。

### 字段

字段用于对索引中的行进行分段，例如用于定义不同的功能组。Pilosa 字段可以对应于关系表中的单个字段，其中标准 Pilosa 字段中的每一行表示关系字段的单个可能值。类似整数字段可以表示关系字段的所有可能的整数值。

#### 关系类比

Pilosa 索引是一种灵活的结构; 它可以代表任何类型的高基数（high-cardinality）二进制矩阵。我们在 Pilosa 用例中探索了许多建模模式; 一个可访问的例子是对关系模型的直接类比，在此总结。

实体:

 数据库       | Pilosa
-------------|----------------------------------------------
 Database    | N/A *(internal: Holder)*
 Table       | Index
 Row         | Column
 Column      | Field
 Value       | Row
 Value (int) | Field.Value (见 [BSI](#bsi-range-encoding))

简单查询:

 数据库                                         | Pilosa
-----------------------------------------------|------------------------------------
 `select ID from People where Name = 'Bob'`    | `Row(Name="Bob")`
 `select ID from People where Age > 30`        | `Row(Age > 30)`
 `select ID from People where Member = true`   | `Row(Member=0)`

请注意，`Row(Member=0)` 选择在 Member 字段的第0行中设置了1的所有实体。
我们也可以使用第1行来存储它，在这种情况下我们会使用`Row(Member=1)`，这看起来更直观。
在关系模型中，通常需要通过连接（joins）。由于 Pilosa 在行和列中都支持极高的基数，因此可以通过跨多个字段的基本 Pilosa 查询来完成许多类型的 join。例如下面这个SQL join：
```sql
select AVG(p.Age) from People p
inner join PersonCar pc on pc.PersonID=p.ID
inner join Cars c on pc.CarID=c.ID
where c.Make = 'Ford'
```

可以使用像这样的 Pilosa 查询来完成（注意[Sum](query-language.md#sum)返回一个包含 sum 和 count 的 json 对象，从中可以很容易地计算平均值）：

```pql
Sum(Row(Car-Make="Ford"), field=Age)
```

这是 Pilosa 的主要能力之一：结合多个数据存储关系的能力。

#### 排名

排名字段按行ID维护列计数的排序缓存（按列设置顶行，每个列设置一个位）。此缓存有助于 TopN 查询。缓存大小默认为50,000，可以在字段创建时设置。

![排名字段图](https://www.pilosa.com/img/docs/field-ranked.png)
* 排名字段图*

#### LRU

LRU高速缓存维护最近访问的行。

![lru 字段图](https://www.pilosa.com/img/docs/field-lru.png)
*LRU 字段图*

### 时间索引

在字段上设置时间索引会创建额外的视图，这些视图允许范围内的行查询缩短到指定的时间间隔。例如，如果时间索引设置为YMD，则支持范围内的行查询到一天的粒度。

### 属性

属性是可以与行或列关联的任意键/值对。此元数据存储在单独的BoltDB数据结构中。

列级属性在索引中很常见。也就是说，每个列属性应用于索引中所有字段的相应列中的所有位。行属性适用于相应行中的所有位。

### 分片

索引被分段成多组分片（shards以前称为slices）。每个分片包含固定数量的列，即 ShardWidth。ShardWidth是一个常量，只能在编译时修改，并且在摄取数据之前。默认值为2<sup>20</sup>。

查询操作并行运行，并通过一致的哈希算法在集群中均匀分布。

### 字段类型

字段在创建后会被为某种类型。Pilosa 支持以下字段类型：`set`, `int`, `bool`, `time`, `mutex`。

#### 集合 set

集合是Pilosa中的默认字段类型。设置字段表示行和列的标准二进制矩阵，其中每个行键表示可能的字段值。以下示例创建一个`set`名为“info” 的字段，其中包含最多100,000条记录的排名缓存。

``` request
curl localhost:10101/index/repository/field/info \
     -X POST \
     -d '{"options": {"type": "set", "cacheType": "ranked", "cacheSize":100000}}'
```
``` response
{"success":true}
```

#### 整数 int

类型字段`int`用于存储整数值。整数字段与索引中的其他字段共享相同的列，但该字段的值必须是介于创建字段时指定的值`min`和`max`值之间的整数。以下示例创建一个`int`名为“quantity” 的字段，该字段能够存储-1000到2000之间的值：

``` request
curl localhost:10101/index/repository/field/quantity \
     -X POST \
     -d '{"options": {"type": "int", "min": -1000, "max":2000}}'
```
``` response
{"success":true}
```

##### BSI 范围编码

编码位切片索引 (BSI: Bit-Sliced Indexing)是 Pilosa 用于表示位图索引中的多位整数的存储方法。整数存储为base-2的n位，范围编码位切片索引，以及指示“not null”的附加行。这意味着16位整数将需要17行：16位编码位切片组件的每个0位一个（1位不需要存储，因为范围编码最高位位置始终为1）和一个非空行。Pilosa 可以评估`Row`, `Min`, `Max`, 和 `Sum`查询这些BSI整数。`Sum`的查询结果包括 count，count 可用于计算没有其他开销的平均值。

内部 Pilosa 将每个BSI存储`field`为`view`。它们的行`view`包含整数值的base-2表示。Pilosa 管理base-2偏移和转换，有效地将整数值打包在最小行集内。

例如， `Set()` 针对BSI字段执行的以下查询将导致下图中描述的数据：

```
Set(1, A=1)
Set(2, A=2)
Set(3, A=3)
Set(4, A=7)
Set(2, B=1)
Set(3, B=6)
```

![BSI 字段图](https://www.pilosa.com/img/docs/field-bsi.png)
*BSI 字段图*

查看此 [博客文章](/blog/range-encoded-bitmaps/) 了解有关 Pilosa 中BSI的更多详细信息。

#### 时间 Time

时间字段类似于 `set` 字段，但除了行和列信息之外，它们还将每位时间值存储到定义的粒度。以下示例创建一个 `time` 名为“event” 的字段，该字段将时间戳信息存储到一天的粒度。

``` request
curl localhost:10101/index/repository/field/event \
     -X POST \
     -d '{"options": {"type": "time", "timeQuantum": "YMD"}}'
```
``` response
{"success":true}
```

对于`time`字段，将为每个定义的时间段生成数据视图。例如，对于时间量为的字段YMD，以下`Set()`查询将生成下图中描述的数据：

```
Set(3, A=8, 2017-05-18T00:00)
Set(3, A=8, 2017-05-19T00:00)
```

![时间索引字段图](https://www.pilosa.com/img/docs/field-time-quantum.png)
*时间索引字段图*

#### 互斥 Mutex

互斥字段与字段类似`set`，区别在于要求每列的行值互斥。换句话说，每列只能有一个字段值。如果在字段上更新列`mutex`的字段值，则将清除该列的上一个字段值。此字段类型类似于 RDBMS 表中的字段，其中每个记录包含特定字段的单个值。

#### 布尔 Boolean

布尔字段类似于`mutex`仅跟踪两个值的字段：`true`和`false`。布尔字段不维护已排序的缓存，也不支持键值。

