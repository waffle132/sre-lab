# ConfigMap 首页实验

恢复已有教学环境的配置包，不需要重新编写 Deployment、Service 或构建新镜像。
仅用于学习集群，在当前 kubectl context 的 default 命名空间创建/更新 web-demo-page、web-demo、web-demo-svc。
与现有同名对象合并更新，不删除其他资源。

## 下载并启动

在 Killercoda 网页终端运行：

```bash
curl -fL https://raw.githubusercontent.com/waffle132/sre-lab/labs/k8s-configmap/labs/kubernetes/configmap/lab.yaml -o lab.yaml
kubectl apply -n default -f lab.yaml
kubectl rollout status -n default deployment/web-demo --timeout=120s
kubectl get service -n default web-demo-svc
```

用当前输出的 ClusterIP 执行 `curl http://实际IP`，页面应包含 Hello Kubernetes。
这是下载配置文件；Nginx镜像由集群节点的容器运行时自动拉取，不是把ConfigMap打包进镜像。

## 下一项实操

记录 `kubectl get pods -n default -l app=web-demo` 的Pod名称。
只修改lab.yaml顶部ConfigMap的index.html文字，再apply同一个文件。
等待挂载目录中的文件同步后重新访问Service；预期页面变化而Pod名称不变。
ConfigMap数据变更不会自动触发Deployment滚动更新，目录挂载同步不是即时的。
应用是否采用新配置取决于其读取方式，本例使用Nginx静态文件。

## 清理

```bash
kubectl delete -n default -f lab.yaml
```

此命令删除本文件定义的三个对象，包括web-demo下的Pod；不要在需要保留它们时执行。
本配置包的静态检查不等于在线环境实操验收；首次下载后仍需实际观察页面。
