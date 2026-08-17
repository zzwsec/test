# syntax=docker/dockerfile:1

ARG OPENRESTY_IMAGE="test-openresty"
ARG RESTY_VERSION="1.31.1.1"
ARG RESTY_OPENSSL_VERSION="3.5.7"
ARG RESTY_OPENSSL_PATCH_VERSION="3.5.5"
ARG RESTY_LUAROCKS_VERSION="3.13.0"

FROM ${OPENRESTY_IMAGE}:${RESTY_VERSION} AS rocks-builder

ARG RESTY_VERSION
ARG RESTY_OPENSSL_VERSION
ARG RESTY_OPENSSL_PATCH_VERSION
ARG RESTY_LUAROCKS_VERSION
ARG RESTY_J=""

# Keep luaossl aligned with the private OpenSSL used by the static image.
ARG RESTY_OPENSSL_URL_BASE="https://github.com/openssl/openssl/releases/download/openssl-${RESTY_OPENSSL_VERSION}"
ARG RESTY_OPENSSL_BUILD_OPTIONS="\
    enable-camellia \
    enable-cms \
    enable-fips \
    enable-ktls \
    enable-md2 \
    enable-rc5 \
    enable-rfc3779 \
    enable-seed \
    enable-ssl3 \
    enable-ssl3-method \
    enable-weak-ssl-ciphers \
    "

ENV OPENRESTY_SOURCE_DIR="/usr/src/openresty-${RESTY_VERSION}" \
    OPENSSL_SOURCE_DIR="/usr/src/openssl-${RESTY_OPENSSL_VERSION}" \
    LUAROCKS_SOURCE_DIR="/usr/src/luarocks-${RESTY_LUAROCKS_VERSION}" \
    LUAROCKS_TREE="/usr/lib/openresty/luajit"

RUN apk add --no-cache build-base curl git linux-headers \
    mariadb-connector-c-dev perl pkgconf unzip wget zlib-dev

WORKDIR /usr/src
RUN curl -fsSL "${RESTY_OPENSSL_URL_BASE}/openssl-${RESTY_OPENSSL_VERSION}.tar.gz" | tar xzf - -C /usr/src \
    && curl -fsSL "https://raw.githubusercontent.com/openresty/openresty/master/patches/openssl-${RESTY_OPENSSL_PATCH_VERSION}-sess_set_get_cb_yield.patch" | patch -d "${OPENSSL_SOURCE_DIR}" -p1 \
    && curl -fsSL "https://openresty.org/download/openresty-${RESTY_VERSION}.tar.gz" | tar xzf - -C /usr/src \
    && curl -fsSL "https://luarocks.org/releases/luarocks-${RESTY_LUAROCKS_VERSION}.tar.gz" | tar xzf - -C /usr/src \
    && mkdir -p /usr/lib/openresty/luajit/include/luajit-2.1 \
    && cp "${OPENRESTY_SOURCE_DIR}"/bundle/LuaJIT-*/src/*.h /usr/lib/openresty/luajit/include/luajit-2.1/

# luaossl is a native rock and needs the OpenSSL development files removed from
# the static runtime image, so build the same patched private OpenSSL here.
WORKDIR ${OPENSSL_SOURCE_DIR}
RUN ROCKS_BUILD_JOBS="${RESTY_J:-$(nproc)}" \
    && ./config \
        shared zlib -g \
        --prefix=/usr/lib/openresty/openssl3 \
        --openssldir=/etc/ssl \
        --libdir=lib \
        -Wl,-rpath,/usr/lib/openresty/openssl3/lib \
        ${RESTY_OPENSSL_BUILD_OPTIONS} \
    && make -j"${ROCKS_BUILD_JOBS}" \
    && make -j"${ROCKS_BUILD_JOBS}" install_sw

WORKDIR ${LUAROCKS_SOURCE_DIR}
RUN ./configure \
        --prefix=/usr/local/luarocks \
        --rocks-tree="${LUAROCKS_TREE}" \
        --with-lua=/usr/lib/openresty/luajit \
        --with-lua-include=/usr/lib/openresty/luajit/include/luajit-2.1 \
    && make \
    && make install

ENV PATH="/usr/local/luarocks/bin:${PATH}"

RUN CFLAGS="-O2 -fPIC" \
        LDFLAGS="-L/usr/lib/openresty/openssl3/lib -Wl,-rpath,/usr/lib/openresty/openssl3/lib" \
        luarocks install luaossl \
        CRYPTO_INCDIR=/usr/lib/openresty/openssl3/include \
        CRYPTO_LIBDIR=/usr/lib/openresty/openssl3/lib \
        OPENSSL_INCDIR=/usr/lib/openresty/openssl3/include \
        OPENSSL_LIBDIR=/usr/lib/openresty/openssl3/lib \
    && luarocks install lapis \
    && luarocks install lua-resty-mlcache \
    && luarocks install lua-resty-jit-uuid \
    && luarocks install luasql-mysql MYSQL_INCDIR=/usr/include/mysql \
    && find "${LUAROCKS_TREE}/lib/lua/5.1" -type f -name '*.so' -exec strip --strip-unneeded '{}' \; \
    && rm -rf \
        "${LUAROCKS_TREE}/include" \
        "${LUAROCKS_TREE}/lib/pkgconfig" \
        "${LUAROCKS_TREE}/share/luajit-2.1" \
    && rm -f \
        "${LUAROCKS_TREE}"/bin/luajit* \
        "${LUAROCKS_TREE}"/lib/libluajit-5.1.so*

FROM ${OPENRESTY_IMAGE}:${RESTY_VERSION}

ARG RESTY_LUAROCKS_VERSION
ENV RESTY_LUAROCKS_VERSION="${RESTY_LUAROCKS_VERSION}"

COPY --from=rocks-builder /usr/lib/openresty/luajit/ /usr/lib/openresty/luajit/

RUN apk add --no-cache mariadb-connector-c
