FROM rust:1.45-slim
RUN rustup component add rustfmt
COPY . .
RUN apt update && apt install -y git pkg-config libssl-dev libpq-dev build-essential
RUN make release