FROM vault:0.10.1

RUN apk add --no-cache --update \
    bash \
    ca-certificates \
    curl \
    jq \
    python \
    su-exec \
    tar \
    wget


COPY vault-unsealer_linux_amd64 /usr/local/bin/vault-unsealer

# gcloud
ENV HOME /vault/file
ENV PATH /google-cloud-sdk/bin:$PATH
ENV CLOUDSDK_PYTHON_SITEPACKAGES 1
RUN wget https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk.zip && unzip google-cloud-sdk.zip && rm google-cloud-sdk.zip
RUN google-cloud-sdk/install.sh --usage-reporting=false --path-update=true --bash-completion=true --rc-path=/.bashrc
RUN google-cloud-sdk/bin/gcloud config set --installation component_manager/disable_update_check true
