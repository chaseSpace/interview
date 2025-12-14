## K8s 面试基础

### 简单介绍 K8s

K8s 是 kubernetes 的简称，其本质是一个开源的容器编排系统，主要用于管理容器化的应用，
其目标是让部署容器化的应用简单并且高效,Kubernetes 提供了应用部署，规划，更新，维护的一种机制。

### K8s 有哪些组件，作用分别是什么

k8s 主要由 master 节点和 node 节点构成。

master 节点负责管理集群，node 节点是容器应用真正运行的地方。

master 节点包含的组件有：kube-api-server、kube-controller-manager、kube-scheduler、etcd。

node 节点包含的组件有：kubelet、kube-proxy、container-runtime。

### 支持哪些容器运行时

主要是通过实现 容器运行时接口 (CRI) 来达成，核心运行时包括 containerd、CRI-O，以及早期的 Docker（通过 Dockershim 适配，现在已移除）。

### 什么是 Pod

Pod 是 Kubernetes 中最小的部署单位，可以理解为容器的一个封装。Pod 中通常包含一个或多个紧密相关的容器，它们共享同一个网络命名空间和存储卷。

**Pod 与容器的关系**

容器是一个独立的进程运行环境，而 Pod 是 Kubernetes 中最小的调度和管理单元，它是一个逻辑概念，可以包含一个或多个紧密关联的容器，
这些容器共享网络、存储，并作为一个整体运行，共同管理应用程序的生命周期。简而言之，容器是执行者，而 Pod 是部署和运行容器的“包装盒”。

### 什么是 Deployment 和 Service

**Deployment** 是一种控制器，用于管理无状态应用的生命周期。它负责确保指定数量的 Pod 副本在集群中运行，并且支持滚动更新、回滚等功能。

**Service** 是 Kubernetes 中的一个抽象层，用于定义一组 Pod 的访问策略。通过 Service，可以暴露应用程序到外部或集群内其他服务。

### 如何在 Kubernetes 中暴露应用？

可以通过 Service 暴露应用。常见的 Service 类型包括 ClusterIP、NodePort、LoadBalancer 和 ExternalName，分别适用于不同的场景。

- ClusterIP：为一组 Pod 提供一个集群内部的虚拟 IP 地址，只能在集群内部访问。
- NodePort：在每个 Kubernetes 节点上打开一个静态端口（默认 30000-32767），将外部流量转发到 ClusterIP。
- LoadBalancer：利用云提供商（如 AWS, GCP, Azure）的外部负载均衡器，对外暴露服务 IP。
- ExternalName：将外部服务映射为 Kubernetes Service，并使用 DNS 映射到外部服务。

### 什么是 ConfigMap 和 Secret？

ConfigMap 用于存储非机密的配置信息，而 Secret 用于存储敏感数据如密码、令牌等。两者都可以作为环境变量或挂载到 Pod 中使用。

### 什么是 DaemonSet？

DaemonSet 确保集群中的每个节点上都有一个 Pod 副本。它通常用于需要在每个节点上运行的服务，例如日志收集和监控。

### 中级问题分界线

### kubelet 的功能、作用是什么？

Kubelet 是 Kubernetes 集群中每个节点上的核心代理，也叫做’节点管家‘。其主要作用是管理节点上的 Pod 生命周期、确保容器按期望运行、汇报节点和
Pod 状态和存储卷管理等，
具体包括：监听 API Server 指令、创建/启动/销毁容器、执行健康检查、管理存储卷、汇报资源使用情况等，是实现 Kubernetes
分布式管理的关键组件。

### kube-api-server 的端口是多少？

Kube-API-Server 主要监听 HTTPS 端口 6443 (默认) 和可选的 HTTP 端口 8080。

**各个 pod 是如何访问 kube-api-server 的？**

Pod 访问 API Server 通常通过集群内部的 Service (如 kubernetes Service，ClusterIP 443)，该 Service 代理到 Master 节点
6443 端口的 API Server Pod，实现透明转发，每个 Pod 都能通过这个虚拟 IP 和端口与 API Server 通信，完成集群管理操作。

**具体原理**

