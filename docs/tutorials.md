+++
title = "教程"
weight = 4
nav = [
    "设置安全集群",
    "设置 Docker 群集",
    "使用整数字段值",
    "存储行和列属性",
]
+++

## 教程

<div class="note">
<!-- this is html because there is a problem putting a list inside a shortcode -->
我们的一些教程作为独立的 repos 更好地工作，因为您可以同时`git clone`获得教程，代码和数据。这里列出了官方支持的教程。<br />
<br />
<ul>
<li><a href="https://github.com/pilosa/cosmosa">使用 Microsoft 的 Azure Cosmos DB 运行 Pilosa </a></li>
</ul>

</div>

### 设置安全群集

#### 介绍

Pilosa 支持使用TLS加密与集群中节点的所有通信。在本教程中，我们将设置在同一台计算机上运行的三节点 Pilosa 集群。相同的步骤可用于多计算机群集，但这需要设置防火墙和其他特定于平台的配置，这超出了本教程的范围。

本教程假定您使用的是类UNIX系统，例如 Linux 或 MacOS。也能在 Windows 10 系统上的[Windows子系统Linux（WSL）](https://msdn.microsoft.com/en-us/commandline/wsl/about) 运行成功。

#### 安装 Pilosa 并创建目录结构

如果您还没有开始，请在您的计算机上安装 Pilosa。对于 Linux 和 WSL（适用于Linux的Windows子系统），请参照[在Linux上安装](../installation/#installing-on-linux)说明。对于 MacOS，请参照[在MacOS上的安装](../installation/#installing-on-macos)。我们不支持其他平台的预编译版本，但您始终可以从源代码自行编译。请参阅[从源代码构建](../installation/#build-from-source)。

安装 Pilosa 后，您可能需要将其添加到您的`$PATH`。检查您是否可以从命令行运行 Pilosa：
``` request
pilosa --help
```
``` response
Pilosa is a fast index to turbocharge your database.

This binary contains Pilosa itself, as well as common
tools for administering pilosa, importing/exporting data,
backing up, and more. Complete documentation is available
at https://www.pilosa.com/docs/.

Version: v1.0.0
Build Time: 2018-05-14T22:14:01+0000

Usage:
  pilosa [command]

Available Commands:
  check           Do a consistency check on a pilosa data file.
  config          Print the current configuration.
  export          Export data from pilosa.
  generate-config Print the default configuration.
  help            Help about any command
  import          Bulk load data into pilosa.
  inspect         Get stats on a pilosa data file.
  server          Run Pilosa.

Flags:
  -c, --config string   Configuration file to read from.
  -h, --help            help for pilosa

Use "pilosa [command] --help" for more information about a command.
```

首先，创建一个目录，用于存放本教程的所有文件。然后切换到该目录：

```
mkdir $HOME/pilosa-tls-tutorial && cd $_
```

#### 创建 TLS 证书和 Gossip 密钥

保护 Pilosa 集群包括使用 TLS 和 Gossip 加密保护节点之间的通信。[Pilosa Enterprise](https://www.pilosa.com/enterprise/) 还支持身份验证和其他安全功能，但本教程不涉及这些功能。

第一步是获取 SSL 证书。您可以购买商业证书或使用免费加密证书[Let's Encrypt](https://letsencrypt.org/) ，但出于实际原因，我们将使用自签名证书。不建议在生产中使用自签名证书，因为它容易被发起中间人攻击。

以下命令将创建一个2048位自签名通配符证书，`*.pilosa.local`该证书将在10年后过期。

```
openssl req -x509 -newkey rsa:2048 -keyout pilosa.local.key -out pilosa.local.crt -days 3650 -nodes -subj "/C=US/ST=Texas/L=Austin/O=Pilosa/OU=Com/CN=*.pilosa.local"
```

上面的命令在当前目录中创建了两个文件：

* `pilosa.local.crt` 是SSL证书。
* `pilosa.local.key` 是私钥文件，必须保密。

创建 SSL 证书后，我们现在可以创建 Gossip 加密密钥。Gossip 加密密钥文件必须精确为16,24或32字节，以选择 AES-128，AES-192 或 AES-256 加密之一。从可靠的随机生成器`/dev/random`中获取非常符合我们的要求：
```
head -c 32 /dev/random > pilosa.local.gossip32
```

在当前目录中我们生成一个`pilosa.local.gossip32`文件，其中包含32个随机字节。


#### 创建配置文件

Pilosa 支持使用命令行选项，环境变量或配置文件传递配置项。在本教程中，我们将使用三个配置文件; 我们三个节点中的每一个都有一个配置文件。

必须选择群集中的一个节点作为*协调员*（coordinator）。我们在本教程中选择第一个节点作为协调员。协调员仅在群集大小调整操作期间很重要，否则就像群集中的任何其他节点一样。协调员可以通过分布式共识算法自动选择，并不推荐线上使用这个选项。

在项目目录中创建 `node1.config.toml`并粘贴以下内容：
```toml
# node1.config.toml

data-dir = "node1_data"
bind = "https://01.pilosa.local:10501"

[cluster]
coordinator = true

[tls]
certificate = "pilosa.local.crt"
key = "pilosa.local.key"
skip-verify = true

[gossip]
seeds = ["01.pilosa.local:15000"]
port = 15000
key = "pilosa.local.gossip32"
```

在项目目录中创建 `node2.config.toml`并粘贴以下内容：
```toml
# node2.config.toml

data-dir = "node2_data"
bind = "https://02.pilosa.local:10502"

[tls]
certificate = "pilosa.local.crt"
key = "pilosa.local.key"
skip-verify = true

[gossip]
seeds = ["01.pilosa.local:15000"]
port = 16000
key = "pilosa.local.gossip32"
```

在项目目录中创建 `node3.config.toml`并粘贴以下内容：
```toml
# node3.config.toml

data-dir = "node3_data"
bind = "https://03.pilosa.local:10503"

[tls]
certificate = "pilosa.local.crt"
key = "pilosa.local.key"
skip-verify = true

[gossip]
seeds = ["01.pilosa.local:15000"]
port = 17000
key = "pilosa.local.gossip32"
```

以下是配置项的一些说明：

* `data-dir`指向 Pilosa 服务写入数据的目录。如果它不存在，服务器将创建它。
* `bind`是服务器侦听传入请求的地址。地址由三部分组成：协议类型，主机和端口。默认方案是`http`，请明确指定`https`以便使用 HTTPS 协议进行节点之间的通信。
* `[集群]`包含群集的设置。我们只设置第一个节点`coordinator=true`来选择它作为协调节点。访问[群集配置](../configuration/#cluster-coordinator)查看其他设置。
* `[tls]` 包含 TLS 设置，包括 SSL 证书的路径和相应的密钥。设置`skip-verify=true`禁用主机名验证和其他安全措施。不要在生产服务器上设置`skip-verify=true`。
* `[gossip]`包含 Gossip 协议的设置。`seeds`包含从中为群集成员资格设定种子的节点列表。必须至少有一个 Gossip 种子。该`port`设置是节点的Gossip监听地址。如果群集的所有节点都在同一台计算机上运行，​​则每个节点的 Gossip 监听地址应该不同。否则，可以将其设置为相同的值。最后，`key`是我们之前创建的 Gossip 加密密钥。

#### 运行群集前的准备

运行群集之前，让我们确保`01.pilosa.local`，`02.pilosa.local`和`03.pilosa.local`解析为 IP 地址。如果您在计算机上运行群集，则将其添加到您的计算机`/etc/hosts`配置中。以下是执行此操作的众多方法之一（请注意“>>”）：
```
sudo sh -c 'printf "\n127.0.0.1 01.pilosa.local 02.pilosa.local 03.pilosa.local\n" >> /etc/hosts'
```

确保我们可以访问群集中的主机：
```
ping -c 1 01.pilosa.local
ping -c 1 02.pilosa.local
ping -c 1 03.pilosa.local
```

如果上述任何命令返回`ping: unknown host`，请确保`/etc/hosts`包含报错的主机名。

#### 运行群集

让我们打开三个终端窗口并在各自的窗口中运行每个节点。这将使我们能够更好地观察每个节点上发生的事情。

切换到第一个终端窗口，切换到项目目录并启动第一个节点：
```
cd $HOME/pilosa-tls-tutorial
pilosa server -c node1.config.toml
```

切换到第二个终端窗口，切换到项目目录并启动第二个节点：
```
cd $HOME/pilosa-tls-tutorial
pilosa server -c node2.config.toml
```

切换到第三个终端窗口，切换到项目目录并启动第三个节点：
```
cd $HOME/pilosa-tls-tutorial
pilosa server -c node3.config.toml
```

让我们确保所有三个 Pilosa 服务都在运行并且它们已连接：
``` request
curl -k --ipv4 https://01.pilosa.local:10501/status
```
``` response
{"state":"NORMAL","nodes":[{"id":"98ebd177-c082-4c54-8d48-7e7c75857b52","uri":{"scheme":"https","host":"02.pilosa.local","port":10502},"isCoordinator":false},{"id":"a33dc0d6-c35f-4559-984a-e582bf032a21","uri":{"scheme":"https","host":"03.pilosa.local","port":10503},"isCoordinator":false},{"id":"e24ac014-ee2f-4cb0-b565-74df6c551f0a","uri":{"scheme":"https","host":"01.pilosa.local","port":10501},"isCoordinator":true}]}
```
`-k`标志用于告诉 curl 忽略服务器提供的证书，并且`--ipv4`标志避免了 curl 在 MacOS 解析`127.0.0.1` 需要很长时间的问题。你可以在 Linux 和 WSL 上使用它们。

如果一切都设置正确，群集状态应该是`NORMAL`。

#### 开始查询

确认我们的集群正常运行后，让我们执行一些查询。首先，我们需要创建一个索引和一个字段：
``` request
curl https://01.pilosa.local:10501/index/sample-index \
     -k --ipv4 \
     -X POST
```
``` response
{"success":true}
```

这将`sample-index`使用默认选项创建索引。我们现在创建该字段：
``` request
curl https://01.pilosa.local:10501/index/sample-index/field/sample-field \
     -k --ipv4 \
     -X POST
```
``` response
{"success":true}
```

我们刚刚使用默认选项创建了字段`sample-field`。
让我们运行一个`Set`查询：
``` request
curl https://01.pilosa.local:10501/index/sample-index/query \
     -k --ipv4 \
     -X POST \
     -d 'Set(100, sample-field=1)'
```
``` response
{"results":[true]}
```

确认该值确实设置成功：
``` request
curl https://01.pilosa.local:10501/index/sample-index/query \
     -k --ipv4 \
     -X POST \
     -d 'Row(sample-field=1)'
```
``` response
{"results":[{"attrs":{},"columns":[100]}]}
```

查询集群中的其他节点时，应返回相同的结果：
``` request
curl https://02.pilosa.local:10501/index/sample-index/query \
     -k --ipv4 \
     -X POST \
     -d 'Row(sample-field=1)'
```
``` response
{"results":[{"attrs":{},"columns":[100]}]}
```

#### 下一步是什么？

查看我们的[管理指南](https://www.pilosa.com/docs/latest/administration/)了解 Pilosa 集群相关配置，使用[配置文档](https://www.pilosa.com/docs/latest/configuration/) 了解如何配置 Pilosa。

### 设置 Docker 群集

在本教程中，我们将使用 Docker 容器设置一个双节点 Pilosa 集群。

#### 在单个服务器上运行Docker群集

以下内容需要 Docker 1.13 或更高版本。

让我们首先确保 Pilosa 镜像是最新的：
```
docker pull pilosa/pilosa:latest
```

然后，创建一个虚拟网络来附加我们的容器。我们将命名我们的网络`pilosanet`:

```
docker network create pilosanet
```

让我们运行第一个 Pilosa 节点并将其连接到该虚拟网络。我们将第一个节点设置为集群协调员，并将其地址用作 gossip 种子。并将服务器地址设置为`pilosa1`：
```
docker run -it --rm --name pilosa1 -p 10101:10101 --network=pilosanet pilosa/pilosa:latest server --bind pilosa1:10101 --cluster.coordinator=true --gossip.seeds=pilosa1:14000
```

让我们运行第二个 Pilosa 节点并将其连接到虚拟网络。请注意，我们将 gossip 种子的地址设置为第一个节点的地址：
```
docker run -it --rm --name pilosa2 -p 10102:10101 --network=pilosanet pilosa/pilosa:latest server --bind pilosa2:10101 --gossip.seeds=pilosa1:14000
```

让我们测试集群中的节点是否连接正常：
``` request
curl localhost:10101/status
```
``` response
{"state":"NORMAL","nodes":[{"id":"2e8332d0-1fee-44dd-a359-e0d6ecbcefc1","uri":{"scheme":"http","host":"pilosa1","port":10101},"isCoordinator":true},{"id":"8c0dbcdc-9503-4265-8ad2-ba85a4bb10fa","uri":{"scheme":"http","host":"pilosa2","port":10101},"isCoordinator":false}],"localID":"2e8332d0-1fee-44dd-a359-e0d6ecbcefc1"}
```

对于第二个节点也类似：
``` request
curl localhost:10102/status
```
``` response
{"state":"NORMAL","nodes":[{"id":"2e8332d0-1fee-44dd-a359-e0d6ecbcefc1","uri":{"scheme":"http","host":"pilosa1","port":10101},"isCoordinator":true},{"id":"8c0dbcdc-9503-4265-8ad2-ba85a4bb10fa","uri":{"scheme":"http","host":"pilosa2","port":10101},"isCoordinator":false}],"localID":"2e8332d0-1fee-44dd-a359-e0d6ecbcefc1"}
```

相应的[Docker Compose](https://docs.docker.com/compose/)文件如下：
```yaml
version: '2'
services: 
  pilosa1:
    image: pilosa/pilosa:latest
    ports:
      - "10101:10101"
    environment:
      - PILOSA_CLUSTER_COORDINATOR=true
      - PILOSA_GOSSIP_SEEDS=pilosa1:14000
    networks:
      - pilosanet
    entrypoint:
      - /pilosa
      - server
      - --bind
      - "pilosa1:10101"
  pilosa2:
    image: pilosa/pilosa:latest
    ports:
      - "10102:10101"
    environment:
      - PILOSA_GOSSIP_SEEDS=pilosa1:14000
    networks:
      - pilosanet
    entrypoint:
      - /pilosa
      - server
      - --bind
      - "pilosa2:10101"
networks: 
  pilosanet:
```

#### 运行 Docker Swarm

使用`Docker Swarm`模式在不同的服务器上运行Pilosa Cluster非常容易。我们所要做的就是创建一个覆盖网络而不是桥接网络。

本节中的说明需要 Docker 17.06 或更高版本。虽然可以在 MacOS 或 Windows 上运行 Docker swarm，但最简单的方法是在 Linux 上运行它。以下假设您在 Linux 上运行。

我们将使用两个服务：管理节点在第一个服务中运行，一个工作节点在第二个服务中运行。

Docker 节点需要从外部访问某些端口。在继续之前，请确保在所有节点上打开以下端口：TCP/2377，TCP/7946，UDP/7946，UDP/4789。

让我们先初始化群。在管理节点上运行以下命令：
```
docker swarm init --advertise-addr=IP-ADDRESS
```

在云上运行的虚拟机通常至少有两个IP：外部IP和内部IP。使用外部接口的IP。

上面命令的输出应类似于：
```
To add a manager to this swarm, run the following command:

    docker swarm join --token SOME-TOKEN MANAGER-IP-ADDRESS:2377
```

让工作节点加入管理节点。将上面的命令复制/粘贴到 worker 的终端中，用正确的值替换 token 和IP地址。`--advertise-addr=WORKER-EXTERNAL-IP-ADDRESS`如果 worker 有多个IP，您可能需要添加参数：
```
docker swarm join --token SOME-TOKEN MANAGER-IP-ADDRESS:2377
```

在管理节点上运行以下命令以检查 worker 是否已加入 swarm：
```
docker node ls
```

哪个应该输出：

ID|主机名|状态|可用性|管理状态|内核版本
---|--------|------|------------|--------------|-------------
MANAGER-ID *|swarm1|Ready|Active|Leader|18.05.0-ce|
WORKER-ID|swarm2|Ready|Active||18.05.0-ce|

如果您`pilosanet`之前已创建过网络，请在继续之前将其删除，否则请跳至下一步：

```
docker network rm pilosanet
```

让我们创建`pilosanet`网络，但`overlay`这次是类型。我们还应该使这个网络可连接，以便能够将容器连接到它。在管理节点上运行以下命令：
```
docker network create -d overlay pilosanet --attachable
```

我们现在可以创建 Pilosa 容器了。让我们先启动协调员节点。在其中一台服务器上运行以下命令：
```
docker run -it --rm --name pilosa1 --network=pilosanet pilosa/pilosa:latest server --bind pilosa1:10101 --cluster.coordinator=true --gossip.seeds=pilosa1:14000
```

以下是另一台服务器：
```
docker run -it --rm --name pilosa2 --network=pilosanet pilosa/pilosa:latest server --bind pilosa2:10101 --gossip.seeds=pilosa1:14000
```

这些是我们在上一节中使用的相同命令，但端口映射除外！让我们在同一个虚拟网络上运行另一个容器来从协调员中读取状态：
``` request
docker run -it --rm --network=pilosanet --name shell alpine wget -q -O- pilosa1:10101/status
```
``` response
{"state":"NORMAL","nodes":[{"id":"3e3b0abd-1945-441a-a01f-5a28272972f5","uri":{"scheme":"http","host":"pilosa1","port":10101},"isCoordinator":true},{"id":"71ed27cc-9443-4f41-88fb-1c22f92bf695","uri":{"scheme":"http","host":"pilosa2","port":10101},"isCoordinator":false}],"localID":"3e3b0abd-1945-441a-a01f-5a28272972f5"}
```

您可以使用上述步骤向群集和 Pilosa 群集添加其他工作节点。

#### 下一步是什么？

查看我们的[管理指南](https://www.pilosa.com/docs/latest/administration/)了解 Pilosa 集群相关配置，使用[配置文档](https://www.pilosa.com/docs/latest/configuration/) 了解如何配置 Pilosa。

请参考[Docker 文档](https://docs.docker.com)以查看有关 Docker 容器运行的选项。在与[覆盖网络组网](https://docs.docker.com/network/network-tutorial-overlay/)是 Docker swarm mode 和 overlay networks 的详细概述。

### 使用整数字段值

#### 介绍

Pilosa 可以存储相关联的索引中的列的整数值，并且这些值将被用来支持`Row`，`Min`，`Max`，和`Sum`查询。在本教程中，我们将展示如何设置整数字段，使用数据填充这些字段以及查询字段。我们要创建的示例索引将代表医疗机构中的虚构患者以及有关这些患者的各种信息。

首先，创建一个叫`patients`的索引：
``` request
curl localhost:10101/index/patients \
     -X POST 
```
``` response
{"success":true}
```

除了存储位行之外，字段还可以存储整数值。接下来的步骤创建三个字段(`age`, `weight`, `tcells`)的`patients`索引。
``` request
curl localhost:10101/index/patients/field/age \
     -X POST \
     -d '{"options":{"type": "int", "min": 0, "max": 120}}'
```
``` response
{"success":true}
```

``` request
curl localhost:10101/index/patients/field/weight \
     -X POST \
     -d '{"options":{"type": "int", "min": 0, "max": 500}}'
```
``` response
{"success":true}
```

``` request
curl localhost:10101/index/patients/field/tcells \
     -X POST \
     -d '{"options":{"type": "int", "min": 0, "max": 2000}}'
```
``` response
{"success":true}
```

接下来，让我们用数据填充我们的字段。有两种方法可以将数据输入到字段中：使用`Set()` PQL 函数单独设置字段，或使用`pilosa import`命令一次导入多个值。首先，让我们使用 PQL 设置一些字段数据。

以下查询我们系统中为患者1设置年龄，体重和T细胞计数：
``` request
curl localhost:10101/index/patients/query \
     -X POST \
     -d 'Set(1, age=34)'
```
``` response
{"results":[true]}
```

``` request
curl localhost:10101/index/patients/query \
     -X POST \
     -d 'Set(1, weight=128)'
```
``` response
{"results":[true]}
```

``` request
curl localhost:10101/index/patients/query \
     -X POST \
     -d 'Set(1, tcells=1145)'
```
``` response
{"results":[true]}
```

在我们需要一次加载大量数据的情况下，我们可以使用该`pilosa import`命令。此方法允许我们从 CSV 文件将数据导入 Pilosa。

假设我们有一个名为`ages.csv`的结构文件：
```
1,34
2,57
3,19
4,40
5,32
6,71
7,28
8,33
9,63
```
如果CSV的第一列代表患者ID，第二列代表患者`age`，那么我们可以`age`通过运行以下命令将数据导入到我们的字段中：


```
pilosa import -i patients --field age ages.csv
```

现在我们的索引中有一些数据，让我们运行一些查询来演示如何使用该数据。

为了找到40岁以上的所有患者，然后只需对该`age`字段进行 Row 查询。
``` request
curl localhost:10101/index/patients/query \
     -X POST \
     -d 'Row(age > 40)'
```
``` response
{"results":[{"attrs":{},"columns":[2,6,9]}]}
```

您可以在[Row（BSI）查询](../query-language/#row-bsi)文档中找到支持的范围运算符列表。

要查找所有患者的平均年龄，请运行`Sum`查询：
``` request
curl localhost:10101/index/patients/query \
     -X POST \
     -d 'Sum(field="age")'
```
``` response
{"results":[{"value":377,"count":9}]}
```
`Sum`查询获得的结果包含所有年龄的总和以及`count`列的总和。您可以通过`value`除以`count`得到平均值。

您还可以为该`Sum()`功能提供过滤器，以查找所有40岁以上患者的平均年龄。
``` request
curl localhost:10101/index/patients/query \
     -X POST \
     -d 'Sum(Row(age > 40), field="age")'
```
``` response
{"results":[{"value":191,"count":3}]}
```
请注意，在这种情况下，`count`只有`3`是执行查询过滤器`age > 40`导致。

要查找所有患者的最低年龄，请运行`Min`查询：
``` request
curl localhost:10101/index/patients/query \
     -X POST \
     -d 'Min(field="age")'
```
``` response
{"results":[{"value":19,"count":1}]}
```
从`Min`查询中获得的结果包含`value`所有值的最小值以及`count`具有该值的列。

您还可以为该`Min()`功能提供过滤器，以查找所有40岁以上患者的最低年龄。
``` request
curl localhost:10101/index/patients/query \
     -X POST \
     -d 'Min(Row(age > 40), field="age")'
```
``` response
{"results":[{"value":57,"count":1}]}
```

要查找所有患者的最大年龄，请运行`Max`查询：
``` request
curl localhost:10101/index/patients/query \
     -X POST \
     -d 'Max(field="age")'
```
``` response
{"results":[{"value":71,"count":1}]}
```
从Max查询中获得的结果包含`value`所有值的最大值以及`count`具有该值的列。

您还可以为该`Max()`功能提供过滤器，以查找所有40岁以下患者的最大年龄。
``` request
curl localhost:10101/index/patients/query \
     -X POST \
     -d 'Max(Row(age < 40), field="age")'
```
``` response
{"results":[{"value":34,"count":1}]}
```

### 存储行和列属性

#### 介绍

Pilosa 可以存储与任何行或列关联的任意值。在 Pilosa，这些被称为`attributes`，并且它们可以是类型的`string`，`integer`，`boolean`，`float`。在本教程中，我们将存储一些属性数据，然后运行一些返回该数据的查询。

首先，为本教程创建一个名为`books`索引：
``` request
curl localhost:10101/index/books \
     -X POST
```
``` response
{"success":true}
```

接下来，在`books`索引中创建一个字段`members`，该字段将表示已阅读书籍的用户。
``` request
curl localhost:10101/index/books/field/members \
     -X POST \
     -d '{}'
```
``` response
{"success":true}
```

现在，让我们往索引中添加一些书籍。
``` request
curl localhost:10101/index/books/query \
     -X POST \
     -d 'SetColumnAttrs(1, name="To Kill a Mockingbird", year=1960)
         SetColumnAttrs(2, name="No Name in the Street", year=1972)
         SetColumnAttrs(3, name="The Tipping Point", year=2000)
         SetColumnAttrs(4, name="Out Stealing Horses", year=2003)
         SetColumnAttrs(5, name="The Forever War", year=2008)'
```
``` response
{"results":[null,null,null,null,null]}
```

并添加一些用户。
``` request
curl localhost:10101/index/books/query \
     -X POST \
     -d 'SetRowAttrs(members, 10001, fullName="John Smith")
         SetRowAttrs(members, 10002, fullName="Sue Perkins")
         SetRowAttrs(members, 10003, fullName="Jennifer Hawks")
         SetRowAttrs(members, 10004, fullName="Pedro Vazquez")
         SetRowAttrs(members, 10005, fullName="Pat Washington")'
```
``` response
{"results":[null,null,null,null,null]}
```

此时，我们可以通过查询`members`来查询其中一条记录。
``` request
curl localhost:10101/index/books/query \
     -X POST \
     -d 'Row(members=10002)'
```
``` response
{"results":[{"attrs":{"fullName":"Sue Perkins"},"columns":[]}]}
```

现在让我们在矩阵中添加一些数据，这样每对代表一个读过那本书的成员。
``` request
curl localhost:10101/index/books/query \
     -X POST \
     -d 'Set(3, members=10001)
         Set(5, members=10001)
         Set(1, members=10002)
         Set(2, members=10002)
         Set(4, members=10002)
         Set(3, members=10003)
         Set(4, members=10004)
         Set(5, members=10004)
         Set(1, members=10005)
         Set(2, members=10005)
         Set(3, members=10005)
         Set(4, members=10005)
         Set(5, members=10005)'
```
``` response
{"results":[true,true,true,true,true,true,true,true,true,true,true,true,true]}
```

现在`Sue Perkins`再次拉开记录。
``` request
curl localhost:10101/index/books/query \
     -X POST \
     -d 'Row(members=10002)'
```
``` response
{"results":[{"attrs":{"fullName":"Sue Perkins"},"columns":[1,2,4]}]}
```
请注意，结果集现在包含`columns`属性中的整数列表。这些整数与Sue读过的书籍的列ID相匹配。

为了检索我们为每本书存储的属性信息，我们需要`columnAttrs=true`在查询中添加一个URL参数。
``` request
curl localhost:10101/index/books/query?columnAttrs=true \
     -X POST \
     -d 'Row(members=10002)'
```
``` response
{
  "results":[{"attrs":{"fullName":"Sue Perkins"},"columns":[1,2,4]}],
  "columnAttrs":[
    {"id":1,"attrs":{"name":"To Kill a Mockingbird","year":1960}},
    {"id":2,"attrs":{"name":"No Name in the Street","year":1972}},
    {"id":4,"attrs":{"name":"Out Stealing Horses","year":2003}}
  ]
}
```
该`book`属性包含在结果集中的`columnAttrs`属性。
最后，如果我们要找出双方都读哪些书`Sue`和`Pedro`，我们只是执行两个成员交集查询：
``` request
curl localhost:10101/index/books/query?columnAttrs=true \
     -X POST \
     -d 'Intersect(Row(members=10002), Row(members=10004))'
```
``` response
{
  "results":[{"attrs":{},"columns":[4]}],
  "columnAttrs":[
    {"id":4,"attrs":{"name":"Out Stealing Horses","year":2003}}
  ]
}
```
请注意，我们没有获得复杂查询的行属性，但我们仍然获得列属性，在本案例中是书信息。

