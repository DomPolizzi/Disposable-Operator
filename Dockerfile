# Image used by the CI process
FROM python:3.10-buster

RUN apt-get update -y >/dev/null && apt-get install jq lsb-release -y

RUN wget --quiet https://dev.mysql.com/get/mysql-apt-config_0.8.22-1_all.deb -O mysql.deb
RUN DEBIAN_FRONTEND=noninteractive dpkg -i mysql.deb
RUN apt-get update -y >/dev/null && apt-get install default-mysql-client -y

COPY requirements.txt requirements.txt
COPY requirements.yml requirements.yml

RUN pip install -r requirements.txt

RUN ansible-galaxy install -r requirements.yml

ARG TARGETARCH="amd64"

RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/${TARGETARCH}/kubectl"  \
  && install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

ARG HELM_VERSION="3.11.3"
ARG HELMFILE_VERSION="0.152.0"
ARG HELM_DIFF_VERSION="3.8.2"

RUN wget https://get.helm.sh/helm-v${HELM_VERSION}-linux-${TARGETARCH}.tar.gz \
  && tar -xvf helm-v${HELM_VERSION}-linux-${TARGETARCH}.tar.gz \
  && mv linux-${TARGETARCH}/helm /usr/local/bin/helm \
  && rm -rv helm-v${HELM_VERSION}-linux-${TARGETARCH}.tar.gz linux-${TARGETARCH}/

# Later switch to new helmfile https://github.com/helmfile/helmfile
RUN wget https://github.com/helmfile/helmfile/releases/download/v${HELMFILE_VERSION}/helmfile_${HELMFILE_VERSION}_linux_${TARGETARCH}.tar.gz \
  && tar -xvf helmfile_${HELMFILE_VERSION}_linux_${TARGETARCH}.tar.gz \
  && mv helmfile /usr/local/bin/helmfile \
  && chmod u+x /usr/local/bin/helmfile \
  && rm -rv helmfile_${HELMFILE_VERSION}_linux_${TARGETARCH}.tar.gz

# Install helm-diff as it's needed by helmfile
RUN helm plugin install https://github.com/databus23/helm-diff
