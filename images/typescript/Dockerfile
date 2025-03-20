ARG NODE_VERSION=20.12

FROM node:${NODE_VERSION}-bookworm

RUN apt update && apt install -y --no-install-recommends tini && apt clean && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["sleep", "infinity"]