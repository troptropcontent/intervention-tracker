package attachments

import (
	"context"
	"io"
)

type Attacher interface {
	Attach(ctx context.Context, file io.ReadSeeker, fileName string, holder_id uint, holder_type string, kind string) error
}
