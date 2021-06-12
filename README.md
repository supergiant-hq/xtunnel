# xTUNNEL

xTUNNEL is a tool to tunnel TCP/UDP traffic between systems.

## Features

- Tunnel TCP/UDP traffic
- Reverse tunnels
- Tunnel between systems even if they're behind a NAT

## Usage

Download the binary for your respective `OS` and `ARCH` from the [Releases][releases] tab

| Argument     | Description           | Type                    | Default       |
| ------------ | --------------------- | ----------------------- | ------------- |
| mode         | Operation Mode        | broker, relay or client | broker        |
| brokerListen | Broker Listen Address | IP:PORT                 | :10000        |
| relayListen  | Relay Listen Address  | IP:PORT                 | :15000        |
| brokerAddr   | Broker Address        | IP:PORT                 | :10000        |
| relayAddr    | Relay Address         | IP:PORT                 | ""            |
| token        | Authentication Token  | String                  | ""            |
| peerId       | Peer to connect to    | String                  | ""            |
| peerMode     | Peer connect mode     | p2p or relay            | p2p           |
| tunType      | Tunnel Type           | tcp or udp              | tcp           |
| tunRev       | Reverse Tunnel        | boolean                 | false         |
| tunFrom      | Tunnel requests from  | IP:PORT                 | ""            |
| tunTo        | Tunnel requests to    | IP:PORT                 | ""            |
| configFile   | Config File           | File Path               | ./config.yaml |
| debug        | Debug Mode            | boolean                 | false         |

> Note: `sudo` is required in `client` mode to `PING` relay servers and choose the nearest one

## Example

1.  Run the `Broker` server

    ```sh
    xtunnel-os-arch --mode=broker --config="./example/config.yaml"
    ```

2.  Run the `Relay` server

    ```sh
    xtunnel-os-arch --mode=relay --token=token-r1 --broker=localhost:10000
    ```

3.  Connect `client-1`

    ```sh
    sudo xtunnel-os-arch --mode=client --token=token-c1 --broker=localhost:10000
    ```

4.  `Tunnel` from `client-2` to `client-1` using `P2P` mode and tunnel `TCP` traffic

    ```sh
    sudo xtunnel-os-arch --mode=client --token=token-c2 --broker=localhost:10000 --peerid=client-1 --peermode=p2p --tuntype=tcp --tunfrom=:8000 --tunto=192.168.1.100:22
    ```

5.  `Tunnel` from `client-3` to `client-1` using `Relay` mode and tunnel `TCP` traffic

    ```sh
    sudo xtunnel-os-arch --mode=client --token=token-c3 --broker=localhost:10000 --peerid=client-1 --peermode=relay --tuntype=tcp --tunfrom=:8100 --tunto=192.168.1.100:22
    ```

6.  `Reverse Tunnel` from `client-1` to `client-4` using `Relay` mode and tunnel `TCP` traffic

    ```sh
    sudo xtunnel-os-arch --mode=client --token=token-c4 --broker=localhost:10000 --peerid=client-3 --peermode=p2p --tuntype=tcp --tunrev=true --tunfrom=:9100 --tunto=192.168.1.100:22
    ```

## Build

This tool depends on [Make][toolmake] and [Protobuf][toolprotobuf] packages. Be sure to install them before building.

```sh
make all
```

## Packages

xTUNNEL depends on the following core packages

| Module             | Link            |
| ------------------ | --------------- |
| supergiant-hq/xnet | [View][pkgxnet] |

## TODO

- Multiple tunnels

## License

Apache License 2.0

[//]: # "Links"
[pkgxnet]: https://github.com/supergiant-hq/xnet
[releases]: https://github.com/supergiant-hq/xtunnel/releases
[toolprotobuf]: https://developers.google.com/protocol-buffers/
[toolmake]: https://www.gnu.org/software/make/
