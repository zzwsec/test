# NGINX ModSecurity WAF 镜像

该目录使用 Alpine 多阶段构建 NGINX，并将 ModSecurity v3、ModSecurity-nginx、OWASP Core Rule Set（CRS）和 WordPress 规则排除插件静态编译或打包进运行时镜像。

镜像包含 WAF 能力和规则文件，但 **默认 NGINX 配置不会自动启用 ModSecurity**。启用 WAF 时需要配置 `modsecurity on` 和 `modsecurity_rules_file`。

## 包含组件

| 组件 | 默认版本 | 用途 |
|---|---:|---|
| Alpine | 3.24.1 | 构建及运行时基础镜像 |
| NGINX | 1.31.3 | Web 服务器和反向代理 |
| ModSecurity | 3.0.16 | WAF 引擎 |
| ModSecurity-nginx | 1.0.4 | NGINX 与 ModSecurity 的连接器 |
| OWASP CRS | 4.28.0 | 通用 Web 攻击检测规则 |
| WordPress Rule Exclusions | 1.2.0 | 降低 WordPress 常见请求的误报 |

## 构建

在仓库根目录执行：

```shell
docker build \
    --tag zzwsec/nginx:1.31.3-waf \
    nginx/static-waf
```

主要构建参数：

| 参数 | 默认值 | 说明 |
|---|---|---|
| `ALPINE_IMAGE` | `alpine:3.24.1` | 构建阶段和运行阶段的基础镜像 |
| `NGINX_VERSION` | `1.31.3` | NGINX 源码版本 |
| `MODSECURITY_VERSION` | `3.0.16` | ModSecurity 源码版本 |
| `MODSECURITY_NGINX_VERSION` | `1.0.4` | ModSecurity-nginx 版本 |
| `CORE_RULESET_VERSION` | `4.28.0` | OWASP CRS 版本 |
| `WORDPRESS_RULE_EXCLUSIONS_VERSION` | `1.2.0` | WordPress 排除插件版本 |

## 启用 WAF

创建 `default.conf`：

```nginx
upstream backend {
    server app:8080;
}

server {
    listen 80;
    server_name example.com;

    modsecurity on;
    modsecurity_rules_file /etc/nginx/modsec/modsecurity.conf;

    # 结合业务设置。NGINX 默认值是 1 MiB。
    client_max_body_size 50m;

    location / {
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_pass http://backend;
    }
}
```

运行容器：

```shell
docker run --rm \
    --name nginx-waf \
    --publish 8080:80 \
    --mount type=bind,src="$PWD/default.conf",dst=/etc/nginx/conf.d/default.conf,readonly \
    zzwsec/nginx:1.31.3-waf
```

示例中的 `app:8080` 表示后端容器的服务名和端口。通过该名称访问时，NGINX 和后端容器需要位于同一个 Docker 网络。

配置检查：

```shell
docker run --rm \
    --mount type=bind,src="$PWD/default.conf",dst=/etc/nginx/conf.d/default.conf,readonly \
    zzwsec/nginx:1.31.3-waf nginx -t
```

## WAF 文件布局

```text
/etc/nginx/modsec/
├── modsecurity.conf
├── unicode.mapping
├── crs-setup.conf
├── plugins/
│   ├── anomaly-threshold-config.conf
│   ├── wordpress-rule-exclusions-config.conf
│   ├── wordpress-rule-exclusions-before.conf
│   └── placeholder-after.conf
└── rules/
    └── *.conf
```

加载顺序为：

1. `crs-setup.conf`
2. 插件 `*-config.conf`
3. 插件 `*-before.conf`
4. CRS `rules/*.conf`
5. 插件 `*-after.conf`

自定义排除规则可放在单独的 before/after 文件中。直接修改 CRS 自带规则会使本地修改与后续 CRS 版本更新混合。

## 当前默认参数

| 参数 | 当前值 | 说明 |
|---|---:|---|
| `SecRuleEngine` | `On` | 加载配置后直接进入阻断模式 |
| `SecRequestBodyLimit` | 50 MiB | 包含上传文件的请求体上限 |
| `SecRequestBodyNoFilesLimit` | 512 KiB | 排除上传文件后的请求数据上限 |
| `SecPcreMatchLimit` | 20000 | 单次正则匹配上限 |
| `SecPcreMatchLimitRecursion` | 20000 | 该参数保留在配置中；ModSecurity v3 不支持该指令的递归限制功能 |
| 入站异常阈值 | 30 | 累计入站评分达到该值时阻断 |
| 出站异常阈值 | 16 | 累计出站评分达到该值时阻断 |
| `SecAuditEngine` | `Off` | 不生成 ModSecurity 审计日志 |

CRS 默认的入站和出站异常阈值分别为 `5` 和 `4`。CRS 中 critical、error、warning 和 notice 级别的默认分值分别为 `5`、`4`、`3` 和 `2`。当前入站阈值 `30` 需要累计评分达到 30 才会阻断，出站阈值 `16` 需要累计评分达到 16 才会阻断。`SecAuditEngine Off` 表示不写入 ModSecurity 审计日志；CRS 规则自身的日志动作仍可将告警写入 NGINX error log。

## 运行模式配置

### DetectionOnly 模式

`DetectionOnly` 模式执行规则检测和日志动作，但不执行规则中的阻断动作。先从镜像复制 `modsecurity.conf`：

```shell
docker run --rm zzwsec/nginx:1.31.3-waf \
    cat /etc/nginx/modsec/modsecurity.conf > modsecurity.conf
```

将规则引擎设置为 `DetectionOnly`，并按日志需求设置审计模式：

```text
SecRuleEngine DetectionOnly
SecAuditEngine RelevantOnly
```