1. Pod 启动时注入 K8s Service 信息环境变量
2. 每个 K8s 集群都有一个内建 Service：`kubernetes`，其 ClusterIP 总是 `10.96.0.1`
3. 这个 kubernetes Service 关联着 API Server Pod (在 kube-system 命名空间中)，它将对 443 端口的请求转发到 API Server Pod
   实际监听的 6443 端口。
4. Pod 向 `kubernetes.default.svc.cluster.local:443` 发送请求，kube-proxy 和 CoreDNS 负责将请求路由到 Master 节点上的 API
   Server 容器，最终到达 6443 端口。

### K8s 中命名空间的作用是什么

- 资源隔离: 将集群划分为独立的逻辑区域，不同命名空间中的同名资源（如 Deployment、Service）不会冲突，且默认情况下跨命名空间资源无法直接访问，实现隔离。
- 权限控制 (RBAC): 允许为不同用户或团队分配对特定命名空间资源的访问权限，实现最小权限原则。
- 资源配额 (Quota): 为每个命名空间设置 CPU、内存、存储等资源的使用上限，防止某个应用或团队耗尽集群资源。
- 组织和管理: 方便按项目、团队或环境（如 dev/test/prod）对资源进行分组和管理，提高可查找性和可操作性。
- 命名作用域: 命名空间内的资源名称必须唯一，但跨命名空间则不要求，提供了一个命名范围。

### K8s Proxy API 接口的作用

主要用于代理 Kubernetes 集群内部的 REST 请求，将 API Server 接收到的请求转发给指定节点上的 kubelet 进程处理，实现对节点和
Pod 的直接、细粒度的管理和调试，
常用于集群外访问 Pod 服务进行管理排查、查看节点级资源（如 Pod 的运行状态、日志）等，是调试、管理集群的重要工具。

### Pod 原理

是利用 Linux 的 Namespaces（如网络、PID）和 Cgroups 技术，通过一个特殊的 pause 容器作为“占位符”来管理共享资源，确保 Pod
内所有容器拥有相同的网络栈和存储，
实现高效的通信和数据共享，常用于实现 Sidecar 等模式。

### Pod 如何进行健康监测

通过 Liveness Probe (存活探针)、Readiness Probe (就绪探针) 和 Startup Probe (启动探针) 三种机制进行健康探测，利用 HTTP
GET、TCP Socket、Exec Command 等方式检测容器状态，实现自动重启不健康实例、控制流量进入等自愈能力，确保服务可用性。

- Liveness Probe: 检测容器是否还在健康运行，如果失败，表明应用可能死锁、崩溃，Kubelet 会根据 restartPolicy 重启容器。
- Readiness Probe: 检测容器是否已准备好接收流量。失败时，Pod 暂时从 Service 的 Endpoints 移除，不接收流量；成功后重新加入。
- Startup Probe: 延迟 Liveness 和 Readiness 探针的启动，用于启动慢的应用。探测成功后，其他探针才开始工作。

### Pod 处于 pending 状态可能有哪些原因

处于 Pending 状态通常是因为调度器无法将 Pod 放置到任何节点上，常见原因包括 节点资源不足（CPU/内存/GPU
不足）、调度约束不满足（nodeSelector,
亲和性, 污点/容忍不匹配）、节点异常（NotReady 状态）、存储问题（PV/PVC 绑定失败）、镜像问题或调度器自身故障。

### Pod 内的初始化容器作用

是在 Pod 主容器启动之前运行的特殊容器，用于执行准备任务，如下载配置文件、初始化数据或等待依赖服务就绪，确保主应用容器启动前满足所有前置条件，
它们按顺序执行直到成功，并且共享卷与主容器，提供一种安全、隔离的方式来做启动前的初始化工作，避免污染主镜像，并可用于实现先决条件检查。

初始化容器一般使用工具容器，如 busybox 这类拥有应用镜像中没有的工具（如 sed, awk, dig）。

> Pod 中的所有初始化容器必须按定义顺序成功退出（返回 0），才能启动应用容器。

### Pod 通常如何限制 cpu 和内存额度

根据应用需求灵活设定，常见的限制范围从很小（如几百 MHz CPU, 几十 MiB 内存）到很大（如几核 CPU,
几个 GB 内存）都有，具体取决于应用是微服务、数据库、还是消息队列等。

- 小型应用（如微服务）：
  CPU: 100m (0.1核) - 500m (0.5核)
  内存: 64Mi - 256Mi

