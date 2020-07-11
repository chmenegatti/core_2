FROM registry.gitlab.com/ascenty/builder:latest AS builder
ARG repo
COPY . /app
WORKDIR /app
RUN glide cc
RUN glide update
RUN mv  vendor src
RUN mkdir /app/src/git-devops.totvs.com.br${repo}
RUN mv /app/service /app/src/git-devops.totvs.com.br${repo}
RUN GOPATH='/app' CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o executavel main.go
RUN upx --brute executavel

FROM scratch
WORKDIR /
COPY --from=builder /app/executavel .
CMD ["/executavel"]
