# envCheck 工具设计文档 V2.0

## 1、特性描述

在原有功能（文件查询 query、文件清理 clean、程序检查 check、时钟同步检查 clock）基础上，增加以下四个核心功能，使工具具备完整的 K8s 集群部署前置检查能力：

1. **操作系统内核检查（kernel）**：检查内核版本是否符合配置要求（版本号、比较操作符可配置）
2. **端口检查（port）**：检查本地端口占用情况（检查端口列表按角色可配置）
3. **磁盘规划检查（disk）**：检查磁盘空间是否满足配置要求（检查路径和大小可配置，支持按角色过滤）
4. **节点分发执行（dispatch）**：支持将检查工具分发到多个节点并行执行，统一收集结果并生成汇总报告

---

## 2、模块设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              envCheck                                       │
├──────────┬──────────┬──────────┬──────────┬──────────┬──────────┬───────────┤
│  query   │  clean   │  check   │  clock   │  kernel  │   port   │   disk    │
│ (已有)   │ (已有)   │ (已有)   │ (已有)   │  (新增)  │  (新增)  │  (新增)   │
├──────────┴──────────┴──────────┴──────────┴──────────┴──────────┴───────────┤
│                              run / run-local                                │
│                         └────── dispatch ──────┘                            │
│                    ┌──────────────┬──────────────┐                          │
│                    │   SyncTasks  │  AsyncTasks  │                          │
│                    └──────────────┴──────────────┘                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 模块列表

| 模块名   | 功能                              | 对应子命令        |
|----------|-----------------------------------|-------------------|
| query    | 残留文件查询（存在性、权限、属主）| `envCheck query`  |
| clean    | 残留文件清理                      | `envCheck clean`  |
| check    | 冲突程序检查（安装状态、版本）    | `envCheck check`  |
| clock    | 节点间时钟同步检查                | `envCheck clock`  |
| kernel   | 内核版本检查（版本号可配置）      | `envCheck kernel` |
| port     | 端口占用检查（按角色配置端口）    | `envCheck port`   |
| disk     | 磁盘空间检查（路径和大小可配置）  | `envCheck disk`   |
| dispatch | 节点分发、远程执行和结果收集      | `envCheck run` / `envCheck run-local` |

### 2.3 命令执行框架

所有子命令通过 `cmd/utils.go` 中的统一框架执行：

```go
type actionHandler struct {
    execute      func(cmd *cobra.Command, cfg *config.AppConfig, log *logger.Logger) (interface{}, error)
    generateHTML func(result interface{}) error
    format       func(result interface{}, outputFormat string) (string, error)
    baseName     string
}
```

`runAction` 负责统一完成：
1. 加载配置
2. 初始化日志
3. 执行业务逻辑
4. 生成 HTML 报告
5. 格式化输出（text/json）
6. 控制台打印（text 模式）
7. 保存 JSON 文件（json 模式）

---

## 3、详细设计

### 3.1 内核检查模块 (kernel)

#### 3.1.1 功能描述

检查操作系统内核版本是否符合配置文件中指定的版本要求。

- 读取当前系统内核版本（通过 `github.com/shirou/gopsutil/v3/host`）
- 与配置文件中指定的期望版本按字符串进行比较
- 支持比较操作符：`>=`、`>`、`==`、`<=、`<`
- 针对 CentOS 7 的 `3.10.0-xxx` 内核版本做了特殊兼容：若版本字符串长度超过 12 位且第 7~10 位包含 `.`（旧版构建号格式），则在 `>=` / `>` 比较时视为不满足，避免旧构建号被误判为通过

#### 3.1.2 数据结构

```go
// KernelCheckResult 内核检查结果
type KernelCheckResult struct {
    Timestamp  time.Time     `json:"timestamp"`
    KernelInfo KernelInfo    `json:"kernel_info"`
    Summary    KernelSummary `json:"summary"`
}

// KernelInfo 内核基本信息
type KernelInfo struct {
    Version     string `json:"version"`
    OS          string `json:"os"`
    Arch        string `json:"arch"`
    ExpectedVer string `json:"expected_version"`
    Operator    string `json:"operator"`
    IsValid     bool   `json:"is_valid"`
    Error       string `json:"error,omitempty"`
}

// KernelSummary 内核检查汇总
type KernelSummary struct {
    TotalChecked int `json:"total_checked"`
    Passed       int `json:"passed"`
    Failed       int `json:"failed"`
}
```

#### 3.1.3 命令行接口

```bash
# 执行内核版本检查
envCheck kernel

# 指定配置文件
envCheck kernel -c config.json