- 中型应用（如 API 网关、缓存）：
  CPU: 500m - 2 核
  内存: 256Mi - 1Gi

### Pod 内定义的 command 和 args 会与 docker 镜像内的 entrypoint 冲突吗

不会。Pod 内定义的 command 和 args 是可选的，若存在，则会覆盖镜像内的 entrypoint。

### Pod 内的 pause 容器作用是什么

是 Pod 的基础和守护者，其核心作用是为 Pod 内的其他业务容器提供共享的 Linux
命名空间（网络、PID、IPC 等）和稳定的运行环境，确保 Pod 拥有唯一的 IP 地址、共享存储卷，并充当 PID 1 进程回收僵尸进程，从而实现
Pod 多容器的隔离与协作。

### Pod 是如何发现 Service 的

一个 Pod 发现 Service 主要通过 Kubernetes 集群内部的 DNS（CoreDNS/Kube-DNS）和环境变量机制，Pod 访问 Service 的域名（如
my-service.my-namespace.svc.cluster.local 或简写 my-service）时，DNS 会解析到 Service 的 ClusterIP (虚拟 IP)，Kube-proxy
负责将请求路由到后端匹配标签的真实 Pod IP，从而实现了服务的发现、负载均衡和稳定访问。

### 如何在集群中代理一个外部的 IP+端口服务

需要定义一个不带 selector 的 ClusterIP 类型的 Service，然后手动创建与之同名的 Endpoint 对象 (或 EndpointSlice) 来指向那个外部
IP 和端口，这使得 K8s 集群内的应用可以像访问内部服务一样，通过 Service 的虚拟 IP 间接访问外部服务。

> 通常是用来数据库、缓存、消息队列等外部服务。

### 无头 Service 的作用是什么

K8s 无头服务 (Headless Service) 是一种特殊类型的 Service，它不分配 ClusterIP (虚拟 IP)，而是通过 DNS 直接暴露后端所有 Pod
的
IP 地址列表，让客户端能绕过 K8s 的代理和默认负载均衡，直接与单个 Pod 通信，常用于有状态应用（如数据库、消息队列）的服务发现和集群管理。

### Deployment 更新的命令有哪些

- `kubectl apply -f ...` 实现配置更新
- `kubectl edit deployment xxx` 实现在线更新
- `kubectl set image ...` 修改镜像来更新

### 如何安全下线 node

kubectl drain 命令。

### 详细介绍 kube-proxy

kube-proxy 是运行在每个节点上的网络代理，它实现 Service 的负载均衡和虚拟 IP（ClusterIP）功能，监听 API
Server 的 Service 和 Endpoints 变化，并利用内核机制（如 iptables 或 IPVS）在节点上配置规则，将发往 Service 的流量重定向到后端的真实
Pod 上，保证服务访问的稳定性和可伸缩性。

**工作原理**

- 监听事件：Kube-proxy 连接到 API Server，监听 Service 和 Endpoints（或 EndpointSlices）的创建、更新、删除事件。
- 配置内核规则：当有新 Service 或 Endpoints 变化时，Kube-proxy 会根据这些信息，在节点上配置 Linux 内核的 netfilter 规则，通常是
  iptables 规则。
- 流量重定向：
    - 当一个 Pod 访问某个 Service 的 ClusterIP:Port时
    - 数据包到达节点后，被 iptables 规则捕获。
    - iptables 规则将目标 IP 和端口转换为一个后端 Pod 的真实 IP:Port。

### kubelet 如何监控节点资源状态

当 kubelet 服务启动时，它会自动启动 cAdvisor 服务，然后 cAdvisor 会实时采集所在节点的性能指标及在节点上运行的容器的性能指标。

> kubelet 的启动参数 --cadvisor-port 可自定义 cAdvisor 对外提供服务的端口号，默认是 4194。

### 简述 Ingress

Ingress 是一个用于管理集群外部流量访问集群内部服务的规则集合，它通过定义路由规则将 HTTP/HTTPS 请求转发到不同的
Service，实现七层负载均衡、域名/路径路由、SSL/TLS 终止等功能，通常需要一个 Ingress Controller（如
Nginx、Traefik）来具体实现这些规则，提供一个统一的入口来暴露服务，比 LoadBalancer/NodePort 更灵活且节省资源。 