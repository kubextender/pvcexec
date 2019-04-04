#!/usr/bin/env bash

if [[ "$OSTYPE" == "linux-gnu" ]]; then
  OS='Linux'
  URL=`curl -s https://api.github.com/repos/kubextender/pvcexec/releases/latest | grep browser_download_url | grep _linux_amd64 | cut -d '"' -f 4`
elif [[ "$OSTYPE" == "darwin"* ]]; then
  OS='MacOS'
  URL=`curl -s https://api.github.com/repos/kubextender/pvcexec/releases/latest | grep browser_download_url | grep _darwin_amd64 | cut -d '"' -f 4`
fi

echo Downloading "$OS" binary from $URL ...
BINARY_PATH='/usr/local/bin/kubectl-pvcexec'
curl -L -s -o "$BINARY_PATH" "$URL"

chmod +x "$BINARY_PATH"

echo "Binary installed successfully to $BINARY_PATH"
echo Done!