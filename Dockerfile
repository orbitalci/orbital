FROM rust:1.38-slim
COPY . .
RUN apt update && apt install -y git pkg-config libssl-dev build-essential
RUN make release