# syntax=docker/dockerfile:1

ARG ALPINE_IMAGE="alpine:3.24.1"
ARG RESTY_VERSION="1.31.1.1"
ARG RESTY_OPENSSL_VERSION="3.5.7"
ARG RESTY_OPENSSL_PATCH_VERSION="3.5.5"
ARG RESTY_PCRE_VERSION="10.47"
ARG RESTY_LUAROCKS_VERSION="3.13.0"
ARG RESTY_USER="openresty"
ARG RESTY_UID="101"
ARG RESTY_GID="101"

FROM ${ALPINE_IMAGE} AS builder

ARG RESTY_VERSION
ARG RESTY_OPENSSL_VERSION
ARG RESTY_OPENSSL_PATCH_VERSION
ARG RESTY_PCRE_VERSION
ARG RESTY_LUAROCKS_VERSION
ARG RESTY_USER
ARG RESTY_J=""
ARG DOCKER_OPENRESTY_REPO="https://github.com/openresty/docker-openresty.git"
ARG NGX_BROTLI_REPO="https://github.com/google/ngx_brotli.git"
ARG NGX_ZSTD_REPO="https://github.com/tokers/zstd-nginx-module.git"
ARG NGX_ACME_REPO="https://github.com/nginx/nginx-acme.git"

# https://github.com/openresty/openresty-packaging/blob/master/alpine/openresty-openssl3/APKBUILD
ARG RESTY_OPENSSL_URL_BASE="https://github.com/openssl/openssl/releases/download/openssl-${RESTY_OPENSSL_VERSION}"
ARG RESTY_OPENSSL_BUILD_OPTIONS="\
    enable-camellia enable-cms enable-fips enable-ktls enable-md2 enable-rc5 \
    enable-rfc3779 enable-seed enable-ssl3 enable-ssl3-method \
    enable-weak-ssl-ciphers \
    "

# https://github.com/openresty/openresty-packaging/blob/master/alpine/openresty-pcre2/APKBUILD
ARG RESTY_PCRE_BUILD_OPTIONS="\
    --disable-bsr-anycrlf --disable-coverage --disable-ebcdic --disable-fuzz-support \
    --disable-jit-sealloc --disable-never-backslash-C \
    --disable-pcre2grep-libbz2 --disable-pcre2grep-libz --disable-pcre2test-libedit \
    --disable-rebuild-chartables --disable-silent-rules --disable-static --disable-valgrind \
    --enable-jit --enable-newline-is-lf \
    --enable-pcre2-8 --enable-pcre2-16 --enable-pcre2-32 \
    --enable-pcre2grep-callout --enable-pcre2grep-callout-fork --enable-pcre2grep-jit \
    --enable-percent-zt --enable-shared --enable-unicode \
    "

# https://github.com/openresty/openresty-packaging/blob/master/alpine/openresty/APKBUILD
ARG RESTY_CONFIG_OPTIONS="\
    --prefix=/usr/lib/openresty \
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
    --without-http_rds_csv_module --without-http_rds_json_module --without-lua_rds_parser \
    "
ARG RESTY_CONFIG_OPTIONS_MORE="\
    --add-module=ngx_brotli \
    --add-module=ngx_zstd \
    --add-module=ngx_http_acme_module \
    "
ARG RESTY_LUAJIT_OPTIONS="--with-luajit-xcflags='-DLUAJIT_NUMMODE=2 -DLUAJIT_ENABLE_LUA52COMPAT'"
ARG RESTY_PCRE_OPTIONS="--with-pcre-jit"
ARG _RESTY_CONFIG_DEPS="\
    --with-cc-opt='-Os -fstack-clash-protection -Wformat -Werror=format-security -fno-plt -g -DNGX_LUA_ABORT_AT_PANIC -I/usr/lib/openresty/pcre2/include -I/usr/lib/openresty/openssl3/include' \
    --with-ld-opt='-L/usr/lib/openresty/pcre2/lib -L/usr/lib/openresty/openssl3/lib -Wl,--as-needed,-O1,--sort-common -Wl,-z,pack-relative-relocs -Wl,-rpath,/usr/lib/openresty/pcre2/lib:/usr/lib/openresty/openssl3/lib' \
    --with-pcre \
    "

