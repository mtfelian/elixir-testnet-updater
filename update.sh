#!/bin/bash

source ./config.sh

identity=~/.ssh/id_rsa
binaryName=elixir-testnet-updater
configName=config.yml

GOOS=linux go build

update() {
  local server=$1
  local path=$2
  echo ">>> Updating server $server:$path"
  ssh -i $identity $server "service elixir-updater stop"
  scp -i $identity $binaryName $server:$path/$binaryName
  scp -i $identity $configName $server:$path/$configName
  ssh -i $identity $server "service elixir-updater start"
}


for (( j=0; j<"${#servers[@]}"; j++ )); do
    server=${servers[$j]}
    path=${paths[$j]}
    update "$server" "$path"
done

