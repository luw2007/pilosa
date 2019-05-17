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

本节将提供Pilosa查询语言（PQL）的详细参考和示例. 所有PQL查询都在单个[索引](../glossary/#index)上运行，并通过 `/index/INDEX_NAME/query` 接口传输给 Pilosa. 
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

在运行下面的任何示例查询之前，请按照[入门指南](../getting-started/)部分中的说明设置索引和字段，并使用一些数据填充它们。

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

* `field` 该字段指定查询将在哪个 Pilosa [字段](../glossary/#field) 运行。有效字段名称是小写字符串; 它们以字母数字字符开头，仅包含字母数字字符和`_-`。它们的长度不得超过64个字符。
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

这在 stargazer 字段中设置了一个位，表示 id=1 的用户为项目 id=10 点赞。

Set还支持提供时间戳。要编写用户为项目点赞的日期：

This sets a bit in the stargazer field, representing that the user with id=1 has starred the repository with id=10.

Set also supports providing a timestamp. To write the date that a user starred a repository:
```request
Set(10, stargazer=1, 2016-01-01T00:00)
```
```response
{"results":[true]}
```

Set multiple bits in a single request:
```request
Set(10, stargazer=1) Set(20, stargazer=1) Set(10, stargazer=2) Set(30, stargazer=2)
```
```response
{"results":[false,true,true,true]}
```

Set the field "pullrequests" to integer value 2 at column 10:
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

`SetRowAttrs` associates arbitrary key/value pairs with a row in a field. Setting a value of `null`, without quotes, deletes an attribute.

**Result Type:** null

SetRowAttrs queries always return `null` upon success.

**Examples:**

Set attributes `username` and `active` on row 10:
```request
SetRowAttrs(stargazer, 10, username="mrpi", active=true)
```
```response
{"results":[null]}
```

Set username value and active status for user 10. These are arbitrary key/value pairs which have no meaning to Pilosa. You can see the attributes you've set on a row with a [Row](../query-language/#row) query like so `Row(stargazer=10)`.

Delete attribute `username` on row 10:
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

`SetColumnAttrs` associates arbitrary key/value pairs with a column in an index.

**Result Type:** null

SetColumnAttrs queries always return `null` upon success. Setting a value of `null`, without quotes, deletes an attribute.

**Examples:**

Set attributes `stars`, `url`, and `active` on column 10:
```request
SetColumnAttrs(10, stars=123, url="http://projects.pilosa.com/10", active=true)
```
```response
{"results":[null]}
```

Set url value and active status for project 10. These are arbitrary key/value pairs which have no meaning to Pilosa.

ColumnAttrs can be requested by adding the URL parameter `columnAttrs=true` to a query. For example:
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

In this example, ColumnAttrs have been set on columns 10 and 20, but not column 30. The relevant attributes are all returned in a single columnAttrs list. See the [query index](../api-reference/#query-index) section for more information.

Delete the `url` attribute on column 10:
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

`Clear` assigns a value of 0 to a bit in the binary matrix, thus disassociating the given row in the given field from the given column.

Note that clearing a column on a time field will remove all data for that column.

**Result Type:** boolean

A return value of `true` indicates that the bit was toggled from 1 to 0.

A return value of `false` indicates that the bit was already set to 0 and nothing changed.

**Examples:**

Clear the bit at row 1 and column 10 in the stargazer field:
```request
Clear(10, stargazer=1)
```
```response
{"results":[true]}
```

This represents removing the relationship between the user with id=1 and the repository with id=10.

#### ClearRow

**Spec:**

```
ClearRow(<FIELD>=<ROW>)
```

**Description:**

`ClearRow` sets all bits to 0 in a given row of the binary matrix, thus disassociating the given row in the given field from all columns.

**Result Type:** boolean

A return value of `true` indicates that at least one column was toggled from 1 to 0.

A return value of `false` indicates that all bits in the row were already 0 and nothing changed.

**Examples:**

Clear all bit in row 1 in the stargazer field:
```request
ClearRow(stargazer=1)
```
```response
{"results":[true]}
```

This represents removing the relationship between the user with id=1 and all repositories.

#### Store

**Spec:**

```
Store(<ROW_CALL>, <FIELD>=<ROW>)
```

**Description:**

`Store` writes the results of `<ROW_CALL>` to the specified row. If the row already exists, it will be replaced. The destination field must be of field type `set`.

**Result Type:** boolean

Upon success, this method always returns `true`. A future version of Pilosa may use this boolean result to indicate whether or not the data in the destination row was changed by the `Store` call.

**Examples:**

Store the contents of stargazer row 1 into stargazer row 2:
```request
Store(Row(stargazer=1), stargazer=2)
```
```response
{"results":[true]}
```

Store the results of the intersection of stargazer rows 10 and 11 into stargazer row 20.
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

`Row` retrieves the indices of all the columns in a row. It also retrieves any attributes set on that row.

**Result Type:** object with attrs and columns.

例如：`{"attrs":{"username":"mrpi","active":true},"columns":[10, 20]}`

**Examples:**

Query all columns with a bit set in row 1 of the field `stargazer` (repositories that are starred by user 1):
```request
Row(stargazer=1)
```
```response
{"attrs":{"username":"mrpi","active":true},"columns":[10, 20]}
```

* attrs are the attributes for user 1
* columns are the repositories which user 1 has starred.


#### Row (Range)

**Spec:**

```
Row(<FIELD>=<ROW>, from=<TIMESTAMP>, to=<TIMESTAMP>)
```

**Description:**

Similar to `Row`, but only returns bits which were set with timestamps between the given `from` (inclusive) and `to` (exclusive) timestamps. Both `from` and `to` parameters are optional. The default for `to` timestamp is current time + 1 day. If a later end timestamp is required, specify it explicitly.

**Result Type:** object with attrs and bits


**Examples:**

Query all columns with a bit set in row 1 of a field (repositories that a user has starred), within a date range:
```request
Row(stargazer=1, from='2010-01-01T00:00', to='2017-03-02T03:00')
```
```response
{{"attrs":{},"columns":[10]}
```

This example assumes timestamps have been set on some bits.

* columns are repositories which were starred by user 1 in the time range 2010-01-01 to 2017-03-02.


#### Row (BSI)

**Spec:**

```
Row([<COMPARISON_VALUE> <COMPARISON_OPERATOR>] <FIELD> <COMPARISON_OPERATOR> <COMPARISON_VALUE>)
```

**Description:**

The `Row` query is overloaded to work on `integer` values as well as `timestamp` values.
Returns bits that are true for the comparison operator.

**Result Type:** object with attrs and columns

**Examples:**

In our source data, commitactivity was counted over the last year.
The following greater-than `Row` query returns all columns with a field value greater than 100 (repositories having more than 100 commits):

```request
Row(commitactivity > 100)
```
```response
{{"attrs":{},"columns":[10]}
```

* columns are repositories which had at least 100 commits in the last year.

BSI range queries support the following operators:

 Operator | Name                          | Value
----------|-------------------------------|--------------------
 `>`      | greater-than, GT              | integer
 `<`      | less-than, LT                 | integer
 `<=`     | less-than-or-equal-to, LTE    | integer
 `>=`     | greater-than-or-equal-to, GTE | integer
 `==`     | equal-to, EQ                  | integer
 `!=`     | not-equal-to, NEQ             | integer or `null`

A bounded interval can be specified by chaining the `<` and `<=` operators (but not others). For example:

```request
Row(50 < commitactivity < 150)
```
```response
{{"attrs":{},"columns":[10]}
```

As of Pilosa 1.0, the "between" syntax `Row(frame=stats, commitactivity >< [50, 150])` is no longer supported.

#### Union

**Spec:**

```
Union([ROW_CALL ...])
```

**Description:**

Union performs a logical OR on the results of all `ROW_CALL` queries passed to it.

**Result Type:** object with attrs and bits

attrs will always be empty

**Examples:**

Query columns with a bit set in either of two rows (repositories that are starred by either of two users):
```request
Union(Row(stargazer=1), Row(stargazer=2))
```
```response
{"attrs":{},"columns":[10, 20, 30]}
```

* columns are repositories that were starred by user 1 OR user 2

#### Intersect

**Spec:**

```
Intersect(<ROW_CALL>, [ROW_CALL ...])
```

**Description:**

Intersect performs a logical AND on the results of all `ROW_CALL` queries passed to it.

**Result Type:** object with attrs and columns

attrs will always be empty

**Examples:**

Query columns with a bit set in both of two rows (repositories that are starred by both of two users):

```request
Intersect(Row(stargazer=1), Row(stargazer=2))
```
```response
{"attrs":{},"columns":[10]}
```

* columns are repositories that were starred by user 1 AND user 2

#### Difference

**Spec:**

```
Difference(<ROW_CALL>, [ROW_CALL ...])
```

**Description:**

Difference returns all of the bits from the first `ROW_CALL` argument passed to it, without the bits from each subsequent `ROW_CALL`.

**Result Type:** object with attrs and columns

attrs will always be empty

**Examples:**

Query columns with a bit set in one row and not another (repositories that are starred by one user and not another):
```request
Difference(Row(stargazer=1), Row(stargazer=2))
```
```response
{"results":[{"attrs":{},"columns":[20]}]}
```

* columns are repositories that were starred by user 1 BUT NOT user 2

Query for the opposite difference:
```request
Difference(Row(stargazer=2), Row(stargazer=1))
```
```response
{"attrs":{},"columns":[30]}
```

* columns are repositories that were starred by user 2 BUT NOT user 1

#### Xor

**Spec:**

```
Xor(<ROW_CALL>, [ROW_CALL ...])
```

**Description:**

Xor performs a logical XOR on the results of each `ROW_CALL` query passed to it.

**Result Type:** object with attrs and columns

attrs will always be empty

**Examples:**

Query columns with a bit set in exactly one of two rows (repositories that are starred by only one of two users):

```request
Xor(Row(stargazer=2), Row(stargazer=1))
```
```response
{"results":[{"attrs":{},"columns":[20,30]}]}
```

* columns are repositories that were starred by user 1 XOR user 2 (user 1 or user 2, but not both)

#### Not

**Spec:**

```
Not(<ROW_CALL>)
```

**Description:**

Not returns the inverse of all of the bits from the `ROW_CALL` argument. The Not query requires that `trackExistence` has been enabled on the Index.

**Result Type:** object with attrs and columns

attrs will always be empty

**Examples:**

Query existing columns that do not have a bit set in the given row.
```request
Not(Row(stargazer=1))
```
```response
{"results":[{"attrs":{},"columns":[30]}]}
```

* columns are repositories that were not starred by user 1

#### Count
**Spec:**

```
Count(<ROW_CALL>)
```

**Description:**

Returns the number of set bits in the `ROW_CALL` passed in.

**Result Type:** int

**Examples:**

Query the number of bits set in a row (the number of repositories a user has starred):
```request
Count(Row(stargazer=1))
```
```response
{"results":[1]}
```

* Result is the number of repositories that user 1 has starred.

#### Shift
**Spec:**

```
Shift(<ROW_CALL>, [n=UINT])
```

**Description:**

Returns the row specified by `ROW_CALL` shifted by `n` bits.

**Result Type:** object with attrs and columns

attrs will always be empty

**Examples:**

Query all columns with a bit set in row 1 of the field `stargazer`
and shift the result by 2:
```request
Shift(Row(stargazer=1), n=2)
```
```response
{"attrs":{},"columns":[12, 22]}
```

* columns are the repositories which user 1 has starred shifted by 2 bits.

#### TopN

**Spec:**

```
TopN(<FIELD>, [ROW_CALL], [n=UINT],
     [attrName=<ATTR_NAME>, attrValues=<[]ATTR_VALUE>])
```

**Description:**

Return the id and count of the top `n` rows (by count of bits) in the field.
The `attrName` and `attrValues` arguments work together to only return rows which
have the attribute specified by `attrName` with one of the values specified in
`attrValues`.

**Result Type:** array of key/count objects

**Caveats:**

* Performing a TopN() query on a field with cache type ranked will return the top rows sorted by count in descending order.
* Fields with cache type lru will maintain an LRU (Least Recently Used replacement policy) cache, thus a TopN query on this type of field will return rows sorted in order of most recently set bit.
* The field's cache size determines the number of sorted rows to maintain in the cache for purposes of TopN queries. There is a tradeoff between performance and accuracy; increasing the cache size will improve accuracy of results at the cost of performance.
* Once full, the cache will truncate the set of rows according to the field option CacheSize. Rows that straddle the limit and have the same count will be truncated in no particular order.
* The TopN query's attribute filter is applied to the existing sorted cache of rows. Rows that fall outside of the sorted cache range, even if they would normally pass the filter, are ignored.

See [field creation](../api-reference/#create-field) for more information about the cache.

**Examples:**

Basic TopN query:
```request
TopN(stargazer)
```
```response
{"results":[[{"id":1240,"count":102},{"id":4734,"count":100},{"id":12709,"count":93},...]]}
```

* `id` is a row ID (user ID)
* `count` is a count of columns (repositories)
* Results are the number of bits set in the corresponding row (repositories that each user starred) in descending order for all rows (users) in the stargazer field. For example user 1240 starred 102 repositories, user 4734 starred 100 repositories, user 12709 starred 93 repository.

Limit the number of results:
```request
TopN(stargazer, n=2)
```
```response
{"results":[[{"id":1240,"count":102},{"id":4734,"count":100}]]}
```

* Results are the top two rows (users) sorted by number of bits set (repositories they've starred) in descending order.

Filter based on an existing row:
```request
TopN(stargazer, Row(language=1), n=2)
```
```response
{"results":[[{"id":1240,"count":35},{"id":7508,"count":32}]]}
```

* Results are the top two users (rows) sorted by the number of bits set in the intersection with row 1 of the language field (repositories that they've starred which are written in language 1).

Filter based on attributes:
```request
TopN(stargazer, n=2, attrName=active, attrValues=[true])
```
```response
{"results":[[{"id":10,"count":1},{"id":13,"count":1}]]}
```

* Results are the top two users (rows) which have the "active" attribute set to "true", sorted by the number of bits set (repositories that they've starred).


#### Min

**Spec:**

```
Min([ROW_CALL], field=<FIELD>)
```

**Description:**

Returns the minimum value of all BSI integer values in this `field`. If the optional `Row` call is supplied, only columns with set bits are considered, otherwise all columns are considered.

**Result Type:** object with the min and count of columns containing the min value.

**Examples:**

Query the minimum value of a field (minimum size of all repositories):
```request
Min(field="diskusage")
```
```response
{"value":4,"count":2}
```

* Result is the smallest value (repository size in kilobytes, here), plus the count of columns with that value.

#### Max

**Spec:**

```
Max([ROW_CALL], field=<FIELD>)
```

**Description:**

Returns the maximum value of all BSI integer values in this `field`. If the optional `Row` call is supplied, only columns with set bits are considered, otherwise all columns are considered.

**Result Type:** object with the max and count of columns containing the max value.

**Examples:**

Query the maximum value of a field (maximum size of all repositories):
```request
Max(field="diskusage")
```
```response
{"value":88,"count":13}
```

* Result is the largest value (repository size in kilobytes, here), plus the count of columns with that value.

#### Sum

**Spec:**

```
Sum([ROW_CALL], field=<FIELD>)
```

**Description:**

Returns the count and computed sum of all BSI integer values in the `field`. If the optional `Row` call is supplied, columns with set bits are summed, otherwise the sum is across all columns.

**Result Type:** object with the computed sum and count of the values in the integer field.

**Examples:**

Query the size of all repositories.
```request
Sum(field="diskusage")
```
```response
{"value":10,"count":3}
```

* Result is the sum of all values (total size of all repositories in kilobytes, here), plus the count of columns.

### Other Operations

#### Options

**Spec:**

```
Options(<CALL>, columnAttrs=<BOOL>, excludeColumns=<BOOL>, excludeRowAttrs=<BOOL>, shards=[UINT ...])
```

**Description:**

Modifies the given query as follows:

* `columnAttrs`: Include column attributes in the result (Default: `false`).
* `excludeColumns`: Exclude column IDs from the result (Default: `false`).
* `excludeRowAttrs`: Exclude row attributes from the result (Default: `false`).
* `shards`: Run the query using only the data from the given shards. By default, the entire data set (i.e. data from all shards) is used.

**Result Type:** Same result type as `<CALL>`.

**Examples:**

Return column attributes:
```request
Options(Row(f1=10), columnAttrs=true)
```
```response
{"attrs":{},"columns":[100]}],"columnAttrs":[{"id":100,"attrs":{"foo":"bar"}}
```

Run the query against shards 0 and 2 only:
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

Rows returns a list of row IDs in the given field which have at least one bit
set. The field argument is mandatory, the others are  optional.

If `previous` is given, rows prior to and including the specified row ID or
key will not be returned. If `column` is given, only rows which have a set bit
in the given column will be returned. `previous` or `column` must be strings if
and only if the field or index respectively is using key translation. If `limit`
is given, the number of rowIDs returned will be less than or equal to
`limit`. The combination of `limit` and `previous` allows for paging over large
result sets. Results are always ordered, so setting `previous` as the last
result of the previous request will start from the next available row.

If the field is of type `time`, the `from` and `to` arguments can be provided
to restrict the result to a specific time span. If `from` and `to` are
not provided, the full range of existing data will be queried.

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

GroupBy returns the count of the intersection of every combination of rows
taking one row each from the specified `Rows` calls. It returns only those
combinations for which the count is greater than 0.

The optional `filter` argument takes any type of `Row` query (例如：Row, Union,
 Intersect, etc.) which will be intersected with each result prior to returning
 the count. This is analagous to a WHERE clause applied to a relational GROUP BY
 query.

The optional `limit` argument limits the number of results returned. The results
are ordered, so as long as the data isn't changing, the same query will return
the same result set.

Paging through results is supported by passing the `previous` argument to each
of the `Rows` calls in the GroupBy. Take the last result from your previous
`GroupBy` query, and pass each row ID in that result as the `previous` argument
to each of the respective `Rows` queries in your next `GroupBy` query.

**Result Type:** Array of "groups". Each group is an object with a group key and
a count key. The count is an integer, and the group is an array of objects which
specify the field and row for each row that was intersected to get that result.

**Examples:**

A single `Rows` query.
```request
GroupBy(Rows(blah))
```
```response
[{"group":[{"field":"blah","rowID":1}],"count":1},
{"group":[{"field":"blah","rowID":9}],"count":1},
{"group":[{"field":"blah","rowID":39}],"count":1}]
```

With two `Rows` queries - one with IDs and one with keys.
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

Getting the rest of the results from the previous example (paging).
```request
GroupBy(Rows(blah, previous=39), Rows(blahk, previous="haha"), limit=7)
```

```response
[{"group":[{"field":"blah","rowID":39},{"field":"blahk","rowKey":"zaaa"}],"count":1},
 {"group":[{"field":"blah","rowID":39},{"field":"blahk","rowKey":"traa"}],"count":1}]
```
