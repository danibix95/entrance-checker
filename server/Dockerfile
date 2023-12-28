FROM debian:bookworm-slim as tini-base

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini.asc /tini.asc

RUN apt-get update \
    && apt-get install -y gpg \
    && apt-get clean autoclean -y \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/*

RUN gpg --batch --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 595E85A6B1B4779EA4DAAEC70B588DFF0527A9B7 \
    && gpg --batch --verify /tini.asc /tini
RUN chmod +x /tini

# ======================================================== #
FROM rust:1.74-slim AS chef 

RUN cargo install cargo-chef 

WORKDIR /app

# ======================================================== #
FROM chef AS planner

COPY . .

RUN cargo chef prepare  --recipe-path recipe.json

# ======================================================== #
FROM chef AS builder

COPY --from=planner /app/recipe.json recipe.json

RUN cargo chef cook --release --recipe-path recipe.json

COPY . .

RUN cargo build -r

# ======================================================== #
FROM debian:bookworm-slim as final

COPY --from=tini-base --chown=1000:1000 --chmod=100 /tini /tini

ENTRYPOINT ["/tini", "--"]

WORKDIR /app

COPY --from=builder --chown=1000:1000 --chmod=100 /app/target/release/ecserver /app/ecserver

USER 1000

CMD [ "/app/ecserver" ]
