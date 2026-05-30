#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive
PHP_VERSION="${VERSION:-8.4}"

apt-get update -y
apt-get install -y --no-install-recommends \
  ca-certificates apt-transport-https lsb-release wget gnupg \
  # PHP build/runtime system libs
  libxml2-dev libcurl4-openssl-dev libzip-dev libpng-dev \
  libjpeg-dev libfreetype6-dev libonig-dev libssl-dev pkg-config

# Sury PHP repository
wget -qO /etc/apt/trusted.gpg.d/php.gpg https://packages.sury.org/php/apt.gpg
echo "deb https://packages.sury.org/php/ $(lsb_release -sc) main" \
  > /etc/apt/sources.list.d/php.list

apt-get update -y
apt-get install -y --no-install-recommends \
  "php${PHP_VERSION}" \
  "php${PHP_VERSION}-cli" \
  "php${PHP_VERSION}-xml" \
  "php${PHP_VERSION}-curl" \
  "php${PHP_VERSION}-zip" \
  "php${PHP_VERSION}-mbstring" \
  "php${PHP_VERSION}-gd" \
  composer

rm -rf /var/lib/apt/lists/*
echo "php feature: installed PHP ${PHP_VERSION} + composer"