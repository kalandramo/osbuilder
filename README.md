# osbuilder: onexstack 技术栈脚手架工具

**osbuilder：** onexstack 技术栈使用的 Go 项目开发脚手架。

## onexstack 技术栈介绍

onexstack 是一整套 Go 开发技术栈。该技术栈包括了以下内容：
- 学习社群（欢迎加入）：[云原生 AI 实战营](https://t.zsxq.com/5T0qC)
- 高质量的 Go 项目：[「云原生 AI 实战营」项目介绍](https://konglingfei.com/cloudai/project/cloudai.html)
- 高质量的课程：[「云原生 AI 实战营」体系课介绍](https://konglingfei.com/cloudai/catalog/cloudai.html)
- 一系列开发规范：[技术栈相关规范](https://konglingfei.com/onex/convention/rest.html)
- 一系列开发标准包/工具：[onexstack 标准化包](https://github.com/onexstack/onexstack)

onexstack 技术栈中，所有的 Web 服务器类型的项目都是使用 `osbuilder` 脚手架自动生成，例如：[miniblog](https://github.com/onexstack/miniblog)。

## osbuilder 工具介绍

### osbuilder 工具介绍

osbuilder 是一个 Go 项目开发脚手架，可以一键生成一个符合 Go 最佳实践的 Go 项目。该项目集合了我过去对 Go 项目开发、对技术、对架构的思考和经验。个人感觉生成的 Go 项目从代码质量、扩展能力、灵活性等方面，都处在一个很不错的水平。是个非常值得学习的 Go 项目构建方式。

osbuilder 具有以下功能特点：
- 支持一键生成以下 7 类应用类型标准化代码：
  - Web服务器：通用 HTTP/RESTful API 服务；
  - WebSocket服务器：高性能实时双向通信服务；
  - 消息处理服务器（基于 kafka）：事件驱动的消息队列 (MQ) 消费者；
  - 异步任务服务器：分布式定时任务与后台作业处理；
    - 支持标准化的 CronJob、Job；
    - 支持完全自定义的作业；
  - 命令行工具：标准 CLI 命令行交互工具；
  - 声明式API服务器（待实现）：类 K8s 风格的资源管理服务端；
  - 声明式API控制器（待实现）：资源状态调和 (Reconcile) 自动化控制器；
- 支持给 Web服务器、Job服务器、MQ 服务器新增资源实现；支持给命令行工具，新增命令；
- 支持一条命令生成一个可直接运行的高质量、高扩展、标准、符合 Go 开发最佳实践的 Go 项目；
- 支持一条命令添加多个 REST 资源的代码实现；
- 支持不同的 Web 框架，例如：**gin**、**grpc**、kratos、kitex、hertz、go-zero、echo、iris等；
- 支持不同的存储后端，例如：**memory**、**mariadb/mysql**、**sqlite**、**postgresql**、mongo、etcd、redis 等；
- 支持自动添加健康检查接口；
- 支持一键实现带用户管理、认证、鉴权功能的 Web 服务；
- 支持符合 OpenTelemetry 规范的全链路可观测，包括：Tracing、Metrics、Logs，并支持生成示例 Metrics 代码；
- 支持自动注册到不同的服务中心，例如：**polaris**、nacos、consul、eureka；
- 支持生成符合最佳实践的 Dockerfile，包括：debug 镜像和 distroless 镜像，并生成 `make image` 构建镜像规则；
- 支持自动生成高质量、结构化的 Makefile 文件，并且自动生成常用的 Makefile 规则：
- 支持指定 Go 模块名；
- 生成匹配、丰富的 README.md 文件；
- 支持消息队列服务器（消费 Kafka 事件）
- 支持连接不同的外部 HTTP 客户端
- 支持数据预加载代码示例；
- 使用 `osbuilder create quickstart` 快速创建一个示例 Go 项目；
- 支持完整的项目发布能力：自动生成语义化的标签、生成 CHANGELOG、执行发布等；
- 自动生成符合行业最佳实践的标准 Dockerfile；
- Dockerfile 支持多阶段构建和本地构建；
- 支持自动生成 Kubernetes 部署 YAML 文件；

生成的 Go 项目具有以下特点：
- 高质量、高扩展、简洁；
- 目录结构、代码架构、代码实现均符合 Go 编码规范及最佳实践；

### 安装

```bash
$ go install github.com/onexstack/osbuilder/cmd/osbuilder@latest
$ osbuilder version
```

## osbuilder 脚手架使用

osbuilder 脚手架可以用来生产一个新的项目，也能够基于已有的项目添加新的 REST 资源。


### 1. 生成新项目

```bash
$ mkdir -p $GOPATH//src/github.com/onexstack
$ cd $GOPATH//src/github.com/onexstack
$ curl -fsSL https://raw.githubusercontent.com/onexstack/osbuilder/master/internal/osbuilder/tpl/project.yaml -o project.yaml
$ osbuilder create project --config project.yaml ./miniblog
...
🍺 Project creation succeeded miniblog
💻 Use the following command to start the project 👇:
...
🤝 Thanks for using osbuilder.
👉 Visit https://t.zsxq.com/5T0qC to learn how to develop miniblog project.
```

执行上述命令后，可以根据提示，执行以下命令来部署并测试服务：
```bash
$ cd ./miniblog # enter project directory
$ make deps # (Optional, executed when dependencies missing) Install tools required by project.
$ make protoc.apiserver # generate gRPC code
$ go mod tidy # tidy dependencies
$ go generate ./... # run all go:generate directives
$ make build BINS=mb-apiserver # build mb-apiserver
$ _output/platforms/linux/amd64/mb-apiserver # run the compiled server
$ curl http://127.0.0.1:5555/healthz # run health client to test the API
{"timestamp":"2025-08-24 13:23:19"}
```

可以看到，整个项目的生成过程很丝滑，而且生成的项目跟 [miniblog](https://github.com/onexstack/miniblog) 保持高度一致。miniblog 项目有完整的开发体系课，想学习的可以加入 [云原生 AI 实战营](https://t.zsxq.com/5T0qC)。


> 提示：如果想生产带认证鉴权的项目实例，需要设置：webserver[0].withUser 为 `true`。

### 2. 基于已有项目添加新的 REST 资源

```bash
# -b 选项指定给 mb-apiserver 资源添加新的 REST 资源：
# - post：文章
# - comment：评论
# - tag：标签	
# - follow：关注
# - follower：粉丝
# - friend：好友
# - block：黑名单
# - like：点赞	
# - bookmark：收藏
# - share：分享
# - report：举报
# - vote：投票
$ kinds="post,comment,tag,follow,follower,friend,block,like,bookmark,share,report,vote"
$ osbuilder create api -b mb-apiserver --kinds $kinds
```

上述命令会添加 2 个新的 REST 资源：CronJob、Job。接下来，你只需要添加核心业务逻辑即可。

执行完 `osbuilder` 命令之后，会提示如何进行编译。按提示编译并测试：
```bash
$ make protoc.apiserver 
$ make build BINS=mb-apiserver
$ _output/platforms/linux/amd64/mb-apiserver
# 提示：如果指定了 withUser: true，则需要给 HTTP 客户端添加认证信息，否则会报：Unauthenticated 错误
# 创建一个空的文章（文章内容为空），具体调用的接口，可以查看 scripts/startup-test.sh 脚本
$ sh scripts/startup-test.sh posts create '{}'
X-Trace-Id: 64c2835d72bb15fc07765de10e6283a1
-----------------------------
{
  "postID": "post-zhwu4c"
}
$ sh scripts/startup-test.sh posts get 'post-zhwu4c' # 获取刚创建的文章详情，传入文章 ID
X-Trace-Id: 95c631460b60aa91ccb477380a8521ba
-----------------------------
{
  "post": {
    "postID": "post-zhwu4c",
    "createdAt": {
      "seconds": 1761728366
    },
    "updatedAt": {
      "seconds": 1761728366,
      "nanos": 834460375
    }
  }
}
```

### 3. 根据需要添加 REST 资源的具体业务逻辑

接下来，只需要根据需要实现 REST 资源的具体业务逻辑即可。例如 修改：`internal/<component_name>/biz/v1/<rest_name>/<rest_name>.go`。

## 快速创建一个示例 Go 项目

osbuilder 脚手架支持一个命令，直接创建一个可运行、可测试的企业级 Go 项目框架，创建方式如下：
```bash
$ osbuilder create quickstart # 根据提示编译项目即可
```

上述命令会在当前目录下创建一个 `miniblog` 项目，按照命令行提示，即可完成项目的编译、运行和测试。

你可以执行 `osbuilder create quickstart -h` 定制更多项目参数。
