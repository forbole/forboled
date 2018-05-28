FROM alpine:edge

# Set up dependencies
ENV PACKAGES go make git libc-dev

# Set up GOPATH & PATH

ENV GOPATH       /root/go
ENV BASE_PATH    $GOPATH/src/github.com/forbole
ENV REPO_PATH    $BASE_PATH/forboled
ENV WORKDIR      /cosmos/
ENV PATH         $GOPATH/bin:$PATH

# Link expected Go repo path

RUN mkdir -p $WORKDIR $GOPATH/pkg $ $GOPATH/bin $BASE_PATH

# Add source files

ADD . $REPO_PATH

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN apk add --no-cache $PACKAGES bash curl jq && \
    cd $REPO_PATH && make get_tools && make get_vendor_deps && make install && \
    apk del $PACKAGES

# Set entrypoint

ENTRYPOINT ["forboled"]
