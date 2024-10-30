#FROM golang:bookworm
#FROM cloudfoundry/cflinuxfs4
FROM eight/builder

# Copy in the binaries
WORKDIR /home/vcap/app
COPY . .
WORKDIR /home/vcap/app/cmd/serve
RUN make container_build
RUN chmod 755 /home/vcap/app/cmd/serve/run.sh

CMD ["/home/vcap/app/cmd/serve/run-fast.sh"]


