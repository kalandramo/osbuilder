# osbuilder-simplified

osbuilder CLI 的**严格简化版**示例，复刻真实仓库的两大核心机制：

1. **命令层级管理**（Cobra 树 + 分组展示）
2. **配置管理**（Cobra + Viper：flag / 配置文件 / 环境变量 三级来源）

## 目录结构

```
osbuilder-simplified/
├── go.mod
├── main.go                         # 入口，对应 cmd/osbuilder/osbuilder.go
└── internal/
    ├── cmd/
    │   ├── cmd.go                   # 根命令 + 分组 + 配置接入，对应 cmd.go
    │   ├── options/options.go       # `osbuilder options` 帮助子命令，对应 cmd/options/options.go
    │   ├── util/factory.go          # Factory 桩 + CheckErr
    │   ├── util/config.go           # OnInitialize 复刻，对应 onexstack/pkg/core/config.go
    │   ├── color/color.go           # 叶子命令 + Options 模式
    │   ├── version/version.go       # 一级叶子，演示读取 Factory 配置
    │   ├── es/es.go                  # 中间节点：管理 ES 资源的【领域配置】范例
    │   ├── es/get/get.go             # 叶子：es get，演示领域配置 + viper 合并
    │   ├── es/options/options.go    # es 领域配置结构体（不进全局 Options）
    │   └── create/
    │       ├── create.go            # 中间节点
    │       └── project/project.go   # 二级叶子
        └── util/options/options.go      # 全局配置结构体 + AddFlags + Complete，对应 util/options/options.go
    ```

    ## es 命令：领域配置（子命令级配置）范例

    `es` 子命令演示「连接 ES 管理 ES 资源」这类**只服务于单一子命令**的配置，属于领域配置，不污染全局 `Options`：

    - 领域配置结构体 `ESOptions` 定义在 `internal/cmd/es/options/options.go`（addr/index/username/password）。
    - flag 挂在 `es get` 自身（`--es.addr` / `--es.index` 等），前缀 `es.` 与全局隔离。
    - 优先级与全局一致：`flag 默认值 < osbuilder.yaml 的 es.* < 环境变量 OSCTL_ES_* < flag 显式值`。
    - `osbuilder options` 只显示全局继承 flag（user.* / config），**不会**出现 es.* —— 证明领域配置解耦成功。

    ```bash
    go run . es get                                      # 用默认值
    go run . es get --es.addr=http://es.prod:9200 --es.index=orders
    OSCTL_ES_ADDR=http://es.prod:9200 go run . es get --es.index=orders
    ```

## 运行

```bash
go run . version
go run . create project -c demo.yaml
go run . color -t bg
```

## 配置机制验证

配置来源优先级（后者覆盖前者）：

```
默认值 < osbuilder.yaml(搜索目录) < 环境变量 OSCTL_* < --config 指定文件 < 命令行 flag
```

```bash
# 1. 默认（无配置）
go run . version

# 2. 配置文件 ./osbuilder.yaml 或 ~/.onexstack/osbuilder.yaml
cat > osbuilder.yaml <<'EOF'
user:
  name: alice
  email: alice@example.com
EOF
go run . version

# 3. 环境变量覆盖（OSCTL_ 前缀，. 和 - 转 _）
OSCTL_USER_NAME=bob go run . version

# 4. 命令行 flag 最高优先级
go run . version --user.name=carol
```

## 与真实仓库的对应

| 真实代码 | 简化版 |
|---------|--------|
| `cmd/osbuilder/osbuilder.go` | `main.go` |
| `internal/osbuilder/cmd/cmd.go`（`NewOSCtlCommand` + `OnInitialize` + `BindPFlags` + `Complete` + `NewFactory(opts)`） | `internal/cmd/cmd.go` |
| `internal/osbuilder/util/options/options.go` | `internal/util/options/options.go` |
| `internal/osbuilder/cmd/options/options.go`（`osbuilder options` 帮助子命令） | `internal/cmd/options/options.go` |
| `onexstack/pkg/core/config.go: OnInitialize` | `internal/cmd/util/config.go` |

## 两个 `options` 包的区别

原仓库有两个同名但职责完全不同的 `options` 包，简化版都已复刻：

| 维度 | `cmd/options`（命令行选项） | `util/options`（配置项） |
|------|------------------------------|---------------------------|
| 路径 | `internal/cmd/options` | `internal/util/options` |
| 作用 | `osbuilder options` 子命令，打印所有命令继承的持久 flag | 全局配置结构体，定义 CLI 配置项并绑定 viper |
| 导出物 | `NewCmdOptions(out)` | `Options`、`NewOptions()`、`AddFlags`、`Complete`、`Validate` |
| 与配置关系 | 无关（只是 help 展示） | 是（`osbuilder.yaml` 的 user 等配置） |

> `cmd.go` 中因两者同名，给配置包起了别名 `clioptions`（`import clioptions "internal/util/options"`），与真实仓库 `cmd.go` 的写法一致。

运行验证：

```bash
go run . options   # 打印 --config / --user.name / --user.email 等继承 flag
```
| `cmdutil.NewFactory(opts)` 注入子命令 | `util/factory.go` |
