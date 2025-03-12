Prysm Development Binaries
==========================

By default, this setup uses official Prysm images. However, you can also use custom-built Prysm binaries for development:

1. Compile the binaries and build the Docker images:
   
   ```bash
   ./build.sh
   ```

2. Run using the development configuration:

   ```bash
   kurtosis run github.com/ethpandaops/ethereum-package --args-file dev/prysm/prysm-dev.yaml
   ```
