# Elixir Testnet v3 Validator Autoupdate

This program registers itself as a `systemd` service and periodically updates docker container with Elixir Testnet v3
Validator node.

## WARNING

THIS SOFTWARE IS PROVIDED AS IS. USE AT YOUR OWN RISK.

AUTHOR DOES NOT HAVE ANY RESPONSIBILITY FOR YOUR ACTIONS WITH IT.

## Steps to install:

- Prepare your VPS or other kind of server.
- Follow the instruction at https://docs.elixir.xyz/running-an-elixir-validator up to 'Running Your Validator' point.
  Don't do commands 'docker pull' and 'docker start'. At this point you should have prepared `validator.env` file
  containing all necessary values.
- If you need TG Bot notifications, create your Telegram Bot for this. Use https://t.me/BotFather. The token it will
  give to you should be used in `config.yml` in next step.
- Prepare configuration file `config.yml` for this tool. See example in `config.example.yml`. Explanation of options  
  written below.
- Install Golang compiler v1.22 or higher. Seek here: https://go.dev/dl/
- Prepare `config.sh` if you will deploy to VPS via script. Example is in `config.example.sh`.
- Compile this tool. You may use `go build` command or look into `update.sh` script. Try `go mod tidy` if you will
  encounter problems with Go modules.
- Deploy the compiled binary.

After tool will started, just wait and explore logs. Commands like:

- `journalctl -u elixir-updater -n 100 -f` for systemd service log,
- `docker logs -n 100 -f $(docker ps | grep elixir | awk '{print $1}')` for Elixir container logs

After starting the tool, write something like `/start` to your created TG bot in Telegram.

## config.yml options

| option             | type   | default value               | meaning                                               |
|--------------------|--------|-----------------------------|-------------------------------------------------------|
| tg_bot_token       | string | ""                          | TG Bot Token                                          |
| tg_force_chat_id   | int64  | 0                           | Forces Chat ID to this value, if known. Leave zero    |
| user               | string | "root"                      | User to run service under                             |
| container_name     | string | "elixir"                    | Docker container name to create                       |
| restart_policy     | string | "unless-stopped"            | Docker container restart policy                       |
| env_file_path      | string | "/opt/elixir/validator.env" | Path to env file for the Docker container             |
| service_name       | string | "elixir-updater"            | Systemd service name                                  |
| host               | string | "http://localhost"          | Path to retrieve metrics over HTTP from the container |
| port               | string | "17690"                     | Port to retrieve metrics over HTTP from the container |
| docker_api_version | string | "1.42"                      | Max supported Docker API version                      |

## config.sh vars

| variable | type            | meaning                                                       |
|----------|-----------------|---------------------------------------------------------------|
| servers  | array of string | Server addr to connect via SSH, format: username@host         |
| ports    | array of string | SSH ports in the same order                                   |
| paths    | array of string | Paths to copy this tool binary, order must be same as servers |

Enjoy.