# 输出格式
envCheck kernel -o json    # 保存为 kernel-check-result.json
envCheck kernel -o html    # 生成 kernel.html
```

#### 3.1.4 配置项

```json
{
  "kernel_check": {
    "min_version": "4.18",
    "operator": ">="
  }
}
```

**配置说明：**
- `min_version`：期望的内核版本（字符串比较），默认 `4.18`
- `operator`：比较操作符，支持 `>=`、`>`、`==`、`<=`、`<`

---

### 3.2 端口检查模块 (port)

#### 3.2.1 功能描述

检查本地端口是否被占用，检查端口列表按角色配置。

- 根据配置文件中按角色划分的端口列表进行检查
- 检查每个端口是否被占用（通过 `net.DialTimeout` 连接 `127.0.0.1:端口`，超时时间可配置，默认 3 秒）
- 支持按节点角色（`bootstrap` / `master` / `worker`）决定执行哪些检查项
- 角色分配逻辑：
  - `bootstrap` → 引导节点端口检查
  - `master` → Master 节点端口检查
  - `worker` → Worker 节点端口检查
- 各角色对应的端口列表相互独立，可配置不同端口号

#### 3.2.2 数据结构

```go
// PortCheckResult 端口检查结果
type PortCheckResult struct {
    Timestamp time.Time       `json:"timestamp"`
    Ports     []PortCheckInfo `json:"ports"`
    Summary   PortSummary     `json:"summary"`
}

// PortCheckInfo 单个端口检查信息
type PortCheckInfo struct {
    Port     string `json:"port"`
    Protocol string `json:"protocol"` // 固定为 tcp
    IsUsed   bool   `json:"is_used"`
    Process  string `json:"process,omitempty"`
    Error    string `json:"error,omitempty"`
}

// PortSummary 端口检查汇总
type PortSummary struct {
    TotalChecked int `json:"total_checked"`
    Used         int `json:"used"`
    Free         int `json:"free"`
}
```

#### 3.2.3 命令行接口

```bash
# 执行端口检查（检查所有角色端口）
envCheck port

# 指定角色检查
envCheck port --roles master

# 指定配置文件
envCheck port -c config.json

# 输出格式
envCheck port -o json    # 保存为 port-check-result.json
envCheck port -o html    # 生成 port.html
```

#### 3.2.4 配置项

```json
{
  "port_check": {
    "ports": {
      "bootstrap": ["36443", "40080", "40443", "38080"],
      "master": ["6443", "30029", "30909", "30903", "3030", "30019", "9012"],
      "worker": ["10250", "10256", "30000", "30001", "30022", "30033", "30024"]
    },
    "timeout": 10
  }
}
```

**配置说明：**
- `ports`：按角色映射的端口列表。分发执行时，Runner 会根据节点角色过滤配置，只保留该节点角色相关的端口。各角色的端口列表相互独立，可配置不同端口号。
- `timeout`：TCP 连接探测的超时时间（秒），默认 3 秒。

---

### 3.3 磁盘检查模块 (disk)

#### 3.3.1 功能描述

检查磁盘空间是否满足配置文件中指定的要求。

- 根据配置文件中指定的路径和最小空间要求进行检查
- 支持按角色配置不同的检查项（`roles` 字段）
- 若路径不存在，自动向上查找最近的父目录进行空间统计
- 若路径是文件，则取其所在目录进行统计
- 输出每个路径的总空间、可用空间、已用百分比、文件系统类型等信息

#### 3.3.2 数据结构

```go
// DiskCheckResult 磁盘检查结果
type DiskCheckResult struct {
    Timestamp time.Time   `json:"timestamp"`
    Spaces    []DiskSpace `json:"spaces"`
    Summary   DiskSummary `json:"summary"`
}

// DiskSpace 磁盘空间检查项
type DiskSpace struct {
    Path         string  `json:"path"`
    Total        uint64  `json:"total_bytes"`
    Free         uint64  `json:"free_bytes"`
    Used         uint64  `json:"used_bytes"`
    UsedPercent  float64 `json:"used_percent"`
    MinFree      uint64  `json:"min_free_bytes"`
    IsSufficient bool    `json:"is_sufficient"`
    Filesystem   string  `json:"filesystem"`
    Error        string  `json:"error,omitempty"`
}

// DiskSummary 磁盘检查汇总
type DiskSummary struct {
    TotalChecked     int `json:"total_checked"`
    SufficientPath   int `json:"sufficient_paths"`
    InsufficientPath int `json:"insufficient_paths"`
}
```

#### 3.3.3 命令行接口

```bash
# 执行磁盘检查
envCheck disk

