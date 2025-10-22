FROM docker.io/library/ubuntu:24.04
ARG STRONGSWAN_VERSION=5.8.4

RUN \
  apt-get update &&\
  apt-get install -y \
  build-essential \
  libgmp-dev \
  libunbound-dev \
  libldns-dev \
  pkg-config \
  libsystemd-dev \
  wget \
  curl \
  supervisor \
  git

RUN curl -sL https://deb.nodesource.com/setup_12.x | bash - &&\
  apt-get install -y nodejs
RUN npm install nodemon -g

RUN curl -LO https://github.com/hairyhenderson/gomplate/releases/download/$(curl -Lso /dev/null -w %{url_effective} https://github.com/hairyhenderson/gomplate/releases/latest | grep -o '[^/]*$')/gomplate_linux-amd64 &&\
  chmod +x gomplate_linux-amd64 &&\
  mv gomplate_linux-amd64 /usr/bin/gomplate

RUN \
  wget https://download.strongswan.org/strongswan-$STRONGSWAN_VERSION.tar.gz &&\
  wget https://download.strongswan.org/strongswan-$STRONGSWAN_VERSION.tar.gz.md5 &&\
  md5sum -c strongswan-$STRONGSWAN_VERSION.tar.gz.md5
RUN \
  tar xvzf strongswan-$STRONGSWAN_VERSION.tar.gz &&\
  cd strongswan-$STRONGSWAN_VERSION &&\
  ./configure \
  --prefix=/usr \
  --sysconfdir=/etc \
  --enable-eap-mschapv2 \
  --enable-kernel-libipsec \
  --enable-swanctl \
  --enable-unity \
  --enable-unbound \
  --enable-vici \
  --enable-xauth-eap \
  --enable-xauth-noauth \
  --enable-eap-identity \
  --enable-md4 \
  --enable-pem \
  --enable-openssl \
  --enable-pubkey \
  --enable-farp \
  --enable-systemd \
  --disable-charon \
  --disable-stroke \
  --disable-scepclient &&\
  make &&\
  make install

ADD . /

CMD ["/start.sh"]