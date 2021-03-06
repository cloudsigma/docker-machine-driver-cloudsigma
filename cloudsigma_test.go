package cloudsigma

import (
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

func TestCloudsigma_SetConfigFromFlags_emptyPassword(t *testing.T) {
	driver := NewDriver("default", "path")
	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"cloudsigma-username": "user@cloudsigma.com",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.Error(t, err)
}

func TestCloudsigma_SetConfigFromFlags_emptyUsername(t *testing.T) {
	driver := NewDriver("default", "path")
	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"cloudsigma-password": "password",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.Error(t, err)
}

func TestCloudsigma_SetConfigFromFlags(t *testing.T) {
	driver := NewDriver("default", "path")
	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"cloudsigma-password": "password",
			"cloudsigma-username": "user@cloudsigma.com",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
	assert.Equal(t, driver.ResolveStorePath("id_rsa"), driver.GetSSHKeyPath())
}

func TestCloudsigma_SetConfigFromFlags_CustomSSHUserAndPort(t *testing.T) {
	driver := NewDriver("default", "path")
	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"cloudsigma-password": "password",
			"cloudsigma-ssh-port": 2222,
			"cloudsigma-ssh-user": "ssh_user",
			"cloudsigma-username": "user@cloudsigma.com",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	sshPort, err := driver.GetSSHPort()
	assert.Equal(t, "ssh_user", driver.GetSSHUsername())
	assert.Equal(t, 2222, sshPort)
	assert.NoError(t, err)
}

func TestCloudsigma_SetConfigFromFlags_CustomServerParameter(t *testing.T) {
	driver := NewDriver("default", "path")
	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"cloudsigma-api-location": "wdc",
			"cloudsigma-cpu":          1500,
			"cloudsigma-drive-name":   "",
			"cloudsigma-drive-size":   15,
			"cloudsigma-drive-uuid":   "generated-uuid",
			"cloudsigma-memory":       512,
			"cloudsigma-password":     "password",
			"cloudsigma-static-ip":    "192.168.0.1",
			"cloudsigma-username":     "user@cloudsigma.com",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Equal(t, "wdc", driver.APILocation)
	assert.Equal(t, 1500, driver.CPU)
	assert.Equal(t, "", driver.DriveName)
	assert.Equal(t, 15, driver.DriveSize)
	assert.Equal(t, "generated-uuid", driver.DriveUUID)
	assert.Equal(t, 512, driver.Memory)
	assert.Equal(t, "192.168.0.1", driver.StaticIP)
}

func TestCloudsigma_SetConfigFromFlags_ExclusiveOptions(t *testing.T) {
	driver := NewDriver("default", "path")
	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"cloudsigma-drive-name": "debian",
			"cloudsigma-drive-uuid": "generated-uuid",
			"cloudsigma-password":   "password",
			"cloudsigma-username":   "user@cloudsigma.com",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.Error(t, err)
}

func TestCloudsigma_PreCreateCheck_InvalidIPAddress(t *testing.T) {
	driver := NewDriver("default", "path")
	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"cloudsigma-password":  "password",
			"cloudsigma-static-ip": "999.999.888.777",
			"cloudsigma-username":  "user@cloudsigma.com",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	_ = driver.SetConfigFromFlags(checkFlags)
	err := driver.PreCreateCheck()

	assert.Error(t, err)
}

func TestCloudsigma_DriverName(t *testing.T) {
	driver := NewDriver("default", "path")

	assert.Equal(t, "cloudsigma", driver.DriverName())
}
