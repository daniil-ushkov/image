package daemon

import "github.com/daniil-ushkov/image/v5/internal/private"

var _ private.ImageDestination = (*daemonImageDestination)(nil)
