+++
title = "入门指南"
weight = 3
nav = [
     "启动 pilosa",
     "例子",
     "下一步做什么?",
]
+++

## 启动 pilosa

Pilosa 默认提供 JSON 数据类型的 HTTP 服务。任何 HTTP 工具都可用于与 Pilosa 服务器进行交互。本文档中的示例将使用 [curl](https://curl.haxx.se/)，许多类 UNIX 系统，包括 Linux 和 MacOS 可以直接使用。Windows 用户可以在[这里](https://curl.haxx.se/download.html)下载 curl 。

<div class="note">
    <p>请注意，Pilosa服务器要求打开文件的上限。检查系统文档，了解如何在达到该限制时增加它。有关详细信息，请参阅<a href="administration.md#open-file-limits">打开文件限制</a>。</p>
</div>

### 开始 Pilosa

按照[安装文档](installation.md)中的步骤安装 Pilosa。在终端中执行以下命令以使用默认配置运行Pilosa（Pilosa 启动在 [localhost:10101](http://localhost:10101)）：

```
pilosa server
```
如果您使用的是 Docker 镜像，则可以使用以下命令在默认地址上运行短暂的 Pilosa 容器：
```
docker run -it --rm --name pilosa -p 10101:10101 pilosa/pilosa:latest
```

让我们确保Pilosa正在运行：
``` request
curl localhost:10101/status
```
``` response
{"state":"NORMAL","nodes":[{"id":"91715a50-7d50-4c54-9a03-873801da1cd1","uri":{"scheme":"http","host":"localhost","port
":10101},"isCoordinator":true}],"localID":"91715a50-7d50-4c54-9a03-873801da1cd1"}
```

### 例子

为了更好地理解 Pilosa 的功能，我们将创建一个名为“Star Trace”的示例项目，其中包含有关1,000名流行的Github 项目，这些项目名称中包含“go”。Star Trace 索引将包括数据点，例如编程语言，标签和点赞用户 - 已经为项目点赞的人。

尽管 Pilosa 没有以表格格式保存数据，但在描述数据模型时我们仍然使用术语“columns（列）”和“rows（行）”（译注：以下全部翻译为列和行）。我们将主对象放在列中，并将这些对象的属性放在行中。例如，Star Trace 项目将包含一个名为“repository”的索引，其中包含表示Github项目的列，以及表示编程语言和标记等属性的行。我们可以通过将行分组为名为Fields的集合来更好地组织行。因此，“项目”索引可能具有“语言”字段以及“标记”字段。您可以在文档的“ [数据模型](data-model.md)”部分中了解有关索引和字段的更多信息。

#### 创建 Schema

注意：
如果您想在任何时候验证数据结构，可以按如下方式查询 schema：



``` request
curl localhost:10101/schema
```
``` response
{"indexes":null}
```

在我们可以导入数据或运行查询之前，我们需要创建索引及其中的字段。让我们先创建项目索引：
``` request
curl localhost:10101/index/repository -X POST
```
``` response
{"success":true}
```
索引名称必须是不大于64个字符，以字母开头，并且只包含小写字母数字或`_-`符号。
让我们创建一个`stargazer`具有stargazers用户ID作为其行的字段：

``` request
curl localhost:10101/index/repository/field/stargazer \
     -X POST \
     -d '{"options": {"type": "time", "timeQuantum": "YMD"}}'
```
``` response
{"success":true}
```

由于我们的数据包含时间戳，这些时间戳代表用户点赞的时间，因此我们将字段类型设置为`time`。时间格式是我们想要使用的时间的分辨率，我们将其设置为`YMD`（年，月，日）`stargazer`。

接下来是`language`字段，用来存储编程语言的ID：
``` request
curl localhost:10101/index/repository/field/language \
     -X POST
```
``` response
{"success":true}
```
`language` 是一个 `set` 字段，但由于默认字段类型是 `set`，我们没有在选项中指定它。

#### 从CSV文件导入数据

<div class="note">
    <p>出于演示目的，我们使用`Pilosa`的内置实用程序来导入特殊格式的`CSV`文件。有关更多常规用法，请参阅各种客户端库如何在<a href="https://github.com/pilosa/go-pilosa/blob/master/docs/imports-exports.md">Go</a>，<a href="https://github.com/pilosa/java-pilosa/blob/master/docs/imports.md">Java</a>和<a href="https://github.com/pilosa/python-pilosa/tree/master/docs/imports.md">Python</a>中公开批量导入功能。</p>
</div>

在这里下载`stargazer.csv`和`language.csv`文件：

```
curl -O https://raw.githubusercontent.com/pilosa/getting-started/master/stargazer.csv
curl -O https://raw.githubusercontent.com/pilosa/getting-started/master/language.csv
```

运行以下命令将数据导入 Pilosa：

```
pilosa import -i repository -f stargazer stargazer.csv
pilosa import -i repository -f language language.csv
```

如果您正在为 Pilosa 使用Docker容器（名称 pilosa），则应将`*.csv`文件复制到容器中，然后导入它们：
```
docker cp stargazer.csv pilosa:/stargazer.csv
docker exec -it pilosa /pilosa import -i repository -f stargazer /stargazer.csv
docker cp language.csv pilosa:/language.csv
docker exec -it pilosa /pilosa import -i repository -f language /language.csv
```

请注意，用户ID和项目ID都重新映射到数据文件中的顺序整数，它们不再对应于实际的 Github ID。您可以查看[languages.txt](https://github.com/pilosa/getting-started/blob/master/languages.txt) 以查看语言的映射。

#### 查询

用户14赞的项目：
``` request
curl localhost:10101/index/repository/query \
     -X POST \
     -d 'Row(stargazer=14)'
```
``` response
{
    "results":[
        {
            "attrs":{},
            "columns":[1,2,3,362,368,391,396,409,416,430,436,450,454,460,461,464,466,469,470,483,484,486,490,491,503,504,514]
        }
    ]
}
```

项目使用的前5中语言是什么:
``` request
curl localhost:10101/index/repository/query \
     -X POST \
     -d 'TopN(language, n=5)'
```
``` response
{
    "results":[
        [
            {"id":5,"count":119},
            {"id":1,"count":50},
            {"id":4,"count":48},
            {"id":9,"count":31},
            {"id":13,"count":25}
        ]
    ]
}
```

用户14和19共同点赞了哪些项目：
``` request
curl localhost:10101/index/repository/query \
     -X POST \
     -d 'Intersect(
            Row(stargazer=14), 
            Row(stargazer=19)
        )'
```
``` response
{
    "results":[
        {
            "attrs":{},
            "columns":[2,3,362,396,416,461,464,466,470,486]
        }
    ]
}
```

用户14和19点赞了哪些项目：
``` request
curl localhost:10101/index/repository/query \
     -X POST \
     -d 'Union(
            Row(stargazer=14), 
            Row(stargazer=19)
        )'
```
``` response
{
    "results":[
        {
            "attrs":{},
            "columns":[1,2,3,361,362,368,376,377,378,382,386,388,391,396,398,400,409,411,412,416,426,428,430,435,436,450,452,453,454,456,460,461,464,465,466,469,470,483,484,486,487,489,490,491,500,503,504,505,512,514]
        }
    ]
}
```

哪些项目由用户14和19点赞，并且还用语言1编写：
``` request
curl localhost:10101/index/repository/query \
     -X POST \
     -d 'Intersect(
            Row(stargazer=14), 
            Row(stargazer=19),
            Row(language=1)
        )'
```
``` response
{
    "results":[
        {
            "attrs":{},
            "columns":[2,362,416,461]
        }
    ]
}
```
将用户99999为存储库77777的点赞：
``` request
curl localhost:10101/index/repository/query \
     -X POST \
     -d 'Set(77777, stargazer=99999)'
```
``` response
{"results":[true]}
```

请注意，虽然用户ID 99999可能与其他列ID不一致，但它仍然是一个相对较低的数字。
不要尝试在Pilosa中使用任意64位整数作为列或行ID - 这将导致诸如性能不佳和内存不足错误等问题。


### 下一步是什么？

下一步是什么？
您可以跳转到[数据模型](data-model.md)以深入了解 Pilosa 的数据模型，或[Query Language](query-language.md)以获取有关 **PQL** （Pilosa Query Language）的更多详细信息。查看[示例](examples.md)了解 Pilosa 真实的示例实现。准备好用您最喜欢的语言？看看我们的小型但不断扩展的[官方客户端库](client-libraries.md)。