# 指定配置文件
envCheck disk -c config.json

# 输出格式
envCheck disk -o json    # 保存为 disk-check-result.json
envCheck disk -o html    # 生成 disk.html
```

#### 3.3.4 配置项

```json
{
  "disk_check": {
    "check_items": [
      {"path": "/", "min_free_gb": 50},
      {"path": "/var/lib/docker", "min_free_gb": 80},
      {"path": "/var/lib/kubelet", "min_free_gb": 80},
      {"path": "/var/lib/etcd", "min_free_gb": 50, "roles": ["master"]}
    ]
  }
}
```

**配置说明：**
- `check_items`：磁盘检查项列表，完全可配置
- `path`：需要检查的目录路径
- `min_free_gb`：该路径要求的最小可用空间（GB）
- `roles`（可选）：只在指定角色的节点上检查此项；为空表示所有节点都检查

---

### 3.4 节点分发模块 (dispatch)

#### 3.4.1 功能描述

通过 SSH 将 envCheck 二进制分发到远程节点并执行检查，统一收集结果生成汇总报告。

1. 建立到所有目标节点的 SSH 连接（并发，受 `concurrent_limit` 限制）
2. 检测远程节点架构（`x86_64` → `amd64`，`aarch64` → `arm64`），自动匹配对应架构的二进制文件
3. 清理远程工作目录
4. 分发 envCheck 二进制和配置文件：
   - 二进制采用 gzip 压缩传输
   - 若远程已存在同名文件且大小一致，则跳过上传
5. 在目标节点后台启动 `envCheck run-local` 执行检查
6. 轮询收集各节点的 `result.json` 或 `error.log`
7. 汇总生成统一的 HTML/JSON 报告

#### 3.4.2 执行流程

```
建立SSH连接 --> 架构探测 --> 环境清理 --> 文件分发 --> 远程执行 --> 结果收集 --> 汇总报告
     │             │             │             │             │             │
     v             v             v             v             v             v
 并发连接所有    执行          创建/清空      压缩上传       后台启动      轮询 result.json
 节点（带并发    uname -m      远程工作目录    二进制和配置    run-local    和 error.log
 限制）          映射          chmod 777      文件                         超时自动标记失败
                amd64/arm64
```

#### 3.4.3 数据结构

```go
// DispatchResult 分发执行结果
type DispatchResult struct {
    Timestamp time.Time       `json:"timestamp"`
    Duration  string          `json:"duration"`
    Nodes     []NodeResult    `json:"nodes"`
    Summary   DispatchSummary `json:"summary"`
}

// NodeResult 节点执行结果
type NodeResult struct {
    IP         string        `json:"ip"`
    Role       []string      `json:"role"`
    Status     string        `json:"status"`      // success / failed / running
    StartTime  time.Time     `json:"start_time"`
    EndTime    time.Time     `json:"end_time,omitempty"`
    ResultFile string        `json:"result_file,omitempty"`
    ErrorFile  string        `json:"error_file,omitempty"`
    Error      string        `json:"error,omitempty"`
    Results    []CheckResult `json:"results,omitempty"`
}

// CheckResult 单项检查结果
type CheckResult struct {
    CheckType  string          `json:"check_type"`  // kernel / port / disk / clock
    Status     string          `json:"status"`      // pass / fail / skip
    Detail     string          `json:"detail"`
    DetailData json.RawMessage `json:"detail_data,omitempty"`
}

// DispatchSummary 分发执行汇总
type DispatchSummary struct {
    TotalNodes   int    `json:"total_nodes"`
    SuccessNodes int    `json:"success_nodes"`
    FailedNodes  int    `json:"failed_nodes"`
    RunningNodes int    `json:"running_nodes"`
    Result       string `json:"result,omitempty"` // PASS / FAIL
}
```

#### 3.4.4 命令行接口

```bash
# 执行全量检查（分发到所有节点执行所有检查项）
envCheck run -c config.json

# 执行指定检查项
envCheck run -c config.json --checks kernel,disk,port,clock

# 跳过某些检查项
envCheck run -c config.json --skip clock

# 指定超时时间（秒）
envCheck run -c config.json --timeout 600

# 只检查特定角色的节点
envCheck run -c config.json --roles master,worker