ENV DOCKER_OPENRESTY_SOURCE_DIR="/usr/src/docker-openresty" \
    OPENRESTY_SOURCE_DIR="/usr/src/openresty-${RESTY_VERSION}" \
    OPENSSL_SOURCE_DIR="/usr/src/openssl-${RESTY_OPENSSL_VERSION}" \
    PCRE_SOURCE_DIR="/usr/src/pcre2-${RESTY_PCRE_VERSION}" \
    LUAROCKS_SOURCE_DIR="/usr/src/luarocks-${RESTY_LUAROCKS_VERSION}" \
    NGX_ACME_STATE_PREFIX="/var/cache/openresty" \
    RUSTFLAGS="-Ctarget-feature=-crt-static"

RUN apk add --no-cache binutils build-base cargo clang clang-dev \
    cmake coreutils curl git linux-headers make mariadb-connector-c-dev \
    perl-dev pkgconf readline-dev unzip wget zlib-dev zstd-dev zstd-static

WORKDIR /usr/src
RUN curl -fsSL "${RESTY_OPENSSL_URL_BASE}/openssl-${RESTY_OPENSSL_VERSION}.tar.gz" | tar xzf - -C /usr/src \
    && curl -fsSL "https://raw.githubusercontent.com/openresty/openresty/master/patches/openssl-${RESTY_OPENSSL_PATCH_VERSION}-sess_set_get_cb_yield.patch" | patch -d "${OPENSSL_SOURCE_DIR}" -p1 \
    && curl -fsSL "https://github.com/PCRE2Project/pcre2/releases/download/pcre2-${RESTY_PCRE_VERSION}/pcre2-${RESTY_PCRE_VERSION}.tar.gz" | tar xzf - -C /usr/src \
    && curl -fsSL "https://openresty.org/download/openresty-${RESTY_VERSION}.tar.gz" | tar xzf - -C /usr/src \
    && curl -fsSL "https://luarocks.org/releases/luarocks-${RESTY_LUAROCKS_VERSION}.tar.gz" | tar xzf - -C /usr/src \
    && git clone --depth=1 --branch "${RESTY_VERSION}-0" "${DOCKER_OPENRESTY_REPO}" "${DOCKER_OPENRESTY_SOURCE_DIR}" \
    && git clone --depth=1 --recurse-submodules --shallow-submodules --jobs="$(nproc)" "${NGX_BROTLI_REPO}" "${OPENRESTY_SOURCE_DIR}/ngx_brotli" \
    && git clone --depth=1 "${NGX_ZSTD_REPO}" "${OPENRESTY_SOURCE_DIR}/ngx_zstd" \
    && git clone --depth=1 "${NGX_ACME_REPO}" "${OPENRESTY_SOURCE_DIR}/ngx_http_acme_module"

# https://github.com/google/ngx_brotli#statically-compiled
WORKDIR ${OPENRESTY_SOURCE_DIR}
RUN RESTY_BUILD_JOBS="${RESTY_J:-$(nproc)}" \
    && cmake \
        -S ngx_brotli/deps/brotli \
        -B ngx_brotli/deps/brotli/out \
        -DCMAKE_BUILD_TYPE=Release \
        -DBUILD_SHARED_LIBS=OFF \
        -DBROTLI_DISABLE_TESTS=ON \
    && cmake --build ngx_brotli/deps/brotli/out \
        --config Release \
        --target brotlienc \
        --parallel "${RESTY_BUILD_JOBS}"

WORKDIR ${OPENSSL_SOURCE_DIR}
RUN RESTY_BUILD_JOBS="${RESTY_J:-$(nproc)}" \
    && ./config \
        shared zlib -g \
        --prefix=/usr/lib/openresty/openssl3 \
        --openssldir=/etc/ssl \
        --libdir=lib \
        -Wl,-rpath,/usr/lib/openresty/openssl3/lib \
        ${RESTY_OPENSSL_BUILD_OPTIONS} \
    && make -j"${RESTY_BUILD_JOBS}" \
    && make -j"${RESTY_BUILD_JOBS}" install_sw

WORKDIR ${PCRE_SOURCE_DIR}
RUN RESTY_BUILD_JOBS="${RESTY_J:-$(nproc)}" \
    && CFLAGS="-g -O3" ./configure \
        --prefix=/usr/lib/openresty/pcre2 \
        --libdir=/usr/lib/openresty/pcre2/lib \
        ${RESTY_PCRE_BUILD_OPTIONS} \
    && make -j"${RESTY_BUILD_JOBS}" \
    && make -j"${RESTY_BUILD_JOBS}" install

