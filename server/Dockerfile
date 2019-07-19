## Component 1: build the application ##
FROM golang:alpine as builder

# install required packages to retrieve
# the dependencies and compile the source code
RUN apk add --no-cache git build-base

# enable the use of Go modules
ENV GO111MODULE on

WORKDIR $GOPATH/src/fdp_server

# retrieve module definition, which contains the list of dependencies
COPY go.mod .
COPY go.sum .

# download dependencies
RUN go mod download && go mod verify

# copy all application files (apart from the one
# excluded in the .dockerignore) into the container
COPY . .

# retrieve app dependencies and build it
RUN mkdir -p $GOPATH/bin
RUN GOOS=linux go build -a -installsuffix cgo -o $GOPATH/bin/fdp_server ./cmd/...


## Component 2: deploy the application ##
FROM golang:alpine as server

# install the package to deal with timezones
RUN apk add --no-cache tzdata
RUN cp /usr/share/zoneinfo/Europe/Rome /etc/localtime && echo "Europe/Rome" > /etc/timezone && date

# Import the user and group files from the builder.
RUN addgroup fdp_user && adduser -D -H -g '' -G fdp_user fdp_user

WORKDIR /app/

# copy application binary file
COPY --from=builder $GOPATH/bin/fdp_server .
COPY --from=builder $GOPATH/src/fdp_server/postgres_info .

EXPOSE 8080

# prepare the environment for the execution
RUN mkdir logs
RUN chmod 444 /app/postgres_info
RUN chown fdp_user:fdp_user -R /app/

# Use an unprivileged user to run the application
USER fdp_user

ENTRYPOINT ["./fdp_server"]