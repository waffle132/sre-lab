# Counter 集中日志实验

链路：counter 标准错误访问日志 → Docker → Alloy → Loki → Grafana。
仅用于本地实验；Docker socket 提供很强的 Docker API 权限，ro 挂载不限制 API 操作。

## 启动

```bash
docker network create training_shared
cd labs/docker-transfer
docker compose up -d --build
cd ../logging
docker compose up -d
```

共享网络已存在时跳过创建。计数器镜像对应当前源码 lab-v2；正在运行的旧实验曾使用 lab-v1。
Loki 卷保存日志，Alloy 卷保存采集位置等状态。
本次 Loki 曾因卷根目录 root:root、755 而无法创建 chunks；教练将该卷根目录所有者修正为 10001:10001 后恢复。新环境若遇同样错误，应先核对卷权限。

## Grafana 接入

复用 phase1-server 中的 Grafana。保留其 default 网络，同时添加 shared_net；在该 Compose 顶层声明：

```yaml
networks:
  default: {}
  shared_net:
    external: true
    name: training_shared
```

Grafana 服务的 networks 列表应包含 default 和 shared_net。应用该服务配置后，在 Grafana 添加 Loki 数据源，URL 为 http://loki:3100，再进入 Explore 查询。
Grafana 页面配置保存在其数据卷中，不包含在本次 Git 提交内。

## 已验证与查询

2026-11-21：Alloy 启动；请求 /alloy-check-1121 后，Loki 查询返回对应 404 访问日志；学生在 Grafana 看到新请求并完成 404 筛选。部署与验证有教练指导，不计独立排障通过。

```logql
{service_name="counter"}
{service_name="counter"} |= "/alloy-check-1121"
{service_name="counter"} |= "\" 404 "
{service_name="counter"} != "/healthz"
```

下一步：解释并练习 count_over_time 与 sum，统计最近五分钟的 404 数量；尚未验收。
