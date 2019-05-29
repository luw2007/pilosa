+++
title = "查询语言"
weight = 6
nav = [
    "Conventions",
    "Arguments and Types",
    "Write Operations",
    "Read Operations",
]
+++

## 查询语言

### 简介

本节将提供Pilosa查询语言（PQL）的详细参考和示例. 所有PQL查询都在单个[索引](glossary.md#index)上运行，并通过 `/index/INDEX_NAME/query` 接口传输给 Pilosa. 
您可以通过简单地将查询连接在一起来在单个请求中传递多个PQL查询（不需要空格）。结果格式始终如下：

```
{"results":[...]}
```

对于请求中的每个PQL查询，`results` 数组中将有一个结果。数组中每个项的类型取决于查询的类型，下面引用中的每个查询都列出了它的结果类型。

#### 约定

* Angle Brackets `<>` 表示必需的参数
* Square Brackets `[]` 表示可选参数
* UPPER_CASE 表示的描述符，将需要用一个具体数值来填充（例如: `ATTR_NAME`，`STRING`）

##### 例子

在运行下面的任何示例查询之前，请按照[入门指南](getting-started.md)部分中的说明设置索引和字段，并使用一些数据填充它们。

这些示例只显示了所需的PQL查询 - `Set(10, stargazer=1)`使用 curl 对服务器运行查询，您将：
``` request
curl localhost:10101/index/repository/query \
     -X POST \
     -d 'Set(10, stargazer=1)'
```
``` response
{"results":[true]}
```

#### 参数和类型

* `field` 该字段指定查询将在哪个 Pilosa [字段](glossary.md#field) 运行。有效字段名称是小写字符串; 它们以字母数字字符开头，仅包含字母数字字符和`_-`。它们的长度不得超过64个字符。
* `TIMESTAMP` 这是以下格式的时间 `YYYY-MM-DDTHH:MM` (例如：2006-01-02T15:04)
* `UINT` 无符号整数 (例如：42839)
* `BOOL` 布尔值, `true` 或 `false`
* `ATTR_NAME` 必须是有效的标识符 `[A-Za-z][A-Za-z0-9._-]*`
* `ATTR_VALUE` 可以是字符串，浮点数，整数或布尔值
* `CALL` 任何查询
* `ROW_CALL` 任何查询返回的行，如 `Row`, `Union`, `Difference`, `Xor`, `Intersect`, `Not`
* `[]ATTR_VALUE` 表示一组 `ATTR_VALUE` (例如：`["a", "b", "c"]`)

### 写操作

#### 集合 Set

**Spec:**

```
Set(<COLUMN>, <FIELD>=<ROW>, [TIMESTAMP])
```

**Description:**

`Set` 将值1赋给二进制矩阵中的一个位，从而将给`<ROW>`定字段中给定的行（值）与给定的列相关联。

<div class="note">
    <p>虽然在PQL中使用“Set”是熟悉Pilosa的一种便捷方式，但使用<a href="https://github.com/pilosa/go-pilosa/blob/master/docs/imports-exports.md">Go</a>，<a href="https://github.com/pilosa/java-pilosa/blob/master/docs/imports.md">Java</a>和<a href="https://github.com/pilosa/python-pilosa/tree/master/docs/imports.md">Python</a>客户端中的导入功能来获取大量数据几乎总是更好。</p>
</div>

**Result Type:** boolean

返回值`true`表示该位已更改为1。

返回值`false`表示该位已设置为1且未更改。


**Examples:**

设置第1行第10列的位：
```request
Set(10, stargazer=1)
```
```response
{"results":[true]}
```

这在 stargazer 字段中设置了一个位，表示用户1为项目10点赞。

Set 还支持提供时间戳。需要添加用户为项目点赞的日期：
```request
Set(10, stargazer=1, 2016-01-01T00:00)
```
```response
{"results":[true]}
```

在单个请求中设置多个位：
```request
Set(10, stargazer=1) Set(20, stargazer=1) Set(10, stargazer=2) Set(30, stargazer=2)
```
```response
{"results":[false,true,true,true]}
```

在第10列将字段“pullrequests”设置为整数值2：
```request
Set(10, pullrequests=2)
```
```response
{"results":[true]}
```

#### SetRowAttrs
**Spec:**

```
SetRowAttrs(<FIELD>, <ROW>,
            <ATTR_NAME=ATTR_VALUE>,
            [ATTR_NAME=ATTR_VALUE ...])
```

**Description:**

`SetRowAttrs` 将任意键/值对与字段中的行相关联。设置值`null`不带引号，则删除属性。

**Result Type:** null

SetRowAttrs 查询总是`null`在成功时返回。

**Examples:**

在第10行设置属性`username`和`active`：
```request
SetRowAttrs(stargazer, 10, username="mrpi", active=true)
```
```response
{"results":[null]}
```
为用户10设置用户名和活动状态。这些是对 Pilosa 没有意义的任意键/值对。您可以使用[Row](query-language.md#row)查询查看您在行上设置的属性`Row(stargazer=10)`。

删除第10行的`username`属性：
```request
SetRowAttrs(stargazer, 10, username=null)
```
```response
{"results":[null]}
```

#### SetColumnAttrs

**Spec:**

```
SetColumnAttrs(<COLUMN>,
               <ATTR_NAME=ATTR_VALUE>,
               [ATTR_NAME=ATTR_VALUE ...])
```

**Description:**

`SetColumnAttrs` 将任意键/值对与索引中的列相关联。

**Result Type:** null

SetColumnAttrs 查询总是`null`在成功时返回。设置值为`null`不带引号，则删除属性。

**Examples:**

设置第10列的属性`stars`,`url`和`active`：

```request
SetColumnAttrs(10, stars=123, url="http://projects.pilosa.com/10", active=true)
```
```response
{"results":[null]}
```
设置项目10的 url 值和活动状态。这些是对 Pilosa 没有意义的任意键/值对。

可以通过将 URL 参数添加`columnAttrs=true`到查询来请求 ColumnAttrs 。例如：

```request
curl localhost:10101/index/repository/query?columnAttrs=true -XPOST -d 'Row(stargazer=1) Row(stargazer=2)'
```
```response
{
  "results":[
    {"attrs":{},"cols":[10,20]},
    {"attrs":{},"cols":[10,30]}
  ],
  "columnAttrs":[
    {"id":10,"attrs":{"active":true,"stars":123,"url":"http://projects.pilosa.com/10"}},
    {"id":20,"attrs":{"active":false,"stars":456,"url":"http://projects.pilosa.com/30"}}
  ]
}
```
在此示例中，ColumnAttrs 已设置在第10列和第20列，但未设置在第30列。相关属性都在单个columnAttrs列表中返回。有关更多信息，请参阅[查询索引](api-reference.md#query-index)部分。

删除第10列上的`url`属性：
```request
SetColumnAttrs(10, url=null)
```
```response
{"results":[null]}
```

#### Clear

**Spec:**

```
Clear(<COLUMN>, <FIELD>=<ROW>)
```

**Description:**

`Clear` 为二进制矩阵中的一个位赋值为0，从而解除给定字段中给定行与给定列的关联。

请注意，清除时间字段上的列将删除该列的所有数据。

**Result Type:** boolean

返回值`true`表示该位从1切换为0。

返回值`false`表示该位已设置为0且未更改任何内容。

**Examples:**

清除第1行的第10列的 stargazer 字段：

```request
Clear(10, stargazer=1)
```
```response
{"results":[true]}
```
这表示删除用户1与项目10的关系。

#### ClearRow

**Spec:**

```
ClearRow(<FIELD>=<ROW>)
```

**Description:**

`ClearRow` 在二进制矩阵的给定行中将所有位设置为 0，从而使给定字段中的给定行与所有列解除关联。

**Result Type:** boolean

返回值`true`表示至少有一列从1切换为0。

返回值`false`表示该行中的所有位都已经为0且没有任何更改。

**Examples:**

清除 stargazer 字段中第1行的所有位：
```request
ClearRow(stargazer=1)
```
```response
{"results":[true]}
```

这表示删除用户1与所有项目之间的关系。


#### Store

**Spec:**

```
Store(<ROW_CALL>, <FIELD>=<ROW>)
```

**Description:**

`Store` 将结果写入`<ROW_CALL>` 指定的行。如果该行已存在，则将替换该行。目标字段必须是字段类型 `set`.

**Result Type:** boolean

成功后，此方法始终返回`true`。未来版本的 Pilosa 可以使用此布尔结果来指示目标行中的数据是否已被`Store`调用更改。

**Examples:**

将用户1的数据存储给用户2:
```request
Store(Row(stargazer=1), stargazer=2)
```
```response
{"results":[true]}
```

将用户10和用户11都点赞的项目存储给用户20。
```request
Store(Intersect(Row(stargazer=10), Row(stargazer=11)), stargazer=20)
```
```response
{"results":[true]}
```

### Read Operations

#### Row

**Spec:**

```
Row(<FIELD>=<ROW>)
```

**Description:**

`Row` 检索一行中所有列的索引。它还检索该行上设置的任何属性。

**Result Type:** object with attrs and columns.

例如：`{"attrs":{"username":"mrpi","active":true},"columns":[10, 20]}`

**Examples:**

使用字段第1行中设置的位查询所有列`stargazer`（用户1点赞过的项目）：

Query all columns with a bit set in row 1 of the field `stargazer` (repositories that are starred by user 1):
```request
Row(stargazer=1)
```
```response
{"attrs":{"username":"mrpi","active":true},"columns":[10, 20]}
```

* attrs 是用户1的属性
* columns 是用户1已经点赞的项目列表.


#### Row (Range)

**Spec:**

```
Row(<FIELD>=<ROW>, from=<TIMESTAMP>, to=<TIMESTAMP>)
```

**Description:**

类似 `Row`, 但仅返回在给定 `from` (包含) 和 `to` (不包含) 之间设置过时间的位. 两者 `from` 和 `to` 参数都是可选项. 默认时间 `to` 是当前时间 +1 天。如果需要更高的结束时间，请具体指定。

**Result Type:** object with attrs and bits


**Examples:**

在日期范围内查询在字段的第1行中设置位的所有列（用户已点赞的项目）：

```request
Row(stargazer=1, from='2010-01-01T00:00', to='2017-03-02T03:00')
```
```response
{{"attrs":{},"columns":[10]}
```

此示例假定已在某些位上设置了时间戳。

* 列是在 2010-01-01 到 2017-03-02 的时间范围内由用户1点赞的项目。


#### Row (BSI)

**Spec:**

```
Row([<COMPARISON_VALUE> <COMPARISON_OPERATOR>] <FIELD> <COMPARISON_OPERATOR> <COMPARISON_VALUE>)
```

**Description:**

 `Row(BSI)` 支持 `integer` 类型和 `timestamp` 类型.
返回运算结果为true的行。

**Result Type:** object with attrs and columns

**Examples:**

在我们的数据中，计算提交比较频繁的项目。
下面是一个大于`Row`的例子，返回 commitactivity 大于100的所有列（项目超过100次提交）：

```request
Row(commitactivity > 100)
```
```response
{{"attrs":{},"columns":[10]}
```

* columns 列是在去年至少有100次提交的项目.

BSI 范围查询支持以下运算符：

 操作符 | 名称                 | 类型
----------|------------------|--------------------
 `>`      | 大于, GT          | integer
 `<`      | 小于, LT          | integer
 `<=`     | 小于或等于, LTE    | integer
 `>=`     | 大于或等于, GTE    | integer
 `==`     | 等于, EQ          | integer
 `!=`     | 不等于, NEQ       | integer or `null`

可以通过联合`<`和`<=`运算符（但不包括其他）来指定有界区间。例如：

```request
Row(50 < commitactivity < 150)
```
```response
{{"attrs":{},"columns":[10]}
```
从Pilosa 1.0开始，不再支持`Row(frame=stats, commitactivity >< [50, 150])`“between”语法。

#### Union

**Spec:**

```
Union([ROW_CALL ...])
```

**Description:**

Union 对`ROW_CALL`传递给它的所有查询的结果执行逻辑 OR 。

**Result Type:** object with attrs and bits

attrs 永远为空

**Examples:**

查询列中设置了两行中的任何一行（由两个用户中的任何一个点赞的项目）：
```request
Union(Row(stargazer=1), Row(stargazer=2))
```
```response
{"attrs":{},"columns":[10, 20, 30]}
```

* 列是由用户1 OR 用户2点赞的项目

#### Intersect

**Spec:**

```
Intersect(<ROW_CALL>, [ROW_CALL ...])
```

**Description:**

Intersect 对`ROW_CALL`传递给它的所有查询的结果执行逻辑 AND 

**Result Type:** object with attrs and columns

attrs 永远为空

**Examples:**

查询两列中都设置了位的列（由两个用户共同加载的项目）：


```request
Intersect(Row(stargazer=1), Row(stargazer=2))
```
```response
{"attrs":{},"columns":[10]}
```

* 列是由用户1 AND 用户点赞的项目

#### Difference

**Spec:**

```
Difference(<ROW_CALL>, [ROW_CALL ...])
```

**Description:**

Difference 返回的第一个 `ROW_CALL` 参数的所有位，而不是后续的 `ROW_CALL`.

**Result Type:** object with attrs and columns

attrs 永远为空

**Examples:**

查询列中的位设置在一行而不是另一行（由一个用户而不是另一个用户赞点赞的项目）：
```request
Difference(Row(stargazer=1), Row(stargazer=2))
```
```response
{"results":[{"attrs":{},"columns":[20]}]}
```

* columns are 是用户1点赞，并且用户2没有点赞的项目

反过来可以查询：
```request
Difference(Row(stargazer=2), Row(stargazer=1))
```
```response
{"attrs":{},"columns":[30]}
```

* columns 是由用户2点赞，但用户1没有点赞的项目

#### Xor

**Spec:**

```
Xor(<ROW_CALL>, [ROW_CALL ...])
```

**Description:**

Xor 对`ROW_CALL`传递给它的每个查询的结果执行逻辑 XOR 。

**Result Type:** object with attrs and columns

attrs 永远为空

**Examples:**

查询列中只有两行中的一行（仅由两个用户中的一个点赞的项目）：

```request
Xor(Row(stargazer=2), Row(stargazer=1))
```
```response
{"results":[{"attrs":{},"columns":[20,30]}]}
```

* columns 是由用户1 XOR 用户2（用户1或用户2，但不是同时）点赞的项目

#### Not

**Spec:**

```
Not(<ROW_CALL>)
```

**Description:**

Not 返回`ROW_CALL`参数中所有位的反转。Not 查询要求`trackExistence`已在索引上启用。


**Result Type:** object with attrs and columns

attrs 永远为空

**Examples:**

查询在给定行中没有设置位的所有列。

```request
Not(Row(stargazer=1))
```
```response
{"results":[{"attrs":{},"columns":[30]}]}
```

* columns 是用户1没有点赞的项目

#### Count
**Spec:**

```
Count(<ROW_CALL>)
```

**Description:**

Count 返回`ROW_CALL`设置1的数量.

**Result Type:** int

**Examples:**

查询一行中设置的位数（用户已点赞的项目数）：

```request
Count(Row(stargazer=1))
```
```response
{"results":[1]}
```

* Result 是用户1已经点赞的项目数。

#### Shift
**Spec:**

```
Shift(<ROW_CALL>, [n=UINT])
```

**Description:**

返回由`ROW_CALL`位移`n`位的行。


**Result Type:** object with attrs and columns

attrs 永远为空

**Examples:**

查询用户1点赞项目位移 2 所有项目
```request
Shift(Row(stargazer=1), n=2)
```
```response
{"attrs":{},"columns":[12, 22]}
```

* columns 用户1点赞并位移2位的项目.

#### TopN

**Spec:**

```
TopN(<FIELD>, [ROW_CALL], [n=UINT],
     [attrName=<ATTR_NAME>, attrValues=<[]ATTR_VALUE>])
```

**Description:**

返回`n`字段中顶行的ID和计数（按位数）。`attrName`和`attrValues`同时起作用，仅当指定的`attrName`的值在`attrValues`中。

**Result Type:** array of key/count objects

**Caveats:**

* 对具有已排名缓存类型的字段执行TopN（）查询将返回按count按降序排序的顶行。
* 具有缓存类型lru的字段将维护 LRU（最近最少使用的替换策略）缓存，因此对此类型字段的TopN查询将返回按最近设置的位的顺序排序的行。
* 字段的高速缓存大小决定了为了TopN查询而在高速缓存中维护的已排序行数。在性能和准确性之间进行权衡; 增加高速缓存大小将以性能为代价提高结果的准确性。
* 一旦填满，缓存将根据字段选项CacheSize截断行集。跨越限制并具有相同计数的行将被截断，没有特定的顺序。
* TopN查询的属性过滤器应用于现有的行排序缓存。排除在排序缓存范围之外的行，即使它们通常会通过过滤器，也会被忽略。

查阅 [字段创建](api-reference.md#create-field)，了解更多缓存信息.

**Examples:**

基本 TopN 查询:
```request
TopN(stargazer)
```
```response
{"results":[[{"id":1240,"count":102},{"id":4734,"count":100},{"id":12709,"count":93},...]]}
```

* `id` 是行 ID (用户 ID)
* `count` 列数（项目数）
* 结果是在 stargazer 字段中所有行（用户）按降序排列的相应行（每个用户点赞的项目）中设置的位数。例如，用户1240点赞了102个项目，用户4734点赞了100个项目，用户12709点赞了93项目。

限制结果数量：
```request
TopN(stargazer, n=2)
```
```response
{"results":[[{"id":1240,"count":102},{"id":4734,"count":100}]]}
```

* 结果是前两行（用户）按降序排列的位数（它们已点赞的项目）排序。

根据现有行过滤：
```request
TopN(stargazer, Row(language=1), n=2)
```
```response
{"results":[[{"id":1240,"count":35},{"id":7508,"count":32}]]}
```

* 结果是按照与语言1的交集中设置的位数排序的前两个用户（行）（他们点赞的项目以语言1编写）。

根据属性过滤：
```request
TopN(stargazer, n=2, attrName=active, attrValues=[true])
```
```response
{"results":[[{"id":10,"count":1},{"id":13,"count":1}]]}
```

* 结果是将“active”属性设置为“true”的前两个用户（行），按设1的位数（他们点赞的项目）排序。

#### Min

**Spec:**

```
Min([ROW_CALL], field=<FIELD>)
```

**Description:**

返回此`field`中所有 BSI 整数值的最小值。如果提供了可选`Row`调用，则仅考虑具有设置位的列，否则将考虑所有列。

**Result Type:** object with the min and count of columns containing the min value.

**Examples:**

查询字段的最小值（所有项目的空间最小）：
```request
Min(field="diskusage")
```
```response
{"value":4,"count":2}
```

* 结果是最小值（项目大小，以千字节为单位），加上具有该值的列数。

#### Max

**Spec:**

```
Max([ROW_CALL], field=<FIELD>)
```

**Description:**

返回此`field`中所有BSI整数值的最大值。如果提供了可选`Row`调用，则仅考虑具有设置位的列，否则将考虑所有列。

**Result Type:** object with the max and count of columns containing the max value.

**Examples:**

查询字段的最大值（所有项目的空间最大）：
```request
Max(field="diskusage")
```
```response
{"value":88,"count":13}
```

* 结果是最大值（项目大小，以千字节为单位），加上具有该值的列数。

#### Sum

**Spec:**

```
Sum([ROW_CALL], field=<FIELD>)
```

**Description:**

返回此`field`中所有BSI整数值的计数和计算总和。如果提供了可选`Row`调用，则仅考虑具有设置位的列，否则将考虑所有列。

**Result Type:** object with the computed sum and count of the values in the integer field.

**Examples:**

统计所有项目的大小：
```request
Sum(field="diskusage")
```
```response
{"value":10,"count":3}
```

* 结果是空间之和（项目大小，以千字节为单位），加上列数。

### Other Operations

#### Options

**Spec:**

```
Options(<CALL>, columnAttrs=<BOOL>, excludeColumns=<BOOL>, excludeRowAttrs=<BOOL>, shards=[UINT ...])
```

**Description:**

修改给定的查询，如下所示：

* `columnAttrs`：在结果中包含列属性（默认值: `false`)。
* `excludeColumns`：从结果中排除列ID（默认值: `false`)。
* `excludeRowAttrs`：从结果中排除行属性（默认值: `false`) 。
* `shards`：仅使用给定分片中的数据运行查询。默认情况下，使用整个数据集（即来自所有分片的数据）。

**Result Type:** Same result type as `<CALL>`.

**Examples:**

返回列属性：
```request
Options(Row(f1=10), columnAttrs=true)
```
```response
{"attrs":{},"columns":[100]}],"columnAttrs":[{"id":100,"attrs":{"foo":"bar"}}
```

仅针对分片0和2运行查询：
```request
Options(Row(f1=10), shards=[0, 2])
```
```response
{"attrs":{},"columns":[100, 2097152]}
```

#### Rows

**Spec:**

```
Rows(<FIELD>, previous=<UINT|STRING>, limit=<UINT>, column=<UINT|STRING>, from=<TIMESTAMP>, to=<TIMESTAMP>)
```

**Description:**


行返回给定字段中的行 ID 列表，其中至少有一个位设置。字段参数是必需的，其他参数是可选的。

如果`previous`给定，则不会返回指定行 ID 或键之前和之后的行。如果`column`给出，则仅返回给定列中具有设置位的行。`previous`或者`column`必须是字符串当且仅当字段或索引分别使用键转换时。如果`limit`给出，则返回的 rowID 数将小于或等于`limit`。组合`limit`并`previous`允许对大型结果集进行分页。结果始终是有序的，因此设置`previous`为上一个请求的最后结果将从下一个可用行开始。

如果字段是`time`类型，则可以提供`from`和`to`参数以将结果限制为特定时间跨度。如果`from`和`to`未提供，将查询所有现有数据。

**Result Type:** Object with `"rows" or "keys" and an array of integers or strings respectively.`

**Examples:**

Without keys:
```request
Rows(blah)
```
```response
{"rows":[1,9,39]}
```

With keys:
```request
Rows(blahk)
```
```response
{"rows":null,"keys":["haha","zaaa","traa"]}
```

#### Group By

**Spec:**

```
GroupBy(<RowsCall>, [RowsCall...], limit=<UINT>, filter=<CALL>)
```

**Description:**

GroupBy 返回从指定`Rows`调用中获取一行的每个行组合的交集计数。它仅返回计数大于 0 的那些组合。

可选`filter`参数采用任何类型的`Row`查询（例如，Row，Union，Intersect等），它们将在返回计数之前与每个结果相交。这与应用于关系 GROUP BY ... WHERE 子句类似。

可选`limit`参数限制返回的结果数。结果是有序的，因此只要数据不更改，相同的查询将返回相同的集合。

通过将`previous`参数传递给`Rows` GroupBy中的每个调用来支持通过结果进行分页。从上一个`GroupBy`查询中获取最后一个结果 ，并将该结果中的每个行ID作为`previous`参数传递`Rows`给下一个`GroupBy`查询中的每个相应查询。


**Result Type:** Array of "groups". Each group is an object with a group key and
a count key. The count is an integer, and the group is an array of objects which
specify the field and row for each row that was intersected to get that result.

**Examples:**

单个`Rows`查询。
```request
GroupBy(Rows(blah))
```
```response
[{"group":[{"field":"blah","rowID":1}],"count":1},
{"group":[{"field":"blah","rowID":9}],"count":1},
{"group":[{"field":"blah","rowID":39}],"count":1}]
```

有两个`Rows`查询，一个有ID，一个有键。
```request
GroupBy(Rows(blah), Rows(blahk), limit=7)
```
```response
[{"group":[{"field":"blah","rowID":1},{"field":"blahk","rowKey":"haha"}],"count":1},
 {"group":[{"field":"blah","rowID":1},{"field":"blahk","rowKey":"zaaa"}],"count":1},
 {"group":[{"field":"blah","rowID":1},{"field":"blahk","rowKey":"traa"}],"count":1},
 {"group":[{"field":"blah","rowID":9},{"field":"blahk","rowKey":"haha"}],"count":1},
 {"group":[{"field":"blah","rowID":9},{"field":"blahk","rowKey":"zaaa"}],"count":1},
 {"group":[{"field":"blah","rowID":9},{"field":"blahk","rowKey":"traa"}],"count":1},
 {"group":[{"field":"blah","rowID":39},{"field":"blahk","rowKey":"haha"}],"count":1}]
```

从前一个示例（分页）获取其余结果。
```request
GroupBy(Rows(blah, previous=39), Rows(blahk, previous="haha"), limit=7)
```

```response
[{"group":[{"field":"blah","rowID":39},{"field":"blahk","rowKey":"zaaa"}],"count":1},
 {"group":[{"field":"blah","rowID":39},{"field":"blahk","rowKey":"traa"}],"count":1}]
```
