# envCheck工具使用指导

目前工具已实现以下功能。

- 给定文件路径列表，查询当前环境中是否残留文件，并输出检查报告。
- 给定文件路径列表，支持安全删除（询问用户是否确认删除）和强制删除（直接删除不询问）文件，并输出检查报告。
- 检查环境中是否安装指定程序，并输出检查报告。
- 支持多机器之间的时间同步校验，检测其余节点和引导节点之间时钟是否同步。

## 前提条件

本工具目前只支持在`Linux amd64`和`Linux arm64`环境下执行，请根据环境自行下载对应版本的工具。

| 下载项       | amd64                                                        | arm64                                                        |
| ------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| envCheck工具 | [下载](https://openfuyao.obs.cn-north-4.myhuaweicloud.com/openFuyao/env-check/releases/download/latest/bin/linux/amd64/envCheck) | [下载](https://openfuyao.obs.cn-north-4.myhuaweicloud.com/openFuyao/env-check/releases/download/latest/bin/linux/arm64/envCheck) |
| 默认配置文件 | [下载](https://openfuyao.obs.cn-north-4.myhuaweicloud.com/openFuyao/env-check/releases/download/latest/bin/linux/amd64/config.json) | [下载](https://openfuyao.obs.cn-north-4.myhuaweicloud.com/openFuyao/env-check/releases/download/latest/bin/linux/arm64/config.json) |

配置文件可以自行新建或者下载默认配置文件进行修改。

## 使用教程

> **说明：**
>
> - 下载后工具名为`envCheck`，且是二进制可执行文件。
> - 可以通过修改配置文件和执行不同的命令使用工具的不同功能。相应的配置文件和工具在同一文件夹下且名为`config.json`（配置文件的位置可以自行放置，但是放置在其余位置时，需要使用命令行参数`--config`显式指定），这里默认在同一文件夹下且内容为下面示例中展示的内容。

### 全量配置文件示例

```json
// config.json
{
  "log_file": "./envCheck.log",
  "output_format": "text",
  "paths": [
    "$HOME/.kube",
    "/etc/kubernetes",
    "/usr/bin/kube*",
    "/usr/local/bin/kube*",
    "/usr/local/bin/crictl",
    "/etc/sysctl.d/k8s.conf",
    "/etc/systemd/system/kubelet.service",
    "/etc/systemd/system/kubelet.service.d",
    "/var/lib/openFuyao/etcd",
    "/var/lib/kubelet",
    "/run/containerd/containerd.sock",
    "/usr/lib/systemd/system/kubelet.service.d",
    "/var/run/containerd/containerd.sock",
    "/var/run/docker.sock"
  ],
  "clean_force": false,
  "program_list": [
    "docker",
    "kubectl",
    "containerd"
  ],
  "hosts": [
    {
      "ip": "192.168.1.1",
      "username": "root",
      "password": "******",
      "port": "22",
      "role": ["bootstrap"]
    },
    {
      "ip": "192.168.1.2",
      "username": "root",
      "password": "******",
      "port": "22",
      "role": ["master"]
    }
  ],
  "clock_threshold": 10
}
```

> **注意：**
>
> 关于配置文件，工具会进行以下校验。
>
> - 使用文件查询或清理功能时`paths`不能为空。
> - 使用冲突应用程序检测功能时`program_list`不能为空。
> - 使用时钟同步检测功能时`hosts`不能为空。

配置文件字段说明：

| 字段名          | 说明                                          |
| --------------- | --------------------------------------------- |
| log_file        | 存放日志文件的路径（包含文件名）              |
| output_format   | 工具执行结果报告输出格式，目前支持 text、json |
| paths           | 文件查询和清理需要指定的文件路径列表          |
| clean_force     | 是否执行强制删除（无需用户确认）              |
| program_list    | 需要检查是否安装的应用名称列表                |
| hosts           | 需要进行时钟同步检测的机器列表                |
| ip              | 用于进行ssh连接的IP地址                       |
| username        | 用于进行ssh连接的用户名                       |
| password        | 用于进行ssh连接的密码                         |
| port            | 用于进行ssh连接的端口                         |
| role            | 机器所属角色（bootstrap 表示是引导节点）      |
| clock_threshold | 机器时钟误差允许阈值（单位：秒）              |

> **说明**：
>
> 工具在运行时，默认会在终端输出运行结果并在当前文件夹下生成对应的html文件，并且运行过程中会将日志记录在当前文件夹下的envCheck.log文件中。用户可以通过修改配置文件中相应字段进行设置，以下章节默认配置文件中`output_format`为`text`。

用户可以执行如下命令获取工具帮助文档。

```bash
envCheck -h
```

### 文件查询功能

- 执行如下命令，进行默认查询。

  ```bash
  envCheck query
  ```

  工具会查询当前文件夹下`config.json`中`paths`指定的路径下是否存在文件，工具运行结束后会在终端输出查询结果并且在当前文件夹下生成`query.html`文件。

- 执行如下命令，在查询时指定配置文件路径。

  ```bash
  envCheck query --config=./queryConfig.json
  ```

  工具会查询当前文件夹下`queryConfig.json`中`paths`指定的路径下是否存在文件，工具运行结束后会在终端输出查询结果并且在当前文件夹下生成`query.html`文件。

### 文件清理功能

> **说明**：
>
> 在执行文件清理功能之前，您需要确认配置文件中的`clean_force`字段值，默认为`false`。`false`表示在删除文件之前会向您进行确认，`true`表示不会进行确认。如果您添加了`--force`执行命令，则表示无论`clean_force`字段为何值，工具都会直接删除对应文件。

- 执行如下命令，进行默认清理（在删除每一文件之前会询问用户）。

  ```bash
  envCheck clean
  ```

  工具会询问是否删除当前文件夹下`config.json`中`paths`指定的路径下的文件，工具运行结束后会在终端输出清理结果并且在当前文件夹下生成`clean.html`文件。

- 执行如下命令，进行强制删除，删除文件不会询问用户，直接进行删除。

  ```bash
  envCheck clean --force
  ```

  工具会直接删除当前文件夹下`config.json`中`paths`指定的路径下的文件，工具运行结束后会在终端输出清理结果并且在当前文件夹下生成`clean.html`文件。

- 执行如下命令，在删除时指定配置文件路径。

  ```bash
  envCheck clean --config=./cleanConfig.json
  ```

  工具会按照`cleanConfig.json`文件中的设置执行文件清理功能，工具运行结束后会在终端输出清理结果并且在当前文件夹下生成`clean.html`文件。

### 程序存在性检测

- 执行如下命令，进行默认查询应用。

  ```bash
  envCheck check
  ```

  工具会查询当前文件夹下`config.json`中`program_list`指定的应用是否安装，工具运行结束后会在终端输出查询结果并且在当前文件夹下生成`check.html`文件。

- 执行如下命令，在查询时指定配置文件路径。

  ```bash
  envCheck check --config=./checkConfig.json
  ```

  工具会查询当前文件夹下`checkConfig.json`中`program_list`指定的应用是否安装，工具运行结束后会在终端输出查询结果并且在当前文件夹下生成`check.html`文件。

### 时钟同步检测功能

**前提条件**

使用时钟同步检测功能需要用户先修改默认配置文件中的相关`host`信息。

**开始使用**

- 执行如下命令，查询其余host与引导节点host的时钟误差。

  ```bash
  envCheck clock
  ```

  工具会根据当前文件夹下`config.json`中的`hosts`和`clock_threshold`判断非引导节点和引导节点（bootstrap）之间的时钟误差不多于指定阈值，工具运行结束后会在终端输出检测结果并且在当前文件夹下生成`clock.html`文件。

- 执行如下命令，查询时指定配置文件。

  ```bash
  envCheck clock --config=./clockConfig.json
  ```

  工具会根据当前文件夹下`clockConfig.json`中的`hosts`和`clock_threshold`判断非引导节点和引导节点（bootstrap）之间的时钟误差不多于指定阈值，工具运行结束后会在终端输出检测结果并且在当前文件夹下生成`clock.html`文件。