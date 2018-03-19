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
			"cloudsigma-username": "user@cloudsigma.com",
			"cloudsigma-password": "password",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
	assert.Equal(t, driver.ResolveStorePath("id_rsa"), driver.GetSSHKeyPath())
}
