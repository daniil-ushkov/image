package archive

import (
	"context"
	"fmt"
	"io"

	"github.com/daniil-ushkov/image/v5/docker/internal/tarfile"
	"github.com/daniil-ushkov/image/v5/internal/private"
	"github.com/daniil-ushkov/image/v5/types"
)

type archiveImageDestination struct {
	*tarfile.Destination // Implements most of types.ImageDestination
	ref                  archiveReference
	archive              *tarfile.Writer // Should only be closed if writer != nil
	writer               io.Closer       // May be nil if the archive is shared
}

func newImageDestination(sys *types.SystemContext, ref archiveReference) (private.ImageDestination, error) {
	if ref.sourceIndex != -1 {
		return nil, fmt.Errorf("Destination reference must not contain a manifest index @%d", ref.sourceIndex)
	}

	var archive *tarfile.Writer
	var writer io.Closer
	if ref.archiveWriter != nil {
		archive = ref.archiveWriter
		writer = nil
	} else {
		fh, err := openArchiveForWriting(ref.path)
		if err != nil {
			return nil, err
		}

		archive = tarfile.NewWriter(fh)
		writer = fh
	}
	tarDest := tarfile.NewDestination(sys, archive, ref.Transport().Name(), ref.ref)
	if sys != nil && sys.DockerArchiveAdditionalTags != nil {
		tarDest.AddRepoTags(sys.DockerArchiveAdditionalTags)
	}
	return &archiveImageDestination{
		Destination: tarDest,
		ref:         ref,
		archive:     archive,
		writer:      writer,
	}, nil
}

// Reference returns the reference used to set up this destination.  Note that this should directly correspond to user's intent,
// e.g. it should use the public hostname instead of the result of resolving CNAMEs or following redirects.
func (d *archiveImageDestination) Reference() types.ImageReference {
	return d.ref
}

// Close removes resources associated with an initialized ImageDestination, if any.
func (d *archiveImageDestination) Close() error {
	if d.writer != nil {
		return d.writer.Close()
	}
	return nil
}

// Commit marks the process of storing the image as successful and asks for the image to be persisted.
// unparsedToplevel contains data about the top-level manifest of the source (which may be a single-arch image or a manifest list
// if PutManifest was only called for the single-arch image with instanceDigest == nil), primarily to allow lookups by the
// original manifest list digest, if desired.
// WARNING: This does not have any transactional semantics:
// - Uploaded data MAY be visible to others before Commit() is called
// - Uploaded data MAY be removed or MAY remain around if Close() is called without Commit() (i.e. rollback is allowed but not guaranteed)
func (d *archiveImageDestination) Commit(ctx context.Context, unparsedToplevel types.UnparsedImage) error {
	if d.writer != nil {
		return d.archive.Close()
	}
	return nil
}
