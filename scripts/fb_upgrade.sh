#!/bin/bash
# Upgrade forboled and restart service

POSITIONAL=()
while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -t|--tag)
    TAG="$2"
    shift # past argument
    shift # past value
    ;;
    --default)
    DEFAULT=YES
    shift # past argument
    ;;
    *)    # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift # past argument
    ;;
esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

if [ -z "$TAG" ]; then
    echo "Usage: fb_upgrade.sh -t v0.19.0"
    exit 1
fi

GOPATH=$HOME/go
PATH=$GOPATH/bin:$PATH

# upgarde system packages
sudo apt upgrade -y

# get latest forboled software
go get github.com/forbole/forboled
cd $GOPATH/src/github.com/forbole/forboled
git fetch --all
git checkout -f $TAG
make update_tools && make get_vendor_deps && make install

echo ""
echo "forboled " $TAG "installed."
echo ""
# Log installed versions
echo "forboled version :" $(forboled version)
echo "fbcli version :" $(fbcli version)
cd
echo "Stopping forboled"
sudo service forboled stop
sudo cp go/bin/forboled /opt/go/bin/
sudo cp go/bin/fbcli /opt/go/bin/
echo "Starting forboled"
sudo service forboled start
echo "forboled restarted."


