FROM node:18.15.0-alpine as build

ARG COMMIT_SHA=<not-specified>
ENV NPM_CONFIG_CACHE="/tmp"

WORKDIR /build-dir

COPY package.json .
COPY package-lock.json .

# install also DEV dependencies to enable Typescript compilation
RUN npm ci

COPY . .

# compile Typescript to Javascript
RUN npm run build

RUN echo "fdp-backend: $COMMIT_SHA" >> ./commit.sha

########################################################################################################################

FROM node:18.15.0-alpine

RUN apk add --no-cache tini

LABEL maintainer="Daniele Bissoli" name="fdp-backend"

ENV NODE_ENV=production
ENV LOG_LEVEL=info
ENV SERVICE_PREFIX=/
ENV HTTP_PORT=3000
ENV NPM_CONFIG_CACHE="/tmp"

WORKDIR /home/node/app

# copy only needed files
COPY --from=build /build-dir/package.json package.json
COPY --from=build /build-dir/package-lock.json package-lock.json
COPY --from=build /build-dir/LICENSE LICENSE
COPY --from=build /build-dir/CHANGELOG.md CHANGELOG.md
COPY --from=build /build-dir/dist dist
COPY --from=build /build-dir/commit.sha commit.sha

# install only dependencies required to run the service
RUN npm ci --omit=dev

USER node

ENTRYPOINT ["/sbin/tini", "--"]

CMD ./node_modules/.bin/lc39 ./dist/index.js --port=${HTTP_PORT} --log-level=${LOG_LEVEL}