# 远程节点内部调用（隐藏命令）
envCheck run-local --checks kernel,port,disk,clock --roles master
```

#### 3.4.5 配置项

```json
{
  "dispatch": {
    "timeout": 600,
    "poll_interval": 15,
    "work_dir": "/tmp/envcheck",
    "concurrent_limit": 10
  }
}
```

**配置说明：**
- `timeout`：结果收集总超时时间（秒）
- `poll_interval`：轮询结果的时间间隔（秒）
- `work_dir`：远程节点上的工作目录
- `concurrent_limit`：SSH 连接/操作的最大并发数

#### 3.4.6 节点配置裁剪

Runner 在上传配置文件时，会根据节点角色对配置进行裁剪：
- 对于 `port_check`，只保留该节点角色对应的端口列表，避免无关端口被检查
- 磁盘检查的角色过滤在远程节点本地执行时由 `disk.Checker` 处理

---

## 4、输出与报告

### 4.1 输出格式

所有子命令支持以下输出格式（通过 `-o` / `--output` 指定，或在配置文件中设置 `output_format`）：

| 格式   | 说明                                              |
|--------|---------------------------------------------------|
| `text` | 控制台表格输出（默认），使用 `gotable` 格式化     |
| `json` | 保存为 `{baseName}.json` 文件                     |
| `html` | 生成对应的 `.html` 报告文件（通过 embed + template）|

### 4.2 HTML 报告文件

| 子命令      | 生成的 HTML 文件 |
|-------------|------------------|
| `query`     | `query.html`     |
| `clean`     | `clean.html`     |
| `check`     | `check.html`     |
| `clock`     | `clock.html`     |
| `kernel`    | `kernel.html`    |
| `port`      | `port.html`      |
| `disk`      | `disk.html`      |
| `run`       | `dispatch.html`（汇总报告，包含各节点详细结果）|

HTML 模板文件位于 `pkg/output/` 目录下，使用 Go 的 `embed` + `html/template` 技术渲染。

---

## 5、多架构构建

### 5.1 构建说明

项目 `Makefile` 支持以下构建目标：

| 目标          | 说明                              |
|---------------|-----------------------------------|
| `make build`  | 当前平台构建                      |
| `make build-amd64` | linux/amd64 构建             |
| `make build-arm64` | linux/arm64 构建             |
| `make build-all`   | 构建所有支持的平台           |
| `make test`   | 运行测试                          |
| `make clean`  | 清理构建产物                      |

分发执行时，Runner 会自动探测远程节点架构，并在以下路径查找对应二进制：
- `./build/envCheck_amd64`
- `./build/envCheck_arm64`
- 当前可执行文件路径（兜底）

---

## 6、配置文件整合

### 6.1 完整配置示例

```json
{
  "log_file": "./envCheck.log",
  "output_format": "text",
  "paths": [
    "$HOME/.kube",
    "/etc/kubernetes",
    "/usr/bin/kube*",
    "/usr/local/bin/kube*",
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
      "ip": "192.168.2.135",
      "username": "root",
      "password": "******",
      "port": "22",
      "role": ["bootstrap"]
    },
    {
      "ip": "192.168.2.221",
      "username": "root",
      "password": "******",
      "port": "22",
      "role": ["master"]
    },
    {
      "ip": "192.168.2.229",
      "username": "root",
      "password": "******",
      "port": "22",
      "role": ["worker"]
    }
  ],
  "clock_threshold": 10,
  "kernel_check": {
    "min_version": "4.18",
    "operator": ">="
  },
  "port_check": {
    "ports": {
      "bootstrap": ["36443", "40080", "40443", "38080"],
      "master": ["6443", "30029", "30909", "30903", "3030", "30019", "9012"],
      "worker": ["10250", "10256", "30000", "30001", "30022", "30033", "30024"]
    },
    "timeout": 10
  },
  "disk_check": {
    "check_items": [
      {"path": "/", "min_free_gb": 50},
      {"path": "/var/lib/docker", "min_free_gb": 80},
      {"path": "/var/lib/kubelet", "min_free_gb": 80},
      {"path": "/var/lib/etcd", "min_free_gb": 50, "roles": ["master"]}
    ]
  },
  "dispatch": {
    "timeout": 600,
    "poll_interval": 15,
    "work_dir": "/tmp/envcheck",
    "concurrent_limit": 10
  }
}
```

---

## 7、模块接口定义

### 7.1 新增模块接口

```go
// kernel.Checker 内核检查器
func NewChecker(cfg *config.KernelCheckConfig, log *logger.Logger) *Checker
func (c *Checker) Execute() (*config.KernelCheckResult, error)

// port.Checker 端口检查器
func NewChecker(cfg *config.PortCheckConfig, log *logger.Logger) *Checker
func (c *Checker) Execute(roles []string) *config.PortCheckResult

// disk.Checker 磁盘检查器
func NewChecker(cfg *config.DiskCheckConfig, log *logger.Logger, roles []string) *Checker
func (c *Checker) Execute() (*config.DiskCheckResult, error)

