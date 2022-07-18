//go:build containers_image_ostree && linux
// +build containers_image_ostree,linux

package alltransports

import (
	// Register the ostree transport
	_ "github.com/daniil-ushkov/image/v5/ostree"
)
