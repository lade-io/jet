# Jet

[![Build Status](https://img.shields.io/github/actions/workflow/status/lade-io/jet/release.yml)](https://github.com/lade-io/jet/actions/workflows/release.yml)
[![Go.Dev Reference](https://img.shields.io/badge/go.dev-reference-blue.svg)](https://pkg.go.dev/github.com/lade-io/jet/pack)
[![Release](https://img.shields.io/github/v/release/lade-io/jet.svg)](https://github.com/lade-io/jet/releases/latest)

Jet is a tool to convert source code into Docker images. Jet inspects your source code to
create a Dockerfile with caching layers and any required system dependencies.

## Language Support

Jet will detect your app from the following languages and package managers:

* [Go](https://golang.org) - [dep](https://github.com/golang/dep), [glide](https://github.com/Masterminds/glide), [godep](https://github.com/tools/godep), [go modules](https://github.com/golang/go/wiki/Modules), [govendor](https://github.com/kardianos/govendor)
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
You can download the latest binary from the [releases page](https://github.com/lade-io/jet/releases) on GitHub.

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

ENV PATH=/home/node/app/node_modules/.bin:$PATH

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
FROM python:3.9

ENV PATH=/home/web/.local/bin:$PATH

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
FROM ruby:2.7.7

RUN wget -qO node.tar.gz "https://nodejs.org/dist/v18.13.0/node-v18.13.0-linux-x64.tar.gz" \
        && tar -xzf node.tar.gz -C /usr/local --strip-components=1 \
        && rm node.tar.gz

RUN corepack enable

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
