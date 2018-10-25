FROM liuchong/rustup:nightly

WORKDIR /usr/src/dev

COPY Cargo.toml Cargo.toml
COPY Cargo.lock Cargo.lock
RUN mkdir src && touch src/main.rs
RUN echo "fn main() { println!(\"Hello World\"); }" >> src/main.rs
RUN cargo check

ENV ROCKET_ADDRESS=0.0.0.0
ENV ROCKET_PORT=8000
ENV API_KEY=4dLAEkep2UxTYXms2ZwGxkgWgwcb7mSd

COPY src src

RUN cargo build --release

CMD cargo run --release
