# syntax=docker/dockerfile:1

# Dynamic modules layered on the OpenResty base image.

ARG ALPINE_IMAGE="alpine:3.24.1"
ARG OPENRESTY_IMAGE="test-openesty-static:latest"
ARG RESTY_VERSION="1.31.1.1"
ARG RESTY_OPENSSL_VERSION="3.5.7"
ARG RESTY_PCRE_VERSION="10.47"

FROM ${ALPINE_IMAGE} AS modules-builder

ARG RESTY_VERSION
ARG RESTY_OPENSSL_VERSION
ARG RESTY_PCRE_VERSION
ARG RESTY_J=""

ARG GEOIP2_REPO="https://github.com/leev/ngx_http_geoip2_module.git"
ARG GEOIP2_REF="3.4"
ARG OTEL_REPO="https://github.com/nginx/nginx-otel.git"
ARG OTEL_REF="v0.1.2"
ARG CACHE_PURGE_REPO="https://github.com/nginx-modules/ngx_cache_purge.git"
ARG CACHE_PURGE_REF="3.0.2"
ARG FANCYINDEX_REPO="https://github.com/aperezdc/ngx-fancyindex.git"
ARG FANCYINDEX_REF="v0.6.0"

# Keep the NGINX core feature set aligned with the static image.
# Unlike OpenResty's top-level configure, NGINX's configure does not enable stream automatically.
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
    --with-compat \
    --with-file-aio \
    --with-threads \
    --without-select_module \
    --without-poll_module \
    --with-http_addition_module \
    --with-http_auth_request_module \
    --with-http_gunzip_module \
    --with-http_gzip_static_module \
    --with-http_random_index_module \
    --with-http_realip_module \
    --with-http_secure_link_module \
    --with-http_slice_module \
    --with-http_ssl_module \
    --with-http_stub_status_module \
    --with-http_sub_module \
    --with-http_v2_module \
    --with-http_v3_module \
    --with-stream \
    --with-stream_realip_module \
    --with-stream_ssl_module \
    --with-stream_ssl_preread_module \
    "

ARG RESTY_PCRE_OPTIONS="--with-pcre-jit"
ARG _RESTY_CC_OPTIONS="\
    -Os \
    -fstack-clash-protection \
    -Wformat \
    -Werror=format-security \
    -fno-plt \
    -g \
    -DNGX_LUA_ABORT_AT_PANIC \
    -I/usr/lib/openresty/pcre2/include \
    -I/usr/lib/openresty/openssl3/include \
    "
ARG _RESTY_LD_OPTIONS="\
    -L/usr/lib/openresty/pcre2/lib \
    -L/usr/lib/openresty/openssl3/lib \
    -Wl,--as-needed,-O1,--sort-common \
    -Wl,-z,pack-relative-relocs \
    -Wl,-rpath,/usr/lib/openresty/pcre2/lib:/usr/lib/openresty/openssl3/lib \
    "
ARG _RESTY_CONFIG_DEPS="\
    --with-cc-opt='${_RESTY_CC_OPTIONS}' \
    --with-ld-opt='${_RESTY_LD_OPTIONS}' \
    --with-pcre \
    "

ENV OPENRESTY_SOURCE_DIR="/usr/src/openresty-${RESTY_VERSION}" \
    NGINX_SOURCE_DIR="/usr/src/openresty-nginx-source" \
    OPENSSL_SOURCE_DIR="/usr/src/openssl-${RESTY_OPENSSL_VERSION}" \
    PCRE_SOURCE_DIR="/usr/src/pcre2-${RESTY_PCRE_VERSION}"

RUN apk add --no-cache build-base c-ares-dev cmake curl git \
    libmaxminddb-dev linux-headers perl-dev zlib-dev

WORKDIR /usr/src
# OpenResty bundles exactly one NGINX source tree. Use a stable link so its
# version does not need to be derived from RESTY_VERSION or passed separately.
RUN curl -fsSL "https://github.com/openssl/openssl/releases/download/openssl-${RESTY_OPENSSL_VERSION}/openssl-${RESTY_OPENSSL_VERSION}.tar.gz" | tar xzf - -C /usr/src \
    && curl -fsSL "https://github.com/PCRE2Project/pcre2/releases/download/pcre2-${RESTY_PCRE_VERSION}/pcre2-${RESTY_PCRE_VERSION}.tar.gz" | tar xzf - -C /usr/src \
    && curl -fsSL "https://openresty.org/download/openresty-${RESTY_VERSION}.tar.gz" | tar xzf - -C /usr/src \
    && nginx_source_dir="$(find "${OPENRESTY_SOURCE_DIR}/bundle" -maxdepth 1 -type d -name 'nginx-*')" \
    && [ -n "${nginx_source_dir}" ] \
    && [ "$(printf '%s\n' "${nginx_source_dir}" | wc -l)" -eq 1 ] \
    && ln -s "${nginx_source_dir}" "${NGINX_SOURCE_DIR}" \
    && git clone --depth=1 --branch "${GEOIP2_REF}" "${GEOIP2_REPO}" /usr/src/ngx_http_geoip2_module \
    && git clone --depth=1 --branch "${OTEL_REF}" "${OTEL_REPO}" /usr/src/nginx-otel \
    && git clone --depth=1 --branch "${CACHE_PURGE_REF}" "${CACHE_PURGE_REPO}" /usr/src/ngx_cache_purge \
    && git clone --depth=1 --branch "${FANCYINDEX_REF}" "${FANCYINDEX_REPO}" /usr/src/ngx-fancyindex

