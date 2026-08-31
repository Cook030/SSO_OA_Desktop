FROM mh-acr-registry-vpc.cn-beijing.cr.aliyuncs.com/base/golang:20250402golang1.23-alinux20250102 as go
WORKDIR /src
COPY ./ /src

ARG GIT_USER
ARG GIT_TOKEN

RUN yum install -y tzdata libpq jemalloc unzip git

RUN echo "machine codeup.aliyun.com login ${GIT_USER} password ${GIT_TOKEN}" | cat > ~/.netrc && \
    source /root/.gvm/scripts/gvm && \
    go env -w GO111MODULE=on && \
    go env -w GOPROXY=https://goproxy.cn,direct && \
    go env -w GOPRIVATE=codeup.aliyun.com && \
    CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/server
###

FROM mh-acr-registry-vpc.cn-beijing.cr.aliyuncs.com/base/golang:20250402golang1.23-alinux20250102

RUN yum install -y tzdata libpq jemalloc procps-ng

RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

COPY --from=go /src/main /app
COPY --from=go /src/config/config.yaml /app/config/config.yaml
COPY --from=go /src/run.sh /app/run.sh

RUN echo "/usr/local/lib64" >> /etc/ld.so.conf.d/holo.conf && \
    echo "/usr/lib" >> /etc/ld.so.conf.d/holo.conf && \
    ldconfig

CMD /app/run.sh

# CMD [ "/app/main" ]
