# Jet

![Jet mascot](https://static.lade.io/jet-mascot.png)

[![Build Status](https://img.shields.io/travis/com/lade-io/jet.svg)](https://travis-ci.com/lade-io/jet)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg)](https://godoc.org/github.com/lade-io/jet/pack)
[![Release](https://img.shields.io/github/release/lade-io/jet.svg)](https://github.com/lade-io/jet/releases/latest)

Jet is a tool to convert source code into Docker images. Jet inspects your source code to
create a Dockerfile with caching layers and any required system dependencies.

## Installation

Jet is supported on MacOS, Linux and Windows as a standalone binary.
You can download the latest binary from the [releases page](https://github.com/lade-io/jet/releases) on Github.

### MacOS

You can install with [Homebrew](https://brew.sh):

```sh
brew install lade-io/tap/jet
```

### Linux

You can download the latest tarball, extract and move to your `$PATH`:

```sh
curl -L https://github.com/lade-io/jet/releases/latest/download/jet-linux-amd64.tar.gz | tar xz
sudo mv jet /usr/local/bin
```

## Build From Source

You can build from source with [Go](https://golang.org):

```sh
go get github.com/lade-io/jet
```

## Examples

Build Node app:

```sh
$ jet build testdata/node/node10/ -n node-app
$ docker run -p 5000:5000 node-app
```

Debug Django app with Gunicorn:

```console
$ jet debug testdata/python/django/
FROM python:3.5

WORKDIR /app/

COPY requirements.txt /app/
RUN pip install -r requirements.txt

COPY . /app/

CMD ["gunicorn", "django_web_app.wsgi:application"]
```

Debug Rails app with Yarn:

```console
$ jet debug testdata/ruby/rails5/
FROM ruby:2.6.5

RUN set -ex \
        && echo "deb http://deb.nodesource.com/node_11.x stretch main" > /etc/apt/sources.list.d/nodesource.list \
        && curl -fsSL https://deb.nodesource.com/gpgkey/nodesource.gpg.key | apt-key add - \
        && echo "deb http://dl.yarnpkg.com/debian/ stable main" > /etc/apt/sources.list.d/yarn.list \
        && curl -fsSL https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - \
        && apt-get update && apt-get install -y \
                nodejs \
                yarn \
        && rm -rf /var/lib/apt/lists/*

WORKDIR /app/

COPY Gemfile Gemfile.lock /app/
RUN bundle install

COPY package.json yarn.lock /app/
RUN yarn install

COPY . /app/

CMD ["sh", "-c", "puma -p ${PORT-3000}"]
```

## Credits

* Test cases imported from [Cloud Foundry Buildpacks](https://github.com/cloudfoundry-community/cf-docs-contrib/wiki/Buildpacks)
licensed under [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0)
* Graphic designed by brgfx on [Freepik](http://www.freepik.com)
