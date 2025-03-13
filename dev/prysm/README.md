Prysm Development Binaries
==========================

By default, this setup uses official Prysm images. However, you can also use custom-built Prysm binaries for development:

1. Compile the binaries and build the Docker images:

  Prysm's build process generates minimal containers suitable for production environments. For development purposes, we build general purpose containers instead. The built images are named `prysm-beacon-chain-dev` and `prysm-validator-dev`.
   
   ```bash
   ./build.sh
   ```

2. Run using the development configuration:

   ```bash
   ./kurtosis_run_args.sh kurtosis/fusaka/000-base.yaml dev/prysm/prysm-dev.yaml
   ```
