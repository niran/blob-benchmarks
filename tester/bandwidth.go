package tester

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/log"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/pkg/errors"
)

func InstallTcCommand(service *services.ServiceContext) error {
	log.Info("Updating apt cache")
	exit, logs, err := service.ExecCommand(strings.Split("apt update", " "))
	if err != nil {
		return errors.Wrap(err, "failed to update apt")
	}
	if exit != 0 {
		return fmt.Errorf("failed to update apt: %s", logs)
	}

	// Alpine Linux: iproute2-tc
	// Ubuntu: iproute2
	log.Info("Installing tc command")
	exit, logs, err = service.ExecCommand(strings.Split("apt install iproute2 -y", " "))
	if err != nil {
		return errors.Wrap(err, "failed to install tc command")
	}
	if exit != 0 {
		return fmt.Errorf("failed to install tc command: %s", logs)
	}

	return nil
}

func SetUploadBandwidthControl(service *services.ServiceContext, uploadBandwidth string) error {
	log.Info("Setting upload bandwidth control", "bandwidth", uploadBandwidth)
	tcCmd := fmt.Sprintf("tc qdisc add dev eth0 root tbf rate %s burst 16kb latency 50ms", uploadBandwidth)
	exit, logs, err := service.ExecCommand(strings.Split(tcCmd, " "))
	if err != nil {
		return errors.Wrap(err, "failed to create qdisc for upload bandwidth control")
	}
	if exit != 0 {
		return fmt.Errorf("failed to create qdisc for upload bandwidth control: %s", logs)
	}

	return nil
}

func SetDownloadBandwidthControl(service *services.ServiceContext, downloadBandwidth string) error {
	log.Info("Creating qdisc for download bandwidth control")
	exit, logs, err := service.ExecCommand(strings.Split("tc qdisc add dev eth0 handle ffff: ingress", " "))
	if err != nil {
		return errors.Wrap(err, "failed to create qdisc for download bandwidth control")
	}
	if exit != 0 {
		return fmt.Errorf("failed to create qdisc for download bandwidth control: %s", logs)
	}

	log.Info("Setting download bandwidth control", "bandwidth", downloadBandwidth)
	filterCmd := fmt.Sprintf("tc filter add dev eth0 parent ffff: protocol ip prio 1 u32 match ip src 0.0.0.0/0 police rate %s burst 16kb drop flowid :1", downloadBandwidth)
	exit, logs, err = service.ExecCommand(strings.Split(filterCmd, " "))
	if err != nil {
		return errors.Wrap(err, "failed to set download bandwidth control")
	}
	if exit != 0 {
		return fmt.Errorf("failed to set download bandwidth control: %s", logs)
	}

	return nil
}

func RemoveUploadBandwidthControl(service *services.ServiceContext) error {
	log.Info("Removing upload bandwidth control")
	exit, logs, err := service.ExecCommand(strings.Split("tc qdisc del dev eth0 root", " "))
	if err != nil {
		return errors.Wrap(err, "failed to remove upload bandwidth control")
	}
	if exit != 0 {
		return fmt.Errorf("failed to remove upload bandwidth control: %s", logs)
	}

	return nil
}

func RemoveDownloadBandwidthControl(service *services.ServiceContext) error {
	log.Info("Removing download bandwidth control")
	exit, logs, err := service.ExecCommand(strings.Split("tc qdisc del dev eth0 handle ffff: ingress", " "))
	if err != nil {
		return errors.Wrap(err, "failed to remove download bandwidth control")
	}
	if exit != 0 {
		return fmt.Errorf("failed to remove download bandwidth control: %s", logs)
	}

	return nil
}

func RemoveBandwidthControls(service *services.ServiceContext) error {
	uploadErr := RemoveUploadBandwidthControl(service)
	downloadErr := RemoveDownloadBandwidthControl(service)

	if uploadErr != nil && downloadErr != nil {
		return fmt.Errorf("failed to remove upload and download bandwidth controls: %s, %s", uploadErr, downloadErr)
	}
	if uploadErr != nil {
		return errors.Wrap(uploadErr, "failed to remove upload bandwidth control")
	}
	if downloadErr != nil {
		return errors.Wrap(downloadErr, "failed to remove download bandwidth control")
	}

	return nil
}
