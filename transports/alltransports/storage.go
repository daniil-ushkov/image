//go:build !containers_image_storage_stub
// +build !containers_image_storage_stub

package alltransports

import (
	// Register the storage transport
	_ "github.com/daniil-ushkov/image/v5/storage"
)
