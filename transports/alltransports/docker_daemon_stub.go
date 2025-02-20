//go:build containers_image_docker_daemon_stub
// +build containers_image_docker_daemon_stub

package alltransports

import "github.com/daniil-ushkov/image/v5/transports"

func init() {
	transports.Register(transports.NewStubTransport("docker-daemon"))
}
