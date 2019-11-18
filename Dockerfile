FROM rust:1.39-slim
RUN rustup component add rustfmt
COPY . .
RUN apt update && apt install -y git pkg-config libssl-dev build-essential
RUN make release