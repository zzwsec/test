ARG ALPINE_IMAGE="alpine:3.24.1"
ARG NGINX_VERSION="1.31.3"
ARG NGINX_IMAGE="zzwsec/nginx:${NGINX_VERSION}"

FROM ${ALPINE_IMAGE} AS builder
ARG NGINX_VERSION

RUN apk add --no-cache \
    build-base pcre2-dev zlib-dev openssl-dev linux-headers git curl \
    gd-dev libmaxminddb-dev libxml2-dev libxslt-dev

WORKDIR /usr/src
RUN curl -fsSL https://nginx.org/download/nginx-${NGINX_VERSION}.tar.gz | tar xzf - 

WORKDIR /usr/src/nginx-${NGINX_VERSION}
RUN git clone --depth=1 https://github.com/nginx/njs.git \
    && git clone --depth=1 https://github.com/leev/ngx_http_geoip2_module.git \
    && mkdir -p /tmp/modules-export

RUN ./configure \
    --with-compat \
    --with-stream \
    --with-http_image_filter_module=dynamic \
    --add-dynamic-module=njs/nginx \
    --add-dynamic-module=ngx_http_geoip2_module \
    && make modules \
    && strip objs/*.so \
    && cp objs/*.so /tmp/modules-export/

FROM ${NGINX_IMAGE}

RUN apk add --no-cache libgd libmaxminddb-libs libxml2 libxslt
COPY --from=builder /tmp/modules-export/ /usr/lib/nginx/modules/
