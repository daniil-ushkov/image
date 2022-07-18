package none

import (
	"github.com/daniil-ushkov/image/v5/types"
)

var _ types.BlobInfoCache = &noCache{}
