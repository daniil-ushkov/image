//go:build !containers_image_ostree || !linux
// +build !containers_image_ostree !linux

package alltransports

import "github.com/daniil-ushkov/image/v5/transports"

func init() {
	transports.Register(transports.NewStubTransport("ostree"))
}
