FROM liuchong/rustup:nightly

WORKDIR /usr/src/dev

ENV ROCKET_ADDRESS=0.0.0.0
ENV ROCKET_PORT=8000
ENV API_KEY=4dLAEkep2UxTYXms2ZwGxkgWgwcb7mSd

COPY src src
COPY Cargo.lock Cargo.lock
COPY Cargo.toml Cargo.toml

RUN cargo build --release

CMD cargo run --release
