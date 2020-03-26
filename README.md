# Jet

![Jet mascot](https://static.lade.io/jet-mascot.png)

[![Build Status](https://travis-ci.com/lade-io/jet.svg?branch=master)](https://travis-ci.com/lade-io/jet)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg)](https://godoc.org/github.com/lade-io/jet/pack)
[![Release](https://img.shields.io/github/v/release/lade-io/jet.svg)](https://github.com/lade-io/jet/releases/latest)

Jet is a tool to convert source code into Docker images. Jet inspects your source code to
create a Dockerfile with caching layers and any required system dependencies.

## Language Support

Jet will detect your app from the following languages and package managers:

* [Go](https://golang.org) - [dep](https://github.com/golang/dep), [glide](https://glide.sh), [godep](https://github.com/tools/godep), [go modules](https://github.com/golang/go/wiki/Modules), [govendor](https://github.com/kardianos/govendor)
* [Node.js](https://nodejs.org) - [npm](https://www.npmjs.com), [yarn](https://yarnpkg.com)
* [PHP](https://www.php.net) - [composer](https://getcomposer.org)
* [Python](https://www.python.org) - [conda](https://docs.conda.io), [pip](https://pip.pypa.io), [pipenv](https://pipenv.pypa.io)
* [Ruby](https://www.ruby-lang.org) - [bundler](https://bundler.io)

## Comparison Table

| Feature | Jet | [Cloud Native Buildpacks](https://buildpacks.io) | [Repo2docker](https://github.com/jupyter/repo2docker) | [Source-to-Image](https://github.com/openshift/source-to-image) |
| --- | --- | --- | --- | --- |
| Supported Languages | Go, Node.js, PHP, Python, Ruby | Java, Node.js | Python | Node.js, Perl, PHP, Python, Ruby |
| Best Practices Dockerfile | :white_check_mark: | :x: | :white_check_mark: | :x: |
| Hourly Runtime Updates | :white_check_mark: | :x: | :x: | :x: |

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

Build Node.js app:

```sh
$ jet build testdata/node/node12/ -n node-app
$ docker run -p 5000:5000 node-app
```

Debug Node.js app:

```console
$ jet debug testdata/node/node12/
FROM node:12

USER node
RUN mkdir -p /home/node/app/
WORKDIR /home/node/app/

COPY --chown=node:node package.json package-lock.json ./
RUN npm ci

COPY --chown=node:node . ./

CMD ["node", "server.js"]
```

Debug Python and Django app:

```console
$ jet debug testdata/python/django/
FROM python:3.5

ENV PATH=/home/web/.local/bin:$PATH
ENV PIP_USER=true

RUN groupadd --gid 1000 web \
        && useradd --uid 1000 --gid web --shell /bin/bash --create-home web

USER web
RUN mkdir -p /home/web/app/
WORKDIR /home/web/app/

COPY --chown=web:web requirements.txt ./
RUN pip install -r requirements.txt

COPY --chown=web:web . ./

CMD ["gunicorn", "django_web_app.wsgi:application"]
```

Debug Ruby on Rails app:

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

RUN groupadd --gid 1000 web \
        && useradd --uid 1000 --gid web --shell /bin/bash --create-home web

USER web
RUN mkdir -p /home/web/app/
WORKDIR /home/web/app/

COPY --chown=web:web Gemfile Gemfile.lock ./
RUN bundle install

COPY --chown=web:web package.json yarn.lock ./
RUN yarn install

COPY --chown=web:web . ./

CMD ["sh", "-c", "puma -p ${PORT-3000}"]
```

## Credits

* Test cases imported from [Cloud Foundry Buildpacks](https://github.com/cloudfoundry-community/cf-docs-contrib/wiki/Buildpacks)
licensed under [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0)
* Graphic designed by brgfx on [Freepik](http://www.freepik.com)