WORKDIR ${OPENRESTY_SOURCE_DIR}
RUN RESTY_BUILD_JOBS="${RESTY_J:-$(nproc)}" \
    && export \
        CARGO_BUILD_JOBS="${RESTY_BUILD_JOBS}" \
        OPENSSL_DIR=/usr/lib/openresty/openssl3 \
        OPENSSL_INCLUDE_DIR=/usr/lib/openresty/openssl3/include \
        OPENSSL_LIB_DIR=/usr/lib/openresty/openssl3/lib \
        PKG_CONFIG_PATH=/usr/lib/openresty/openssl3/lib/pkgconfig \
        ZSTD_INC=/usr/include \
        ZSTD_LIB=/usr/lib \
    && eval ./configure -j"${RESTY_BUILD_JOBS}" ${_RESTY_CONFIG_DEPS} ${RESTY_CONFIG_OPTIONS} ${RESTY_CONFIG_OPTIONS_MORE} ${RESTY_LUAJIT_OPTIONS} ${RESTY_PCRE_OPTIONS} \
    && sed -i \
        -e "s|--add-module=${PWD}/ngx_brotli|--add-module=ngx_brotli|g" \
        -e "s|--add-module=${PWD}/ngx_zstd|--add-module=ngx_zstd|g" \
        -e "s|--add-module=${PWD}/ngx_http_acme_module|--add-module=ngx_http_acme_module|g" \
        -e "s|--add-module=\.\./|--add-module=|g" \
        build/nginx-*/objs/ngx_auto_config.h \
    && make -j"${RESTY_BUILD_JOBS}" \
    && make -j"${RESTY_BUILD_JOBS}" install

WORKDIR ${LUAROCKS_SOURCE_DIR}
RUN ./configure \
        --prefix=/usr/local/luarocks \
        --rocks-tree=/usr/lib/openresty/luajit \
        --with-lua=/usr/lib/openresty/luajit \
        --with-lua-include=/usr/lib/openresty/luajit/include/luajit-2.1 \
    && make \
    && make install

RUN export PATH="${PATH}:/usr/local/luarocks/bin" \
    && CFLAGS="-O2 -fPIC" \
        LDFLAGS="-L/usr/lib/openresty/openssl3/lib -Wl,-rpath,/usr/lib/openresty/openssl3/lib" \
        luarocks install luaossl \
        CRYPTO_INCDIR=/usr/lib/openresty/openssl3/include \
        CRYPTO_LIBDIR=/usr/lib/openresty/openssl3/lib \
        OPENSSL_INCDIR=/usr/lib/openresty/openssl3/include \
        OPENSSL_LIBDIR=/usr/lib/openresty/openssl3/lib \
    && luarocks install lapis \
    && luarocks install lua-resty-mlcache \
    && luarocks install lua-resty-jit-uuid \
    && luarocks install luasql-mysql MYSQL_INCDIR=/usr/include/mysql

