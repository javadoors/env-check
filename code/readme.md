# **env-check** 仓库架构与功能清单梳理

## 🏗️ 整体架构
```
env-check/
├── cmd/                    # 命令行入口层 (Cobra CLI)
│   ├── main.go             # 根命令 & 子命令注册
│   ├── config.go           # 配置加载逻辑
│   ├── utils.go            # 通用运行器 (runAction 模板方法)
│   ├── query.go            # query 子命令
│   ├── clean.go            # clean 子命令
│   ├── check.go            # check 子命令 (程序存在性)
│   ├── clock.go            # clock 子命令
│   ├── kernel.go           # kernel 子命令
│   ├── port.go             # port 子命令
│   ├── disk.go             # disk 子命令
│   └── run.go              # run / run-local 子命令 (远程调度)
│
├── pkg/                    # 核心业务逻辑层
│   ├── config/             # 配置类型定义 & 校验
│   │   ├── types.go        # 所有数据结构 (AppConfig, Result 类型)
│   │   └── config.go       # 配置加载、路径展开、校验
│   ├── query/              # 文件残留查询
│   ├── clean/              # 文件残留清理
│   ├── program/            # 冲突应用检测
│   ├── clock/              # 时钟同步检测
│   ├── kernel/             # 内核版本检测
│   ├── port/               # 端口占用检测
│   ├── disk/               # 磁盘空间检测
│   ├── dispatch/           # 远程调度执行
│   ├── output/             # 输出格式化 & HTML 报告生成
│   │   ├── output.go       # Formatter (text/json 双格式)
│   │   ├── html_generator.go  # HTML 报告生成 (embed 模板)
│   │   └── *.tpl           # 各模块 HTML 模板
│   ├── logger/             # 日志模块 (控制台彩色 + 文件)
│   └── utils/              # 工具库 (路径展开)
│
├── config.json             # 默认配置文件
├── build/Dockerfile        # 构建镜像
└── go.mod                  # Go 1.24.5
```

