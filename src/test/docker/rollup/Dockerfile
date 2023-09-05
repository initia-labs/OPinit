ARG INITIA_VERSION=latest

FROM ghcr.io/initia-labs/initiad:${INITIA_VERSION}

USER root
COPY --chmod=755 entrypoint.sh /usr/local/bin/entrypoint.sh

RUN set -eux &&\
    mkdir -p /app/config && \
    mkdir -p /app/data && \
    chown -R initia:initia /app && \
    initiad init local-initia --home /app --chain-id local-initia && \
    echo '{"height": "0","round": 0,"step": 0}' > /app/data/priv_validator_state.json && \
    sed -e '/^\[api\]/,/\[rosetta\]/ s|^enable *=.*|enable = true|' \
        -e '/^\[api\]/,/\[rosetta\]/ s|^swagger *=.*|swagger = true|' \ 
        -e '/^\[api\]/,/\[rosetta\]/ s|^enabled-unsafe-cors *=.*|enabled-unsafe-cors = true|' \ 
        -i /app/config/app.toml 

COPY ./initia/genesis.json \
    ./initia/priv_validator_key.json \
    ./initia/addrbook.json \
    ./initia/gentx \
    /app/config/


ENTRYPOINT [ "entrypoint.sh" ]

# rest server
EXPOSE 1317
# nginx
EXPOSE 8080
# grpc
EXPOSE 9090
# grpc-web
EXPOSE 9091
# tendermint p2p
EXPOSE 26656
# tendermint rpc
EXPOSE 26657

CMD initiad start --home /app
