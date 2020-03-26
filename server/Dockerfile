## Component 1: build the application ##
FROM golang:alpine as builder

# install required packages to retrieve
# the dependencies and compile the source code
RUN apk add --no-cache git build-base

# Note: from Go 1.13 modules are enabled by default
ENV GOBIN /app/bin

WORKDIR /app/fdp_server

# copy module definition for retrieving dependencies packages
COPY fdp_server/go.mod .
COPY fdp_server/go.sum .

# download dependencies
RUN go mod download && go mod verify

# copy application files (excluding everithing reported in the .dockerignore file)
COPY fdp_server/app.go .
COPY fdp_server/controller ./controller
COPY fdp_server/dbconn ./dbconn

# retrieve app dependencies and build it
RUN GOOS=linux go install -a -installsuffix cgo


## Component 2: deploy the application ##
FROM golang:alpine as server

# install the package to deal with timezones
RUN apk add --no-cache tzdata
RUN cp /usr/share/zoneinfo/Europe/Rome /etc/localtime && echo "Europe/Rome" > /etc/timezone && date

ENV GOBIN /app/bin

# Import the user and group files from the builder.
RUN addgroup fdp_user && adduser -D -g '' -G fdp_user fdp_user

WORKDIR /app/

# copy application binary file
COPY --from=builder $GOBIN/fdp_server .
COPY postgres_info .

EXPOSE 8080

# prepare the environment for the execution
RUN mkdir logs
RUN chmod 400 /app/postgres_info
RUN chown fdp_user:fdp_user -R /app/

# Use an unprivileged user to run the application
USER fdp_user

ENTRYPOINT ["./fdp_server"]