### 架构设计模式
| 设计模式 | 体现位置 | 说明 |
|---------|---------|------|
| **模板方法** | [cmd/utils.go](file:///d:/code/github/env-check/cmd/utils.go) `runAction` | 统一执行流程：加载配置 → 初始化日志 → 执行业务 → 生成HTML → 格式化输出 → 保存文件 |
| **策略模式** | [cmd/utils.go](file:///d:/code/github/env-check/cmd/utils.go) `actionHandler` | 每个子命令注入 `execute`/`generateHTML`/`format` 三个回调 |
| **工厂模式** | 各 `pkg/` 模块的 `New*` 构造函数 | 如 `NewQuery`、`NewCleaner`、`NewChecker` 等 |
| **并发模式** | [pkg/clock/clock.go](file:///d:/code/github/env-check/pkg/clock/clock.go)、[pkg/dispatch/dispatch.go](file:///d:/code/github/env-check/pkg/dispatch/dispatch.go) | `sync.WaitGroup` + goroutine 并发获取多节点时间/调度执行 |

## 📋 功能清单

### 一、文件残留查询 (`envCheck query`)
| 项目 | 说明 |
|------|------|
| **源码** | [pkg/query/query.go](file:///d:/code/github/env-check/pkg/query/query.go) |
| **功能** | 检查配置文件中 `paths` 列表指定的路径是否存在残留文件/目录 |
| **检测内容** | 路径是否存在、类型(文件/目录)、属主(owner)、属组(group)、权限(permissions) |
| **路径展开** | 支持 `$HOME` 环境变量展开、`~` 家目录展开、`*` 通配符匹配 (如 `/usr/bin/kube*`) |
| **输出** | 终端表格 + `query.html` 报告 + 可选 JSON 文件 |
| **配置校验** | `paths` 不能为空 |

### 二、文件残留清理 (`envCheck clean`)
| 项目 | 说明 |
|------|------|
| **源码** | [pkg/clean/clean.go](file:///d:/code/github/env-check/pkg/clean/clean.go) |
| **功能** | 删除配置文件中 `paths` 列表指定的残留文件/目录 |
| **安全模式** | `clean_force=false` 时逐个询问用户确认 (`y/n`) |
| **强制模式** | `--force` 参数或 `clean_force=true` 直接删除不询问 |
| **输出** | 终端表格 (已删除/失败/跳过) + `clean.html` 报告 |
| **删除策略** | 目录用 `os.RemoveAll`，文件用 `os.Remove` |

### 三、冲突应用检测 (`envCheck check`)
| 项目 | 说明 |
|------|------|
| **源码** | [pkg/program/check.go](file:///d:/code/github/env-check/pkg/program/check.go) |
| **功能** | 检查环境中是否安装了指定程序（如 docker、kubectl、containerd） |
| **检测方式** | `exec.LookPath` 查找程序路径 + 执行版本命令获取版本号 |
| **版本获取** | 针对不同程序使用不同命令（如 `go version`、`kubectl version --client`、`java -version`） |
| **输出** | 终端表格 (程序名/状态/版本/路径) + `check.html` 报告 |
| **配置校验** | `program_list` 不能为空 |

### 四、时钟同步检测 (`envCheck clock`)
| 项目 | 说明 |
|------|------|
| **源码** | [pkg/clock/clock.go](file:///d:/code/github/env-check/pkg/clock/clock.go) |
| **功能** | 检测集群中非引导节点与引导节点(bootstrap)之间的时钟误差 |
| **检测方式** | SSH 连接远程主机执行 `date +%s` 获取 Unix 时间戳，与 bootstrap 节点对比 |
| **本地优化** | 自动识别本机 IP（`isLocalHost`），本机直接取本地时间无需 SSH |
| **并发** | 多节点并发获取时间（`sync.WaitGroup`） |
| **判定** | 误差 ≤ `clock_threshold`（秒）为同步，否则为未同步 |
| **输出** | 终端表格 + `clock.html` 报告 |
| **配置校验** | `hosts` 不能为空，且必须包含 bootstrap 角色 |

### 五、内核版本检测 (`envCheck kernel`)
| 项目 | 说明 |
|------|------|
| **源码** | [pkg/kernel/check.go](file:///d:/code/github/env-check/pkg/kernel/check.go) |
| **功能** | 检查当前系统内核版本是否满足最低要求 |
| **检测方式** | 使用 `gopsutil/host` 获取内核版本，与 `min_version` 做字符串比较 |
| **比较运算符** | 支持 `>=`、`>`、`==`、`<=`、`<` |
| **特殊处理** | 对 CentOS 7 的 `3.10.0` 旧内核构建号格式做了兼容处理 |
| **默认配置** | `min_version: "4.18"`, `operator: ">="` |
| **输出** | 终端表格 + `kernel.html` 报告 |

### 六、端口占用检测 (`envCheck port`)
| 项目 | 说明 |
|------|------|
| **源码** | [pkg/port/check.go](file:///d:/code/github/env-check/pkg/port/check.go) |
| **功能** | 检查本机指定端口是否被占用 |
| **检测方式** | TCP `DialTimeout` 连接 `127.0.0.1:port`，连接成功则端口被占用 |
| **角色区分** | 按 bootstrap/master/worker 角色配置不同端口列表 |
| **`--roles` 过滤** | 支持通过 `--roles` 参数指定只检查特定角色的端口 |
| **超时** | 默认 3 秒 TCP 连接超时 |
| **输出** | 终端表格 (端口/协议/状态/进程) + `port.html` 报告 |

### 七、磁盘空间检测 (`envCheck disk`)
| 项目 | 说明 |
|------|------|
| **源码** | [pkg/disk/check.go](file:///d:/code/github/env-check/pkg/disk/check.go) |
| **功能** | 检查指定路径的磁盘剩余空间是否满足最低要求 |
| **检测方式** | 使用 `gopsutil/disk` 获取磁盘使用情况 |
| **角色过滤** | 每个检查项可配置 `roles`，只对匹配角色的节点生效 |
| **路径回退** | 若指定路径不存在，自动向上查找父目录进行检测 |
| **默认配置** | `/` ≥ 50GB, `/var/lib` ≥ 50GB, `/var/run` ≥ 50GB |
| **输出** | 终端表格 (路径/文件系统/总量/可用/已用/使用率/状态) + `disk.html` 报告 |

### 八、远程调度执行 (`envCheck run` / `envCheck run-local`)
| 项目 | 说明 |
|------|------|
| **源码** | [pkg/dispatch/dispatch.go](file:///d:/code/github/env-check/pkg/dispatch/dispatch.go)、[cmd/run.go](file:///d:/code/github/env-check/cmd/run.go) |
| **功能** | 将 envCheck 二进制分发到远程节点，执行全量检查并收集结果 |
| **执行流程** | SSH建连 → 架构检测 → 远程环境清理 → 文件分发 → 远程执行 → 结果收集 |
| **文件分发** | SFTP 上传 + gzip 压缩传输 + 大小校验 + 重试机制 (3次) |
| **架构自适应** | 自动检测远程节点架构 (`uname -m`)，选择对应二进制 (`envCheck_amd64`/`envCheck_arm64`) |
| **并发控制** | `concurrent_limit` 限制并发连接数，使用 semaphore 模式 |
| **配置生成** | 为每个节点按其角色生成专属配置（过滤端口列表等） |
| **`--checks`** | 指定要执行的检查项：`kernel,port,disk,clock,fileQuery,programCheck` |
| **`--skip`** | 指定要跳过的检查项 |
| **`--roles`** | 按角色过滤目标节点 |
| **`run-local`** | 内部隐藏命令，远程节点本地执行检查并将结果写入 `result.json` |
| **输出** | 终端汇总表格 + `dispatch.html` 报告 |

## 🔧 基础设施模块

### 配置管理 ([pkg/config](file:///d:/code/github/env-check/pkg/config))
| 功能 | 说明 |
|------|------|
| 配置加载 | `LoadConfigFromFile` 从 JSON 文件加载 |
| 路径展开 | `ExpandPaths` 展开环境变量、`~`、通配符 |
| 配置校验 | `ValidateConfig` 按模式校验必填字段 |
| 类型定义 | [types.go](file:///d:/code/github/env-check/pkg/config/types.go) 定义了全部数据结构（9 种 Result 类型 + 对应 Summary） |

### 日志模块 ([pkg/logger](file:///d:/code/github/env-check/pkg/logger))
| 功能 | 说明 |
|------|------|
| 四级日志 | INFO(蓝) / SUCCESS(绿) / WARNING(黄) / ERROR(红) |
| 双输出 | 控制台彩色输出 + 文件追加写入 |
| 线程安全 | `sync.Mutex` 保护文件写入 |

### 输出模块 ([pkg/output](file:///d:/code/github/env-check/pkg/output))
| 功能 | 说明 |
|------|------|
| 双格式 | `text`（终端表格，使用 `gotable`）和 `json`（`json.MarshalIndent`） |
| HTML 报告 | 使用 Go `embed` 嵌入模板，9 个 `.tpl` 模板对应 9 种报告 |
| 模板函数 | 自定义 `divideFloat`、`getPortDetails`、`getDiskDetails` 等辅助函数 |

### 路径工具 ([pkg/utils](file:///d:/code/github/env-check/pkg/utils))
| 功能 | 说明 |
|------|------|
| 环境变量展开 | `$HOME`、`$USER` 等 |
| 家目录展开 | `~` 和 `~/path` |
| 通配符展开 | `*`、`?`、`[...]`，使用 `filepath.Glob` |

## 📊 配置文件结构总览
```json
{
  "log_file": "./envCheck.log",          // 日志路径
  "output_format": "text",               // 输出格式: text / json
  "paths": [...],                        // 文件查询/清理路径列表
  "clean_force": false,                  // 强制删除开关
  "program_list": [...],                 // 冲突应用检测列表
  "hosts": [{ip, username, password, port, role}],  // 主机列表
  "clock_threshold": 10,                 // 时钟同步阈值(秒)
  "kernel_check": {min_version, operator},          // 内核版本要求
  "port_check": {ports, timeout},        // 端口检测配置
  "disk_check": {check_items},           // 磁盘检测配置
  "dispatch": {timeout, poll_interval, work_dir, concurrent_limit, checks, skip_checks}  // 调度配置
}
```

## 🧪 测试覆盖
每个 `pkg/` 模块均有对应的 `*_test.go` 文件，`cmd/` 层使用 Ginkgo/Gomega BDD 测试框架（[cmd/cmd_suite_test.go](file:///d:/code/github/env-check/cmd/cmd_suite_test.go)）。
