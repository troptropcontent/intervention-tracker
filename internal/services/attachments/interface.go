package attachments

import (
	"context"
	"io"
)

type Attacher interface {
	Attach(ctx context.Context, file io.ReadSeeker, fileName string, holder_id uint, holder_type string, kind string) error
	Delete(ctx context.Context, attachmentID uint) error
	DeleteByHolder(ctx context.Context, holderType string, holderID uint) error
}
