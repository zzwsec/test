# OpenResty Alpine 镜像

使用 Alpine 多阶段构建 OpenResty，并单独编译 OpenSSL 3 和 PCRE2。运行时镜像包含 OpenResty 及其运行依赖。

## 构建

在仓库根目录执行：

```shell
docker build \
    --tag zzwsec/openresty:1.31.1-alpine \
    openresty/static
```

主要构建参数：

| 参数 | 默认值 | 说明 |
|---|---|---|
| `ALPINE_IMAGE` | `alpine:3.24.1` | 构建阶段和运行阶段的 Alpine 基础镜像 |
| `RESTY_VERSION` | `1.31.1.1` | OpenResty 源码版本 |
| `RESTY_OPENSSL_VERSION` | `3.5.7` | OpenSSL 源码版本 |
| `RESTY_OPENSSL_PATCH_VERSION` | `3.5.5` | OpenResty OpenSSL 补丁版本 |
| `RESTY_OPENSSL_URL_BASE` | OpenSSL GitHub Release 地址 | OpenSSL 源码下载地址前缀 |
| `RESTY_PCRE_VERSION` | `10.47` | PCRE2 源码版本 |
| `RESTY_USER` | `openresty` | worker 进程使用的系统账号名称 |
| `RESTY_UID` | `101` | `openresty` 系统账号 UID |
| `RESTY_GID` | `101` | `openresty` 系统组 GID |
| `RESTY_J` | 空 | 并行编译任务数；空值使用 `nproc` |
| `DOCKER_OPENRESTY_REPO` | `https://github.com/openresty/docker-openresty.git` | OpenResty 容器配置仓库 |
| `NGX_BROTLI_REPO` | `https://github.com/google/ngx_brotli.git` | Brotli 模块源码仓库 |
| `NGX_ZSTD_REPO` | `https://github.com/tokers/zstd-nginx-module.git` | Zstandard 模块源码仓库 |
| `NGX_ACME_REPO` | `https://github.com/nginx/nginx-acme.git` | ACME 模块源码仓库 |

可覆盖的编译参数：

| 参数 | 用途 |
|---|---|
| `RESTY_OPENSSL_BUILD_OPTIONS` | 传递给 OpenSSL `config` 的功能参数 |
| `RESTY_PCRE_BUILD_OPTIONS` | 传递给 PCRE2 `configure` 的功能参数 |
| `RESTY_CONFIG_OPTIONS` | OpenResty/NGINX 内置模块和编译功能参数 |
| `RESTY_CONFIG_OPTIONS_MORE` | 第三方静态模块参数 |
| `RESTY_LUAJIT_OPTIONS` | LuaJIT 编译参数 |
| `RESTY_PCRE_OPTIONS` | OpenResty 使用 PCRE2 时的参数 |

上述参数通过 `--build-arg` 传入时会替换对应的默认值，不会与默认值合并。

指定并行编译任务数：

```shell
docker build \
    --build-arg RESTY_J=8 \
    --tag zzwsec/openresty:1.31.1-alpine \
    openresty/static
```

检查镜像：

```shell
docker run --rm zzwsec/openresty:1.31.1-alpine openresty -V
docker run --rm zzwsec/openresty:1.31.1-alpine openresty -t
```

## 目录结构

镜像按照系统软件目录分工组织文件：

| 路径 | 内容 |
|---|---|
| `/usr/sbin/openresty` | OpenResty 服务程序 |
| `/usr/sbin/nginx` | 指向 `openresty` 的兼容命令链接 |
| `/usr/bin/openresty` | 指向服务程序的命令链接 |
| `/usr/bin/luajit` | LuaJIT 命令链接 |
| `/etc/openresty/nginx.conf` | 主配置文件 |
| `/etc/openresty/conf.d/` | HTTP 和 stream 扩展配置 |
| `/usr/lib/openresty/` | LuaJIT、Lua 模块、OpenSSL 和 PCRE2 私有运行库 |
| `/usr/share/openresty/html/` | 默认静态文件 |
| `/run/openresty/` | PID 和锁文件 |
| `/var/log/openresty/` | 访问日志和错误日志；容器内链接至标准输出和标准错误 |
| `/var/cache/openresty/` | 临时请求体、代理、FastCGI、uWSGI、SCGI 和 ACME 状态数据 |

镜像设置的 `LUA_PATH` 和 `LUA_CPATH` 包含 `/usr/lib/openresty/lualib`，通过 `/usr/bin/luajit` 执行脚本时可以加载镜像内置的 Lua 模块。

镜像预先创建 `/var/cache/openresty` 父目录。`client_temp`、`proxy_temp`、`fastcgi_temp`、`uwsgi_temp` 和 `scgi_temp` 子目录由 OpenResty 在读取配置并启动时创建。

`/usr/sbin/nginx` 链接到 `/usr/sbin/openresty`，用于兼容执行 `nginx -t`、`nginx -s reload` 的管理脚本。镜像仍以 `openresty` 作为启动命令。

`/etc/openresty` 保留运行配置使用的 `mime.types`、`fastcgi.conf`、`fastcgi_params`、`scgi_params` 和 `uwsgi_params`。构建时删除上游安装生成的 `*.default` 配置备份和字符集映射示例文件。

## 运行账号

容器启动时，OpenResty master 进程使用 root 账号绑定 80 和 443 端口，worker 进程根据 `/etc/openresty/nginx.conf` 中的 `user` 指令切换为 `openresty` 账号。该账号不创建主目录，登录 shell 为 `/sbin/nologin`。

以下可写目录归 `openresty:openresty` 所有：

- `/run/openresty`
- `/var/cache/openresty`
- `/var/log/openresty`

挂载自定义配置：

```shell
docker run --rm \
    --publish 8080:80 \
    --volume ./nginx.conf:/etc/openresty/nginx.conf:ro \
    --volume ./conf.d:/etc/openresty/conf.d:ro \
    zzwsec/openresty:1.31.1-alpine
```

## 静态模块

以下第三方模块静态链接至 NGINX 主程序：

- `ngx_brotli`：Brotli 压缩
- `zstd-nginx-module`：Zstandard 压缩
- `nginx-acme`：ACME 证书管理
