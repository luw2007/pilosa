+++
title = "安装"
weight = 2
nav = [
    "Installing on MacOS",
    "Installing on Linux",
]
+++


## 安装

Pilosa 目前支持 [MacOS](#installing-on-macos) 和 [Linux](#installing-on-linux)。

### Installing on MacOS

在MacOS上，Pilosa有四种安装方法：使用[Homebrew](https://brew.sh/)（推荐），下载二进制文件，从源码构建或使用 [Docker](#docker)。

#### 使用 Homebrew

1. 更新你的 Homebrew 仓库：
    ```
    brew update
    ```

2. 安装 Pilosa ：
    ```
    brew install pilosa
    ```

3. 检查是否成功安装 Pilosa：
    ```
    pilosa
    ```

    如果你看到类似的东西：
    ```
    Pilosa is a fast index to turbocharge your database.

    This binary contains Pilosa itself, as well as common
    tools for administering pilosa, importing/exporting data,
    backing up, and more. Complete documentation is available
    at https://www.pilosa.com/docs/.

    Version: v1.3.0
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

    安装成功。

#### 下载二进制文件

1. 下载最新版本：
    ```
    curl -L -O https://github.com/pilosa/pilosa/releases/download/v1.3.0/pilosa-v1.3.0-darwin-amd64.tar.gz
    ```
		其他版本可以从 Github 的[Releases](https://github.com/pilosa/pilosa/releases)页面下载。

2. 解压二进制文件：
    ```
    tar xfz pilosa-v1.3.0-darwin-amd64.tar.gz
    ```

3. 将二进制文件移动到你的执行目录中，以便可以 `pilosa` 在任何地方运行：
    ```
    cp -i pilosa-v1.3.0-darwin-amd64/pilosa /usr/local/bin
    ```

4. 检查是否成功安装 Pilosa：
    ```
    pilosa
    ```

    If you see something like:
    ```
    Pilosa is a fast index to turbocharge your database.

    This binary contains Pilosa itself, as well as common
    tools for administering pilosa, importing/exporting data,
    backing up, and more. Complete documentation is available
    at https://www.pilosa.com/docs/.

    Version: v1.3.0
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

    安装成功。

#### 从源码安装

<div class="note">
		<p>有关从源码构建的高级说明，请查看我们的<a href="https://github.com/pilosa/pilosa/blob/master/CONTRIBUTING.md">贡献者指南</a>。</p>
</div>

1. 安装依赖环境：

    * [Go](https://golang.org/doc/install). 请确保按照[GO官方](https://golang.org/doc/code.html#GOPATH) 设置了 `$GOPATH` 和 `$PATH` 环境变量。
    * [Git](https://git-scm.com/)

2. 克隆项目：
    ```
    mkdir -p ${GOPATH}/src/github.com/pilosa && cd $_
    git clone https://github.com/pilosa/pilosa.git
    ```

3. 编译 Pilosa 项目：
    ```
    cd $GOPATH/src/github.com/pilosa/pilosa
    make install-build-deps
    make install
    ```

4. 检查是否成功安装 Pilosa ：
    ```
    pilosa
    ```

    If you see something like:
    ```
    Pilosa is a fast index to turbocharge your database.

    This binary contains Pilosa itself, as well as common
    tools for administering pilosa, importing/exporting data,
    backing up, and more. Complete documentation is available
    at https://www.pilosa.com/docs/.

    Version: v1.3.0
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

    安装成功。

#### 下一步做什么?

跳转到 [入门指南](../getting-started/) 中创建你的第一个索引。

### 在 Linux 中安装

在 Linux 中安装 Pilosa 有三种方法：下载二进制文件（推荐），从源码构建或使用[Docker](#docker)。

#### 下载二进制

1. 要安装最新版本的Pilosa，请下载最新版本：
    ```
    curl -L -O https://github.com/pilosa/pilosa/releases/download/v1.3.0/pilosa-v1.3.0-linux-amd64.tar.gz
    ```
		注意：这假设您使用的是 `amd64` 兼容的体系结构。其他版本可以从 Github 的[Releases](https://github.com/pilosa/pilosa/releases)页面下载。

2. 解压二进制文件：
    ```
    tar xfz pilosa-v1.3.0-linux-amd64.tar.gz
    ```

3. 将二进制文件移动到你的执行目录中，以便可以 `pilosa` 在任何地方运行：
    ```
    cp -i pilosa-v1.3.0-linux-amd64/pilosa /usr/local/bin
    ```

4. 检查是否成功安装 Pilosa：
    ```
    pilosa
    ```

    If you see something like:
    ```
    Pilosa is a fast index to turbocharge your database.

    This binary contains Pilosa itself, as well as common
    tools for administering pilosa, importing/exporting data,
    backing up, and more. Complete documentation is available
    at https://www.pilosa.com/docs/.

    Version: v1.3.0
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

    安装成功。

#### 从源码安装

<div class="note">
    <p>有关从源码构建的高级说明，请查看我们的<a href="https://github.com/pilosa/pilosa/blob/master/CONTRIBUTING.md">贡献者指南</a>。</p>
</div>

1. 安装依赖环境：

    * [Go](https://golang.org/doc/install). 请确保按照[GO官方](https://golang.org/doc/code.html#GOPATH) 设置了 `$GOPATH` 和 `$PATH` 环境变量。
    * [Git](https://git-scm.com/)
    
2. 克隆项目：
    ```
    mkdir -p ${GOPATH}/src/github.com/pilosa && cd $_
    git clone https://github.com/pilosa/pilosa.git
    ```

3. 编译 Pilosa 项目：
    ```
    cd $GOPATH/src/github.com/pilosa/pilosa
    make install-build-deps
    make install
    ```

4. 检查是否成功安装 Pilosa ：
    ```
    pilosa
    ```

    If you see something like:
    ```
    Pilosa is a fast index to turbocharge your database.

    This binary contains Pilosa itself, as well as common
    tools for administering pilosa, importing/exporting data,
    backing up, and more. Complete documentation is available
    at https://www.pilosa.com/docs/.

    Version: v1.3.0
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

    安装成功。

#### 下一步做什么?

跳转到 [入门指南](../getting-started/) 中创建你的第一个索引。


### Windows

目前不支持 Windows 作为 Pilosa 的目标部署平台，但是 Docker 可以开发和运行 Pilosa。有关使用Docker for Windows 和 Docker Toolbox 的信息，请参阅[Docker](#docker)文档。

目前不支持 Windows 的 Linux 子系统。

### Docker

1. 为您的平台安装 Docker。在 Linux 上，Docker 可通过包管理器获得。在 MacOS 上，您可以使用 Docker for Mac 或 Docker Toolbox。在 Windows 上，您可以使用 Docker for Windows 或 Docker Toolbox。

2. **只有在使用 Docker Toolbox 时才需要执行此步骤**, 否则请跳至步骤3:
	
		a. `docker-machine start` 在终端中启动 Docker 支持。需要更新相应终端的环境变量，运行 `docker-machine env` 显示必要的命令。

    b. 在 VirtualBox GUI 或命令行中设置端口转发。客户端端口应为10101。对于主机端口，建议使用10101。如果 `VBoxManage` 命令在您的中 `PATH`，您可以使用以下命令（假设您使用默认VM）：

    ```
    VBoxManage modifyvm "default" --natpf1 "pilosa,tcp,,10101,,10101"
    ```

3. 确认Docker守护程序在后台运行：
    ```
    docker version
    ```
		如果您收到“command not found”或类似错误，请检查`docker`命令是否在您的路径中。如果未能正确执行命令，请启动 Docker 应用程序。


4. 从 Docker Hub 中提取官方 Pilosa 镜像：

    ```
    docker pull pilosa/pilosa:latest
    ```

5. 确保Pilosa已成功安装，并使其可访问:

    ```
    docker run -d --rm --name pilosa -p 10101:10101 pilosa/pilosa:latest server --bind 0.0.0.0:10101
    ```

6. 检查是否可以从容器外部访问它。

    在单独的终端中运行以下命令：
    ```
    curl localhost:10101/schema
    ```
		如果返回 `{"indexes":null}` 或类似，则可以从容器外部访问 Pilosa。否则请检查运行 Pilosa 容器时`-p 10101:10101`是否正确，VirtualBox中的端口映射是否正确（仅限 Docker Toolbox）。

7. 如果要终止 Pilosa 容器，可以运行以下命令：
    ```
    docker stop pilosa
    ```

#### 下一步做什么？
转到[入门指南](../getting-started/)，创建您的第一个Pilosa索引。


