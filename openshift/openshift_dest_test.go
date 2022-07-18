package openshift

import "github.com/daniil-ushkov/image/v5/internal/private"

var _ private.ImageDestination = (*openshiftImageDestination)(nil)
