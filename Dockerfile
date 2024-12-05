FROM ubuntu:24.04

# Install systemd and other dependencies
RUN apt-get update && \
    apt-get install -y systemd && \
    apt-get clean

# Set up systemd as the init system
ENV container docker
STOPSIGNAL SIGRTMIN+3

# Expose the necessary directory for the service
VOLUME [ "/sys/fs/cgroup" ]

# Copy your systemd unit file and any scripts or binaries needed
COPY go-self-update.service /etc/systemd/system/go-self-update.service
COPY main /usr/local/bin/main
RUN chmod +x /usr/local/bin/main

# Enable the service
RUN systemctl enable go-self-update

# Start systemd
CMD ["/lib/systemd/systemd"]
