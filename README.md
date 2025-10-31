Traefik v3 插件：Rewrite Request Body Key

将 JSON 请求体中的某个键名改为新键名，例如把 oldKey 改为 newKey。

功能
- 仅在 Content-Type 包含 application/json 且未压缩的请求上生效
- 仅当顶层 JSON 为对象时改名（数组等顶层结构不会处理）
- 自动更新 Content-Length

配置
- oldKey: 需要被改名的原键名
- newKey: 新键名

作为 Traefik 插件使用（file provider 示例）

```yaml
http:
  middlewares:
    rewrite-body-key:
      plugin:
        traefik-plugin-rewrite-request-body:
          oldKey: "oldKey"
          newKey: "newKey"

  routers:
    my-router:
      rule: "PathPrefix(`/api`)"
      service: my-service
      middlewares:
        - rewrite-body-key

  services:
    my-service:
      loadBalancer:
        servers:
          - url: "http://backend:8080"
```

其中 traefik-plugin-rewrite-request-body 的名称需与实际插件模块名一致（go.mod 的 module 路径）。

行为说明
- 请求 Content-Type 非 JSON、或 Content-Encoding 非空（如 gzip）时不处理
- 顶层为对象且存在 oldKey 时：
  - 将其移动到 newKey（覆盖同名键）
  - 删除原 oldKey

开发

```bash
go build ./...
```

许可
MIT


