# mikrotik-speedtest

A small multi-arch image (amd64, arm64, arm/v7) for running the
[Ookla Speedtest CLI](https://www.speedtest.net/apps/cli) as a RouterOS container, so you can
measure the WAN from the router itself.

The image does not contain Ookla's binary, because its licence does not allow redistribution. On
first run the entrypoint downloads the right build from `install.speedtest.net` into
`/opt/speedtest`. RouterOS keeps the container's root directory between runs, so later runs reuse
it. Running the image means you accept Ookla's [EULA](https://www.speedtest.net/about/eula) and
[privacy policy](https://www.speedtest.net/about/privacy).

```
ghcr.io/tgenov/mikrotik-speedtest:latest
```

## Run it on RouterOS 7

Containers have to be enabled in device mode first. This needs physical access: run the command,
then press the reset button or power-cycle the router before the confirmation window closes.

```
/system/device-mode/update container=yes
```

Install the `container` package for your architecture and reboot. Then:

```
/interface/veth/add name=veth-st address=172.17.0.2/24 gateway=172.17.0.1
/interface/bridge/add name=containers
/ip/address/add address=172.17.0.1/24 interface=containers
/interface/bridge/port/add bridge=containers interface=veth-st
/ip/firewall/nat/add chain=srcnat src-address=172.17.0.0/24 action=masquerade

/container/config/set registry-url=https://ghcr.io tmpdir=tmp
/container/add remote-image=tgenov/mikrotik-speedtest:latest interface=veth-st \
    root-dir=speedtest logging=yes
```

Wait for the container to reach `stopped` (the image has been extracted), then start a test:

```
/container/start [find root-dir=speedtest]
/log/print where topics~"container"
```

The container prints the result to the log and exits. Start it again for each new test. To pass
CLI flags, such as a specific server or JSON output, set `cmd`:

```
/container/set [find root-dir=speedtest] cmd="--server-id=12345 --format=json"
```

## Reading the result

The test runs on the router's CPU, and its traffic goes through the router's NAT and firewall like
any other traffic. On a small router, especially over PPPoE, the CPU can top out below your line
rate. Watch `/system/resource/monitor` during a run: if the CPU is close to 100%, the number is the
router's limit, not your ISP's.

## Build locally

```
docker buildx build --platform linux/arm/v7 -t mikrotik-speedtest --load .
docker run --rm mikrotik-speedtest
```
