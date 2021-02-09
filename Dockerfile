#FROM rust:1.49-slim
#RUN rustup component add rustfmt
#COPY . .
#RUN apt update && apt install -y git pkg-config libssl-dev libpq-dev build-essential
#RUN make release

FROM rust:1.49-alpine
RUN rustup component add rustfmt
COPY . .
RUN apk update && apk add make build-base openssl-libs-static protoc
RUN make release