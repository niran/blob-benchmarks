Blob Benchmarks
===============

This repository scripts benchmarks for blobs within a Kurtosis environment. It also contains utilities for running a development binary in this environment rather than just the container images from ethpandaops.

```bash
go run ./tester/cmd min-bandwidth
```

## Kurtosis Fork

Our network benchmarks need to be able to reduce the bandwidth available to nodes that have been launched by the `ethpandaops/ethereum-package` Kurtosis package. The minimally invasive way to do this is to maintain a ~one line fork of Kurtosis that adds the `NET_ADMIN` capability to each container launched as a user service (i.e. containers other than the Kurtosis engine containers).

To ensure that you're actually running on the fork, run `kurtosis engine stop` to completely stop Kurtosis, then run any `kurtosis` command using the forked binary. Kurtosis will launch an engine whenever there isn't a running engine, but it will reuse the running engine even if you're using a different binary. The changes we need to apply the `NET_ADMIN` capability run in the engine itself, not in the CLI.


### Building Kurtosis

#### How We Actually Did It

```bash
cd forks/kurtosis

# These commands seem to be the critical path to the containers we need.
./enclave-manager/scripts/build.sh 
./engine/scripts/build.sh
./core/scripts/build.sh

# We thought we needed to build the API as well, but it probably wasn't necessary.
nix-shell -p rustc
./api/scripts/build.sh
exit

# We also directly built containers during the process, but the build scripts should do this automatically.
./scripts/docker-image-builder.sh false engine/server/Dockerfile kurtosistech/engine:caa9a3-dirty
./scripts/docker-image-builder.sh false core/server/Dockerfile kurtosistech/core:caa9a3-dirty
```

#### What We Think Will Work

```bash
# The full build process fails because of a rustc version issue. Start a shell with the latest version of rustc.
nix-shell -p rustc

./scripts/build.sh
exit
```
