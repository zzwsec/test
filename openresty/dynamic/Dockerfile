# syntax=docker/dockerfile:1

# Dynamic modules layered on the OpenResty base image.

ARG ALPINE_IMAGE="alpine:3.24.1"
ARG OPENRESTY_IMAGE="zzwsec/openresty:1.31.1-alpine"
ARG RESTY_VERSION="1.31.1.1"
ARG NGINX_VERSION="1.31.1"

FROM ${ALPINE_IMAGE} AS modules-builder

ARG RESTY_VERSION
ARG NGINX_VERSION
ARG RESTY_J=""

ARG GEOIP2_REPO="https://github.com/leev/ngx_http_geoip2_module.git"
ARG GEOIP2_REF="3.4"
ARG OTEL_REPO="https://github.com/nginx/nginx-otel.git"
ARG OTEL_REF="v0.1.2"
ARG CACHE_PURGE_REPO="https://github.com/nginx-modules/ngx_cache_purge.git"
ARG CACHE_PURGE_REF="3.0.2"
ARG FANCYINDEX_REPO="https://github.com/aperezdc/ngx-fancyindex.git"
ARG FANCYINDEX_REF="v0.6.0"

# Keep the NGINX core feature set aligned with openresty/static.
ARG RESTY_CONFIG_OPTIONS="\
    --prefix=/usr/lib/openresty/nginx \
    --sbin-path=/usr/sbin/openresty \
    --modules-path=/usr/lib/openresty/modules \
    --conf-path=/etc/openresty/nginx.conf \
    --error-log-path=/var/log/openresty/error.log \
    --http-log-path=/var/log/openresty/access.log \
    --pid-path=/run/openresty/openresty.pid \
    --lock-path=/run/openresty/openresty.lock \
    --http-client-body-temp-path=/var/cache/openresty/client_temp \
    --http-proxy-temp-path=/var/cache/openresty/proxy_temp \
    --http-fastcgi-temp-path=/var/cache/openresty/fastcgi_temp \
    --http-uwsgi-temp-path=/var/cache/openresty/uwsgi_temp \
    --http-scgi-temp-path=/var/cache/openresty/scgi_temp \
    --with-compat --with-file-aio --with-threads \
    --without-select_module --without-poll_module \
    --with-http_addition_module --with-http_auth_request_module \
    --with-http_gunzip_module --with-http_gzip_static_module \
    --with-http_random_index_module --with-http_realip_module --with-http_secure_link_module \
    --with-http_slice_module --with-http_ssl_module --with-http_stub_status_module \
    --with-http_sub_module --with-http_v2_module --with-http_v3_module \
    --with-stream --with-stream_realip_module \
    --with-stream_ssl_module --with-stream_ssl_preread_module \
    --with-pcre-jit \
    "

ENV OPENRESTY_SOURCE_DIR="/usr/src/openresty-${RESTY_VERSION}" \
    NGINX_SOURCE_DIR="/usr/src/openresty-${RESTY_VERSION}/bundle/nginx-${NGINX_VERSION}"

RUN apk add --no-cache build-base c-ares-dev cmake curl git \
    linux-headers libmaxminddb-dev openssl-dev pcre2-dev zlib-dev

WORKDIR /usr/src
RUN curl -fsSL "https://openresty.org/download/openresty-${RESTY_VERSION}.tar.gz" | tar xzf - -C /usr/src \
    && git clone --depth=1 --branch "${GEOIP2_REF}" "${GEOIP2_REPO}" /usr/src/ngx_http_geoip2_module \
    && git clone --depth=1 --branch "${OTEL_REF}" "${OTEL_REPO}" /usr/src/nginx-otel \
    && git clone --depth=1 --branch "${CACHE_PURGE_REF}" "${CACHE_PURGE_REPO}" /usr/src/ngx_cache_purge \
    && git clone --depth=1 --branch "${FANCYINDEX_REF}" "${FANCYINDEX_REPO}" /usr/src/ngx-fancyindex

WORKDIR ${NGINX_SOURCE_DIR}
RUN ./configure \
        ${RESTY_CONFIG_OPTIONS} \
        --add-dynamic-module=/usr/src/ngx_http_geoip2_module \
        --add-dynamic-module=/usr/src/ngx_cache_purge \
        --add-dynamic-module=/usr/src/ngx-fancyindex \
    && MODULE_BUILD_JOBS="${RESTY_J:-$(nproc)}" \
    && make -j"${MODULE_BUILD_JOBS}" modules

# https://github.com/nginx/nginx-otel/blob/main/config
# https://cmake.org/cmake/help/latest/variable/CMAKE_POLICY_VERSION_MINIMUM.html
# _LARGEFILE64_SOURCE fixes Abseil's missing off64_t declaration on Alpine/musl.
RUN cmake \
        -S /usr/src/nginx-otel \
        -B /usr/src/nginx-otel/build \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_CXX_FLAGS=-D_LARGEFILE64_SOURCE \
        -DCMAKE_POLICY_VERSION_MINIMUM=3.5 \
        -DNGX_OTEL_NGINX_BUILD_DIR="${PWD}/objs"

RUN MODULE_BUILD_JOBS="${RESTY_J:-$(nproc)}" \
    && cmake --build /usr/src/nginx-otel/build \
        --config Release \
        --parallel "${MODULE_BUILD_JOBS}" \
    && mkdir -p /tmp/export/usr/lib/openresty/modules \
    && cp objs/*.so /tmp/export/usr/lib/openresty/modules/ \
    && cp /usr/src/nginx-otel/build/ngx_otel_module.so /tmp/export/usr/lib/openresty/modules/ \
    && strip --strip-unneeded /tmp/export/usr/lib/openresty/modules/*.so

FROM ${OPENRESTY_IMAGE}

RUN apk add --no-cache c-ares libmaxminddb-libs libstdc++

COPY --from=modules-builder /tmp/export/ /
