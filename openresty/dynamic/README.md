# OpenResty 动态模块镜像

该镜像为指定版本的 OpenResty/NGINX 编译动态模块，并将运行依赖和模块文件复制到 OpenResty 基础镜像中。

## 模块

| 模块 | 版本 | 文件 | 用途 |
|---|---:|---|---|
| GeoIP2 | 3.4 | `ngx_http_geoip2_module.so` | HTTP IP 地理信息查询 |
| GeoIP2 | 3.4 | `ngx_stream_geoip2_module.so` | Stream IP 地理信息查询 |
| OpenTelemetry | 0.1.2 | `ngx_otel_module.so` | W3C Trace Context 与 OTLP/gRPC 分布式追踪 |
| cache-purge | 3.0.2 | `ngx_http_cache_purge_module.so` | NGINX 缓存清理 |
| fancyindex | 0.6.0 | `ngx_http_fancyindex_module.so` | 可定制目录索引 |

模块安装目录：

```text
/usr/lib/openresty/modules/
```

## 构建

在仓库根目录执行：

```shell
docker build \
    --build-arg OPENRESTY_IMAGE=zzwsec/openresty:1.31.1-alpine \
    --tag zzwsec/openresty:1.31.1-alpine-dynamic \
    openresty/dynamic
```

默认构建结果以 `OPENRESTY_IMAGE` 为运行阶段基础镜像，并包含全部动态模块及其运行依赖。

主要构建参数：

| 参数 | 默认值 | 说明 |
|---|---|---|
| `ALPINE_IMAGE` | `alpine:3.24.1` | 动态模块构建阶段的 Alpine 基础镜像 |
| `OPENRESTY_IMAGE` | `zzwsec/openresty:1.31.1-alpine` | 最终运行阶段的基础镜像 |
| `RESTY_VERSION` | `1.31.1.1` | OpenResty 源码版本 |
| `NGINX_VERSION` | `1.31.1` | OpenResty bundle 中的 NGINX 源码版本 |
| `RESTY_J` | 空 | 并行编译任务数；空值使用 `nproc` |

## 启用模块

`load_module` 必须位于 `nginx.conf` 主配置上下文，并置于 `events`、`http` 和 `stream` 配置块之前。模块可按需加载，无须全部启用。

```nginx
load_module /usr/lib/openresty/modules/ngx_http_geoip2_module.so;
load_module /usr/lib/openresty/modules/ngx_stream_geoip2_module.so;
load_module /usr/lib/openresty/modules/ngx_otel_module.so;
load_module /usr/lib/openresty/modules/ngx_http_cache_purge_module.so;
load_module /usr/lib/openresty/modules/ngx_http_fancyindex_module.so;
```

检查配置：

```shell
openresty -t
```

## GeoIP2

GeoIP2 模块不包含 MaxMind 数据库。数据库文件须单独下载或挂载，并保证 OpenResty 工作进程具有读取权限。数据库路径可自定义，`/usr/share/GeoIP/` 仅作为常用目录。

HTTP 配置示例：

```nginx
http {
    geoip2 /usr/share/GeoIP/GeoLite2-Country.mmdb {
        $geoip2_country_code country iso_code;
        $geoip2_country_name  country names zh-CN;
    }

    server {
        location / {
            add_header X-Country-Code $geoip2_country_code always;
            proxy_pass http://backend;
        }
    }
}
```

Stream 配置示例：

```nginx
stream {
    geoip2 /usr/share/GeoIP/GeoLite2-Country.mmdb {
        $geoip2_stream_country_code country iso_code;
    }

    log_format geoip '$remote_addr $geoip2_stream_country_code';
    access_log /var/log/openresty/stream-access.log geoip;
}
```

## OpenTelemetry

OpenTelemetry 模块默认不发送追踪数据。启用追踪时须配置支持 OTLP/gRPC 的 Collector。

```nginx
http {
    otel_exporter {
        endpoint otel-collector:4317;
    }

    otel_service_name openresty;

    server {
        location / {
            otel_trace on;
            otel_trace_context propagate;
            proxy_pass http://backend;
        }
    }
}
```

## cache-purge

缓存键和清理键必须保持一致：

```nginx
http {
    proxy_cache_path /var/cache/openresty/proxy_cache
        levels=1:2
        keys_zone=content_cache:10m;

    server {
        location / {
            proxy_cache content_cache;
            proxy_cache_key "$scheme$host$request_uri";
            proxy_pass http://backend;
        }

        location ~ ^/purge(/.*)$ {
            allow 127.0.0.1;
            deny all;
            proxy_cache_purge content_cache "$scheme$host$1$is_args$args";
        }
    }
}
```

缓存清理接口不得直接暴露至公网。访问范围须通过 `allow`、`deny`、认证或内部网络进行限制。

## fancyindex

```nginx
http {
    server {
        location /downloads/ {
            alias /srv/downloads/;
            fancyindex on;
        }
    }
}
```

## 版本兼容

动态模块必须与运行时 OpenResty/NGINX ABI 兼容。`--with-compat` 不能替代 NGINX 版本匹配。

升级基础镜像后，必须同步更新 `RESTY_VERSION` 和 `NGINX_VERSION`，重新构建全部动态模块，并执行以下检查：

```shell
openresty -V
openresty -t
```

旧版本生成的 `.so` 文件不得直接复用于不同 NGINX 版本。
