# syntax=docker/dockerfile:1

ARG ALPINE_IMAGE="alpine:3.24.1"

FROM ${ALPINE_IMAGE} AS builder

ARG NGINX_VERSION="1.31.3"

ENV NGINX_SOURCE_DIR="/usr/src/nginx-${NGINX_VERSION}" \
    DOCKER_NGINX_ENTRYPOINT_DIR="/usr/src/docker-nginx/entrypoint" \
    PKG_OSS_ALPINE_CONFIG_DIR="/usr/src/pkg-oss/alpine/alpine" \
    NGX_ACME_STATE_PREFIX="/var/cache/nginx" \
    RUSTFLAGS="-Ctarget-feature=-crt-static"

RUN apk add --no-cache autoconf automake brotli-dev \
    build-base cargo clang clang-dev curl curl-dev git libtool \
    linux-headers openssl-dev pcre2-dev pkgconfig wget zlib-dev zstd-dev zstd-static

WORKDIR /usr/src
RUN curl -fsSL "https://nginx.org/download/nginx-${NGINX_VERSION}.tar.gz" | tar xzf - -C /usr/src \
    && git clone --depth=1 https://github.com/nginx/docker-nginx.git \
    && git clone --depth=1 --branch "${NGINX_VERSION}-1" https://github.com/nginx/pkg-oss.git \
    && git clone --depth=1 https://github.com/openresty/headers-more-nginx-module.git "${NGINX_SOURCE_DIR}/ngx_headers_more" \
    && git clone --depth=1 https://github.com/tokers/zstd-nginx-module.git "${NGINX_SOURCE_DIR}/ngx_zstd" \
    && git clone --depth=1 https://github.com/nginx/nginx-acme.git "${NGINX_SOURCE_DIR}/ngx_http_acme_module" \
    && git clone --depth=1 --recurse-submodules --shallow-submodules --jobs="$(nproc)" https://github.com/google/ngx_brotli.git "${NGINX_SOURCE_DIR}/ngx_brotli"

WORKDIR ${NGINX_SOURCE_DIR}
RUN ./configure \
    --prefix=/etc/nginx \
    --sbin-path=/usr/sbin/nginx \
    --modules-path=/usr/lib/nginx/modules \
    --conf-path=/etc/nginx/nginx.conf \
    --error-log-path=/var/log/nginx/error.log \
    --http-log-path=/var/log/nginx/access.log \
    --pid-path=/run/nginx.pid \
    --lock-path=/run/nginx.lock \
    --http-client-body-temp-path=/var/cache/nginx/client_temp \
    --http-proxy-temp-path=/var/cache/nginx/proxy_temp \
    --http-fastcgi-temp-path=/var/cache/nginx/fastcgi_temp \
    --http-uwsgi-temp-path=/var/cache/nginx/uwsgi_temp \
    --http-scgi-temp-path=/var/cache/nginx/scgi_temp \
    --user=nginx \
    --group=nginx \
    --with-compat --with-file-aio --with-threads \
    --without-select_module --without-poll_module \
    --with-http_addition_module --with-http_auth_request_module \
    --with-http_gunzip_module --with-http_gzip_static_module \
    --with-http_random_index_module --with-http_realip_module \
    --with-http_secure_link_module --with-http_slice_module --with-http_ssl_module \
    --with-http_stub_status_module --with-http_sub_module --with-http_v2_module --with-http_v3_module \
    --with-stream --with-stream_realip_module --with-stream_ssl_module --with-stream_ssl_preread_module \
    --with-cc-opt='-Os -fstack-clash-protection -Wformat -Werror=format-security -fno-plt -g' \
    --with-ld-opt='-Wl,--as-needed,-O1,--sort-common -Wl,-z,pack-relative-relocs' \
    --add-module=ngx_headers_more \
    --add-module=ngx_zstd \
    --add-module=ngx_http_acme_module \
    --add-module=ngx_brotli \
    && make -j"$(nproc)" \
    && strip "${NGINX_SOURCE_DIR}/objs/nginx"

RUN mkdir -p /tmp/export/etc \
        /tmp/export/usr/share/nginx \
        /tmp/export/var/log/nginx \
        /tmp/export/usr/sbin \
        /tmp/export/var/cache/nginx \
        /tmp/export/usr/lib/nginx/modules \
    && mv "${NGINX_SOURCE_DIR}/objs/nginx" /tmp/export/usr/sbin/nginx \
    && mv "${NGINX_SOURCE_DIR}/html" /tmp/export/usr/share/nginx/ \
    && mv "${NGINX_SOURCE_DIR}/conf" /tmp/export/etc/nginx \
    && mkdir -p /tmp/export/etc/nginx/conf.d \
    && cp "${PKG_OSS_ALPINE_CONFIG_DIR}/nginx.conf" /tmp/export/etc/nginx/nginx.conf \
    && cp "${PKG_OSS_ALPINE_CONFIG_DIR}/nginx.default.conf" /tmp/export/etc/nginx/conf.d/default.conf \
    && rm -f /tmp/export/etc/nginx/koi-* \
        /tmp/export/etc/nginx/win-utf \
    && mv "${DOCKER_NGINX_ENTRYPOINT_DIR}/docker-entrypoint.sh" /tmp/export/ \
    && mv "${DOCKER_NGINX_ENTRYPOINT_DIR}" /tmp/export/docker-entrypoint.d

FROM ${ALPINE_IMAGE}

COPY --from=builder /tmp/export/ /
RUN set -x \
    && apk add --no-cache brotli-libs ca-certificates gettext-envsubst libcrypto3 libgcc libssl3 pcre2 tzdata zlib \
    && addgroup -g 101 -S nginx \
    && adduser -S -D -H -u 101 -h /var/cache/nginx -s /sbin/nologin -G nginx -g nginx nginx \
    && ln -sf /usr/lib/nginx/modules /etc/nginx/modules \
    && ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log

EXPOSE 80/tcp 443/tcp 443/udp
STOPSIGNAL SIGQUIT
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
