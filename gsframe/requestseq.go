package gsframe

import (
	"context"

	"github.com/tanenking/gsframe/internal/constants"
)

func WithSeq(ctx context.Context, seq int32) context.Context {
	return constants.WithSeq(ctx, seq)
}

func ParseSeq(ctx context.Context) int32 {
	return constants.ParseSeq(ctx)
}

// ////////////////////////////////////////////////////////////////////////////////////////
func WithSeqForRpcContext(ctx context.Context, seq int32) context.Context {
	return constants.WithSeqForRpcContext(ctx, seq)
}

func ParseSeqFromRpcContext(ctx context.Context) int32 {
	return constants.ParseSeqFromRpcContext(ctx)
}
