package attachments

import (
	"context"
	"os"
)

type Attacher interface {
	Attach(ctx context.Context, file *os.File, record_id uint, record_type string, contentType string) error
}