// dispatch.Runner 节点分发器
func NewDispatcher(cfg *config.DispatchConfig, appCfg *config.AppConfig, log *logger.Logger) *Runner
func (d *Runner) Execute() (*config.DispatchResult, error)
func SaveResults(results []config.CheckResult, resultPath string) error
```

### 7.2 核心包职责

| 包路径            | 职责                                              |
|-------------------|---------------------------------------------------|
| `cmd/`            | Cobra 子命令定义与路由，包含 `run-local` 隐藏命令 |
| `pkg/kernel/`     | 内核版本读取与比较                                |
| `pkg/port/`       | TCP 端口占用探测，按角色分配检查项                |
| `pkg/disk/`       | 磁盘空间统计，支持路径不存在时回退到父目录        |
| `pkg/dispatch/`   | SSH 连接管理、架构探测、文件分发、远程执行、结果轮询收集 |
| `pkg/output/`     | Text/JSON 格式化输出、HTML 模板渲染               |
| `pkg/config/`     | 配置加载、校验、类型定义                          |
| `pkg/query/`      | 文件查询                                          |
| `pkg/clean/`      | 文件清理                                          |
| `pkg/program/`    | 程序安装状态检查                                  |
| `pkg/clock/`      | 时钟同步检查                                      |
| `pkg/logger/`     | 日志记录                                          |
| `pkg/utils/`      | 路径处理工具                                      |

---

## 8、Story 分解

### Story 1: 内核检查功能
- **功能描述**：用户需要检查各节点的内核版本是否符合配置要求
- **验收标准**：
  - 能正确读取当前系统内核版本
  - 能根据配置的比较操作符进行版本比较
  - 针对 CentOS 7 旧构建号内核做特殊兼容处理
  - 输出清晰的检查结果报告（当前版本、期望版本、比较结果）
  - 版本号完全可配置

### Story 2: 端口检查功能
- **功能描述**：用户需要检查本地端口占用情况
- **验收标准**：
  - 能根据配置文件中按角色划分的端口列表进行检查
  - 能正确检测端口是否被占用（TCP 连接探测）
  - 分发执行时只检查当前节点角色对应的端口
  - 输出端口占用报告
  - 检查端口列表完全可配置

### Story 3: 磁盘检查功能
- **功能描述**：用户需要检查磁盘空间是否满足配置要求
- **验收标准**：
  - 能根据配置文件中指定的路径和最小空间要求进行检查
  - 能检查每个路径的可用空间（路径不存在时回退父目录）
  - 支持按角色检查不同路径
  - 输出磁盘空间检查报告
  - 检查路径和大小完全可配置

### Story 4: 节点分发功能
- **功能描述**：用户需要在多个节点上并行执行检查并统一收集结果
- **验收标准**：
  - 能将工具二进制分发到所有节点（支持架构自动探测、gzip 压缩传输、跳过已存在文件）
  - 能在各节点并行执行指定检查项（通过 `run-local`）
  - 能定时轮询收集检查结果
  - 能汇总生成统一的 HTML/JSON 报告
  - 支持超时控制和错误处理
  - 支持按角色过滤节点和按角色裁剪端口配置

---

## 9、质量属性设计

### 9.1 性能规格

| 规格名称       | 规格指标                      |
|----------------|-------------------------------|
| 单节点检查耗时 | < 10s                         |
| 100节点分发耗时| < 10min                       |
| 并发连接数     | 默认 10 个，可配置            |
| 文件传输优化   | gzip 压缩 + 大小一致跳过上传  |

### 9.2 可靠性设计

- SSH 连接使用密码认证，连接失败会在日志中记录，不影响其他节点
- 节点检查超时自动标记为失败，不影响其他节点
- 结果文件下载失败会通过 `error.log` 兜底获取错误信息
- 二进制文件分发时校验大小，一致则跳过重复上传
- SFTP 上传支持重试（最多 3 次）和大小校验

### 9.3 兼容性设计

- 支持 Linux 主流发行版（CentOS 7/8, Ubuntu 18/20/22, RHEL 7/8）
- 支持 x86_64（映射为 `amd64`）和 aarch64（映射为 `arm64`）架构
- 自动探测远程节点架构并选择对应二进制文件（`build/envCheck_amd64` / `build/envCheck_arm64`）

---

## 10、修改日志

| 版本  | 发布说明                                                   |
|-------|------------------------------------------------------------|
| v2.0  | 新增内核检查、端口检查、磁盘检查、节点分发功能；支持多架构构建 |
