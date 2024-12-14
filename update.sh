#!/bin/bash

source ./config.sh

identity=~/.ssh/id_rsa
binaryName=elixir-testnet-updater
configName=config.yml

GOOS=linux go build

update() {
  local server=$1
  local port=$2
  local path=$3
  local tag=$4
  echo ">>> Updating server $server:$path, SSH port $port, tag is $tag"
  ssh -p $port -i $identity $server "service elixir-updater stop"
  scp -P $port -i $identity $binaryName $server:$path/$binaryName
  sed -E -i 's/elixirprotocol\/validator:[a-z]+/elixirprotocol\/validator:'$tag'/g' $configName
  scp -P $port -i $identity $configName $server:$path/$configName
  ssh -p $port -i $identity $server "service elixir-updater start"
}


for (( j=0; j<"${#servers[@]}"; j++ )); do
    server=${servers[$j]}
    port=${ports[$j]}
    path=${paths[$j]}
    tag=${tags[$j]}
    update "$server" "$port" "$path" "$tag"
done

