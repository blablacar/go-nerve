FROM golang:1.12 as builder

RUN apt-get update && apt-get install -y \
    xz-utils \
&& rm -rf /var/lib/apt/lists/*

ADD https://github.com/upx/upx/releases/download/v3.94/upx-3.94-amd64_linux.tar.xz /usr/local
RUN xz -d -c /usr/local/upx-3.94-amd64_linux.tar.xz | \
    tar -xOf - upx-3.94-amd64_linux/upx > /bin/upx && \
    chmod a+x /bin/upx

WORKDIR /app
COPY . ./
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor && cp go-nerve /nerve
RUN ./gomake build && cp ./dist/nerve-v0-linux-amd64/nerve /

FROM busybox

COPY --from=builder /nerve /
COPY ./examples/nerve-full-templated.yml /nerve.yml

CMD ["/nerve", "/nerve.yml"]