# Remove build artifacts after all native rocks have been compiled.
RUN rm -rf \
        /usr/lib/openresty/openssl3/bin/c_rehash \
        /usr/lib/openresty/openssl3/lib/*.a \
        /usr/lib/openresty/openssl3/include \
    && find /usr/lib/openresty/openssl3 -type f -perm -u+x -exec strip --strip-unneeded '{}' \; \
    && rm -rf \
        /usr/lib/openresty/pcre2/bin \
        /usr/lib/openresty/pcre2/include \
        /usr/lib/openresty/pcre2/share \
        /usr/lib/openresty/pcre2/lib/*.a \
        /usr/lib/openresty/pcre2/lib/*.la \
    && find /usr/lib/openresty/pcre2 -type f \( -name '*.so' -o -name '*.so.*' \) -exec strip --strip-unneeded '{}' \; \
    && rm -rf \
        /usr/lib/openresty/luajit/lib/*.a \
        /usr/lib/openresty/luajit/include \
        /usr/lib/openresty/luajit/share/man \
    && find /usr/lib/openresty/luajit -type f \( -name '*.so' -o -name '*.so.*' \) -exec strip --strip-unneeded '{}' \; \
    && strip --strip-unneeded /usr/lib/openresty/luajit/bin/luajit \
    && strip --strip-unneeded /usr/sbin/openresty

RUN mkdir -p \
        /tmp/export/etc/openresty/conf.d \
        /tmp/export/usr/bin \
        /tmp/export/usr/lib \
        /tmp/export/usr/sbin \
        /tmp/export/usr/share/openresty \
        /tmp/export/var/cache/openresty \
        /tmp/export/var/log/openresty \
        /tmp/export/run/openresty \
    && mv /usr/lib/openresty /tmp/export/usr/lib/ \
    && mv /usr/sbin/openresty /tmp/export/usr/sbin/ \
    && mv /etc/openresty/* /tmp/export/etc/openresty/ \
    && mv /tmp/export/usr/lib/openresty/nginx/html /tmp/export/usr/share/openresty/ \
    && rm -rf \
        /tmp/export/usr/lib/openresty/bin \
        /tmp/export/usr/lib/openresty/nginx \
        /tmp/export/usr/lib/openresty/pod \
        /tmp/export/usr/lib/openresty/resty.index \
    && rm -f \
        /tmp/export/etc/openresty/*.default \
        /tmp/export/etc/openresty/koi-* \
        /tmp/export/etc/openresty/win-utf \
    && ln -s ../sbin/openresty /tmp/export/usr/bin/openresty \
    && ln -s ../lib/openresty/openssl3/bin/openssl /tmp/export/usr/bin/openssl \
    && ln -s openresty /tmp/export/usr/sbin/nginx \
    && sed \
        -e "s|#user  nobody;|user ${RESTY_USER};|g" \
        -e 's|/usr/local/openresty/nginx/conf/nginx.conf|/etc/openresty/nginx.conf|g' \
        -e 's|/etc/nginx|/etc/openresty|g' \
        -e 's|/var/run/openresty/nginx-client-body|/var/cache/openresty/client_temp|g' \
        -e 's|/var/run/openresty/nginx-proxy|/var/cache/openresty/proxy_temp|g' \
        -e 's|/var/run/openresty/nginx-fastcgi|/var/cache/openresty/fastcgi_temp|g' \
        -e 's|/var/run/openresty/nginx-uwsgi|/var/cache/openresty/uwsgi_temp|g' \
        -e 's|/var/run/openresty/nginx-scgi|/var/cache/openresty/scgi_temp|g' \
        "${DOCKER_OPENRESTY_SOURCE_DIR}/nginx.conf" \
        > /tmp/export/etc/openresty/nginx.conf \
    && sed \
        -e 's|/etc/nginx|/etc/openresty|g' \
        -e 's|/usr/local/openresty/nginx/html|/usr/share/openresty/html|g' \
        "${DOCKER_OPENRESTY_SOURCE_DIR}/nginx.vh.default.conf" \
        > /tmp/export/etc/openresty/conf.d/default.conf

FROM ${ALPINE_IMAGE}

ARG RESTY_VERSION
ARG RESTY_OPENSSL_VERSION
ARG RESTY_OPENSSL_PATCH_VERSION
ARG RESTY_PCRE_VERSION
ARG RESTY_LUAROCKS_VERSION
ARG RESTY_USER
ARG RESTY_UID
ARG RESTY_GID

ENV RESTY_VERSION="${RESTY_VERSION}" \
    RESTY_OPENSSL_VERSION="${RESTY_OPENSSL_VERSION}" \
    RESTY_OPENSSL_PATCH_VERSION="${RESTY_OPENSSL_PATCH_VERSION}" \
    RESTY_PCRE_VERSION="${RESTY_PCRE_VERSION}" \
    RESTY_LUAROCKS_VERSION="${RESTY_LUAROCKS_VERSION}" \
    PATH="/usr/lib/openresty/luajit/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" \
    LUA_PATH="/usr/lib/openresty/site/lualib/?.lua;/usr/lib/openresty/site/lualib/?/init.lua;/usr/lib/openresty/lualib/?.lua;/usr/lib/openresty/lualib/?/init.lua;/usr/lib/openresty/luajit/share/luajit-2.1/?.lua;/usr/lib/openresty/luajit/share/lua/5.1/?.lua;/usr/lib/openresty/luajit/share/lua/5.1/?/init.lua" \
    LUA_CPATH="/usr/lib/openresty/site/lualib/?.so;/usr/lib/openresty/lualib/?.so;/usr/lib/openresty/luajit/lib/lua/5.1/?.so"

COPY --from=builder /tmp/export/ /

# apk-tools in the Alpine base image already depends on ca-certificates-bundle.
RUN apk add --no-cache libgcc mariadb-connector-c tzdata zlib \
    && addgroup -S -g "${RESTY_GID}" "${RESTY_USER}" \
    && adduser -S -D -H \
        -u "${RESTY_UID}" \
        -h /var/cache/openresty \
        -s /sbin/nologin \
        -G "${RESTY_USER}" \
        "${RESTY_USER}" \
    && ln -sf /dev/stdout /var/log/openresty/access.log \
    && ln -sf /dev/stderr /var/log/openresty/error.log

EXPOSE 80
# https://github.com/openresty/docker-openresty/blob/master/README.md#tips--pitfalls
STOPSIGNAL SIGQUIT
CMD ["openresty", "-g", "daemon off;"]
