ARG PYTHON_VERSION=3.13

FROM python:${PYTHON_VERSION}-bookworm

RUN apt update && apt install -y --no-install-recommends tini && apt clean && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["sleep", "infinity"]