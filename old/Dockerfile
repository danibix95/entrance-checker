FROM node:18.15.0-alpine as build

ARG COMMIT_SHA=<not-specified>
ENV NPM_CONFIG_CACHE="/tmp"

WORKDIR /build-dir

COPY package.json .
COPY package-lock.json .

# install also DEV dependencies to enable Typescript compilation
RUN npm ci

COPY . .

########################################################################################################################

FROM node:18.15.0-alpine

RUN apk add --no-cache tini

LABEL maintainer="Daniele Bissoli" name="fdp-app"

ENV NODE_ENV=production
ENV LOG_LEVEL=info
ENV SERVICE_PREFIX=/
ENV HTTP_PORT=3000
ENV NPM_CONFIG_CACHE="/tmp"

WORKDIR /home/node/app

COPY --from=build /build-dir ./

USER node

ENTRYPOINT ["/sbin/tini", "--"]

CMD node index.js