# ngx_otel_module.so uses OpenSSL through its gRPC/OTLP dependencies, so the
# modules builder must compile the same private OpenSSL version as the static image.
# PCRE2 is also kept aligned with the static image for the NGINX module build.
WORKDIR ${OPENSSL_SOURCE_DIR}
RUN MODULE_BUILD_JOBS="${RESTY_J:-$(nproc)}" \
    && ./config \
        shared zlib -g \
        --prefix=/usr/lib/openresty/openssl3 \
        --openssldir=/etc/ssl \
        --libdir=lib \
    && make -j"${MODULE_BUILD_JOBS}" \
    && make -j"${MODULE_BUILD_JOBS}" install_sw

WORKDIR ${PCRE_SOURCE_DIR}
RUN MODULE_BUILD_JOBS="${RESTY_J:-$(nproc)}" \
    && CFLAGS="-g -O3" ./configure \
        --prefix=/usr/lib/openresty/pcre2 \
        --libdir=/usr/lib/openresty/pcre2/lib \
        --disable-static \
        --enable-jit \
        --enable-shared \
        --enable-unicode \
    && make -j"${MODULE_BUILD_JOBS}" \
    && make -j"${MODULE_BUILD_JOBS}" install

WORKDIR ${NGINX_SOURCE_DIR}
RUN MODULE_BUILD_JOBS="${RESTY_J:-$(nproc)}" \
    && eval ./configure \
        ${_RESTY_CONFIG_DEPS} \
        ${RESTY_CONFIG_OPTIONS} \
        ${RESTY_PCRE_OPTIONS} \
        --add-dynamic-module=/usr/src/ngx_http_geoip2_module \
        --add-dynamic-module=/usr/src/ngx_cache_purge \
        --add-dynamic-module=/usr/src/ngx-fancyindex \
    && make -j"${MODULE_BUILD_JOBS}" modules

# https://github.com/nginx/nginx-otel/blob/main/config
# https://cmake.org/cmake/help/latest/variable/CMAKE_POLICY_VERSION_MINIMUM.html
# _LARGEFILE64_SOURCE fixes Abseil's missing off64_t declaration on Alpine/musl.
RUN cmake \
        -S /usr/src/nginx-otel \
        -B /usr/src/nginx-otel/build \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_BUILD_RPATH=/usr/lib/openresty/openssl3/lib \
        -DCMAKE_C_FLAGS=-I/usr/lib/openresty/pcre2/include \
        -DCMAKE_CXX_FLAGS="-D_LARGEFILE64_SOURCE -I/usr/lib/openresty/pcre2/include" \
        -DCMAKE_POLICY_VERSION_MINIMUM=3.5 \
        -DNGX_OTEL_NGINX_BUILD_DIR="${PWD}/objs" \
        -DOPENSSL_ROOT_DIR=/usr/lib/openresty/openssl3 \
        -DOPENSSL_USE_STATIC_LIBS=FALSE

RUN MODULE_BUILD_JOBS="${RESTY_J:-$(nproc)}" \
    && cmake --build /usr/src/nginx-otel/build \
        --config Release \
        --parallel "${MODULE_BUILD_JOBS}" \
    && mkdir -p /usr/lib/openresty/modules \
    && cp objs/*.so /usr/lib/openresty/modules/ \
    && cp /usr/src/nginx-otel/build/ngx_otel_module.so /usr/lib/openresty/modules/ \
    && strip --strip-unneeded /usr/lib/openresty/modules/*.so

FROM ${OPENRESTY_IMAGE}

COPY --from=modules-builder /usr/lib/openresty/modules/ /usr/lib/openresty/modules/

# zlib and libgcc are already provided by the static base image.
# Install only the additional shared libraries required by GeoIP2 and OpenTelemetry.
RUN apk add --no-cache c-ares libmaxminddb-libs libstdc++
