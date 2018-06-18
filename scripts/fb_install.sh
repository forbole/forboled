#!/bin/bash

# Upgrade the system and install go
sudo apt update
sudo apt upgrade -y
sudo apt install gcc git make -y
sudo snap install --classic go
sudo mkdir -p /opt/go/bin

# Export environment variables
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$PATH

# Create a system user for running the service
sudo useradd -m -d /opt/forboled --system --shell /usr/sbin/nologin forboled
sudo -u forboled mkdir -p /opt/forboled/config

# Get Cosmos SDK and build binaries
go get github.com/forbole/forboled
cd $HOME/go/src/github.com/forbole/forboled
git fetch --all

make get_tools && make get_vendor_deps && make install
cd
# Copy the binaries to /opt/go/bin/
sudo cp $HOME/go/bin/forboled /opt/go/bin/
sudo cp $HOME/go/bin/fbcli /opt/go/bin/

# Create systemd unit file
echo "[Unit]
Description=Forbole Node
After=network-online.target
[Service]
User=forboled
ExecStart=/bin/sh -c '/opt/go/bin/forboled start --home=/opt/forboled/'
[Install]
WantedBy=multi-user.target" > forboled.service

sudo mv forboled.service /etc/systemd/system/
sudo systemctl enable forboled.service

# Create the config skeleton as user forboled
sudo -u forboled /opt/go/bin/forboled unsafe_reset_all --home=/opt/forboled

echo "You can prepare the genesis.json file in /opt/forboled/config/ and edit /opt/forboled/config/config.toml."
echo "Run 'sudo service forboled start' to start the service."

