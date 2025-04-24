FROM golang
RUN mkdir -p /go/src/pipeline
WORKDIR /go/src/pipeline
ADD main.go .
ADD go.mod .
RUN go install .

FROM alpine:latest
LABEL version="1.0"
LABEL maintainer="Serzh-Zolotarev<serzh.zolotarev2014-2015@yandex.ru>"
WORKDIR /root/
COPY --from=0 /go/bin/pipeline .
ENTRYPOINT ./pipeline
EXPOSE 8080