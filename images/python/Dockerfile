ARG PYTHON_VERSION=3.13

FROM python:${PYTHON_VERSION}-bookworm

RUN apt update && apt install -y --no-install-recommends tini && apt clean && rm -rf /var/lib/apt/lists/* \
    && curl -LsSf https://astral.sh/uv/install.sh | sh

WORKDIR /root

# Set global PATH for uv
ENV PATH="/root/.local/bin:${PATH}"

# Install python packages
RUN uv pip install --system numpy scipy pandas scikit-learn requests

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["sleep", "infinity"]