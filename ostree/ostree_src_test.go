//go:build containers_image_ostree
// +build containers_image_ostree

package ostree

import "github.com/daniil-ushkov/image/v5/internal/private"

var _ private.ImageSource = (*ostreeImageSource)(nil)
