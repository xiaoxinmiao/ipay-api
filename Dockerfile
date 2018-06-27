FROM pangpanglabs/golang:jan AS builder
WORKDIR /go/src/ipay-api/
COPY ./ /go/src/ipay-api/
# disable cgo 
ENV CGO_ENABLED=0
# build steps
RUN echo ">>> 1: go version" && go version \
    && echo ">>> 2: go get" && go-wrapper download \
    && echo ">>> 3: go install" && go-wrapper install

# make application docker image use alpine
FROM  alpine:3.6
RUN apk --no-cache add ca-certificates
WORKDIR /go/bin
# copy config cert to image 
COPY ./tmp/ ./tmp/
RUN tar xf ./tmp/wxcert.tar.gz -C ./
# copy execute file to image
COPY --from=builder /go/bin/ ./
COPY --from=builder /go/src/ipay-api/*.yml ./
COPY --from=builder /swagger-ui/ ./swagger-ui/
COPY --from=builder /go/src/ipay-api/index.html ./swagger-ui/
EXPOSE 8080
CMD ["./ipay-api"]