以下内容将异常阈值设置为 CRS 默认值：

```text
SecAction "id:900110,phase:1,pass,t:none,nolog,setvar:tx.inbound_anomaly_score_threshold=5,setvar:tx.outbound_anomaly_score_threshold=4"
```

将修改后的文件只读挂载到容器：

```shell
docker run --rm \
    --publish 8080:80 \
    --mount type=bind,src="$PWD/default.conf",dst=/etc/nginx/conf.d/default.conf,readonly \
    --mount type=bind,src="$PWD/modsecurity.conf",dst=/etc/nginx/modsec/modsecurity.conf,readonly \
    --mount type=bind,src="$PWD/anomaly-threshold-config.conf",dst=/etc/nginx/modsec/plugins/anomaly-threshold-config.conf,readonly \
    zzwsec/nginx:1.31.3-waf
```

规则命中信息包含规则 ID、匹配变量和请求信息，可作为编写规则排除项的依据。

### On 模式

`On` 模式执行规则检测，并执行达到异常阈值后的阻断动作：

```text
SecRuleEngine On
SecAuditEngine RelevantOnly
```

CRS 默认异常阈值为：

```text
inbound_anomaly_score_threshold=5
outbound_anomaly_score_threshold=4
```

异常阈值越高，需要同时命中的规则或累计的异常分值越多。阈值变更通过 `anomaly-threshold-config.conf` 生效。

## 请求体和上传限制

请求体大小可能同时受以下参数限制：

| 层级 | 常用参数 |
|---|---|
| NGINX | `client_max_body_size` |
| ModSecurity | `SecRequestBodyLimit`、`SecRequestBodyNoFilesLimit` |
| PHP | `upload_max_filesize`、`post_max_size` |

如果 NGINX 没有显式设置 `client_max_body_size`，其默认值为 1 MiB。超过该值的请求由 NGINX 返回 413，不会进入后续的 ModSecurity 请求体大小检查。PHP 请求还会受到 `upload_max_filesize` 和 `post_max_size` 限制。

`client_max_body_size` 可配置在 `http`、`server` 或 `location` 上下文。`SecRequestBodyLimit` 统计完整请求体，`SecRequestBodyNoFilesLimit` 统计扣除上传文件后的请求体数据。在当前 `SecRuleEngine On` 和 `SecRequestBodyLimitAction Reject` 配置下，超过相应限制的请求会被返回 413；`DetectionOnly` 模式会将请求体限制动作切换为部分处理。

## WordPress 排除插件

镜像包含 WordPress 规则排除插件。加载 `/etc/nginx/modsec/modsecurity.conf` 时，该插件随插件目录一起加载，并默认启用。它会针对 WordPress 登录、后台、REST API 和编辑器等请求移除部分 CRS 检查目标。

插件规则按请求路径和参数匹配，不会自动识别后端应用类型。多站点配置可以通过事务变量 `tx.wordpress-rule-exclusions-plugin_enabled` 控制每个请求是否启用该插件，也可以为不同 `server` 加载不同的 ModSecurity 配置。

禁用 WordPress 排除插件可在其 config 文件中设置：

```text
SecAction "id:9507010,phase:1,pass,t:none,nolog,setvar:'tx.wordpress-rule-exclusions-plugin_enabled=0'"
```

如果仅对部分域名启用，可以在插件 config 文件中根据请求对应的虚拟主机设置 `tx.wordpress-rule-exclusions-plugin_enabled`。用于匹配的域名需要与 NGINX `server_name` 配置保持一致。

## 日志

CRS 规则的日志动作将告警写入 NGINX error log。镜像把 `/var/log/nginx/error.log` 链接到标准错误，可通过 `docker logs` 查看。

`SecAuditEngine` 支持以下模式：

| 模式 | 行为 |
|---|---|
| `Off` | 不写入 ModSecurity 审计日志 |
| `RelevantOnly` | 记录被规则标记或状态码符合 `SecAuditLogRelevantStatus` 的事务 |
| `On` | 记录全部事务 |

启用审计日志后，日志路径由 `SecAuditLog` 指定。容器被删除时，未挂载到持久存储的容器内日志文件会随容器文件系统一同删除。审计记录可能包含请求头、Cookie 和请求参数，具体内容由 `SecAuditLogParts` 控制。

## 验证 WAF

先验证正常请求：

```shell
curl -i http://127.0.0.1:8080/
```

SQL 注入：

```shell
curl -i --get \
    --data-urlencode "id=1' OR '1'='1" \
    http://127.0.0.1:8080/
```

在 `SecRuleEngine On`、CRS 默认阈值且没有相关排除规则时，该测试请求会命中 SQL 注入规则；达到入站异常阈值后，CRS 默认阻断动作返回 403。返回结果还会受到自定义规则、排除项和请求路由影响。

运行状态检查：

```shell
docker exec nginx-waf nginx -t
docker exec nginx-waf nginx -T
docker logs nginx-waf
```

启用 WAF 后，`nginx -T` 的输出中包含 `modsecurity on` 和 `modsecurity_rules_file`。如果测试请求未产生 ModSecurity 日志，可检查这两个指令是否位于实际接收请求的 `http`、`server` 或 `location` 上下文。

## 参考资料

- [ModSecurity](https://github.com/owasp-modsecurity/ModSecurity)
- [ModSecurity-nginx](https://github.com/owasp-modsecurity/ModSecurity-nginx)
- [OWASP Core Rule Set](https://github.com/coreruleset/coreruleset)
- [CRS 文档](https://coreruleset.org/docs/)
- [WordPress Rule Exclusions Plugin](https://github.com/coreruleset/wordpress-rule-exclusions-plugin)
