# xTUNNEL

xTUNNEL is a tool to tunnel TCP/UDP traffic between systems.

## Features

- Tunnel TCP/UDP traffic
- Reverse tunnels
- Tunnel between systems even if they're behind a NAT

## Supports

| OS      | Arch         |
| ------- | ------------ |
| linux   | amd64, arm64 |
| darwin  | amd64, arm64 |
| windows | amd64        |

## Usage

Download the binary for your respective `OS` and `ARCH` from the [Releases][releases] tab

| Argument     | Description               | Type                    | Default       |
| ------------ | ------------------------- | ----------------------- | ------------- |
| mode         | Operation Mode            | broker, relay or client | broker        |
| brokerListen | Broker Listen Address     | IP:PORT                 | :10000        |
| relayListen  | Relay Listen Address      | IP:PORT                 | :15000        |
| brokerAddr   | Broker Address            | IP:PORT                 | :10000        |
| relayAddr    | Relay Address             | IP:PORT                 | ""            |
| token        | Authentication Token      | String                  | ""            |
| tunPeer      | Tunnel Peer to connect to | String                  | ""            |
| tunPeerMode  | Tunnel Peer connect mode  | p2p or relay            | p2p           |
| tunType      | Tunnel Type               | tcp or udp              | tcp           |
| tunRev       | Reverse Tunnel            | boolean                 | false         |
| tunFrom      | Tunnel requests from      | IP:PORT                 | ""            |
| tunTo        | Tunnel requests to        | IP:PORT                 | ""            |
| config       | Config File               | File Path               | ./config.yaml |
| debug        | Debug Mode                | boolean                 | false         |

> Note: `sudo` is required in `client` mode to `PING` relay servers and choose the nearest one

## Example

1.  Run the `Broker` server

    ```sh
    xtunnel-os-arch --mode=broker --config="./example/server.yaml"
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
    sudo xtunnel-os-arch --mode=client --token=token-c2 --broker=localhost:10000 --tunPeer=client-1 --tunPeerMode=p2p --tunType=tcp --tunFrom=:8000 --tunTo=192.168.1.100:22
    ```

5.  `Tunnel` from `client-3` to `client-1` using `Relay` mode and tunnel `TCP` traffic

    ```sh
    sudo xtunnel-os-arch --mode=client --token=token-c3 --broker=localhost:10000 --tunPeer=client-1 --tunPeerMode=relay --tunType=tcp --tunFrom=:8100 --tunTo=192.168.1.100:22
    ```

6.  `Reverse Tunnel` from `client-1` to `client-4` using `P2P` mode and tunnel `TCP` traffic

    ```sh
    sudo xtunnel-os-arch --mode=client --token=token-c4 --broker=localhost:10000 --tunPeer=client-3 --tunPeerMode=p2p --tunType=tcp --tunRev=true --tunFrom=:9000 --tunTo=192.168.1.100:22
    ```

7.  `Multiple Tunnels` from `client-5`

    ```sh
    sudo xtunnel-os-arch --mode=client --token=token-c5 --broker=localhost:10000 --config="./example/client-5.yaml"
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

## License

Apache License 2.0

[//]: # "Links"
[pkgxnet]: https://github.com/supergiant-hq/xnet
[releases]: https://github.com/supergiant-hq/xtunnel/releases
[toolprotobuf]: https://developers.google.com/protocol-buffers/
[toolmake]: https://www.gnu.org/software/make/
