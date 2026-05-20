package constants

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tanenking/gsframe/gsinf"
	"google.golang.org/grpc/metadata"
)

const RequestSeqKey gsinf.ContextKey = "requestseq"

func WithSeq(ctx context.Context, seq int32) context.Context {
	return context.WithValue(ctx, RequestSeqKey, seq)
}

func ParseSeq(ctx context.Context) int32 {
	r := ctx.Value(RequestSeqKey)
	if r == nil {
		return 0
	}
	seq, ok := r.(int32)
	if !ok {
		return 0
	}
	return seq
}

// ////////////////////////////////////////////////////////////////////////////////////////
func WithSeqForRpcContext(ctx context.Context, seq int32) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		string(RequestSeqKey): fmt.Sprintf("%v", seq),
	}))
}

func ParseSeqFromRpcContext(ctx context.Context) int32 {
	key := string(RequestSeqKey)
	if md, ok := metadata.FromIncomingContext(ctx); ok && md != nil {
		if len(md[key]) > 0 {
			_seq := md[key][0]
			seq, err := strconv.Atoi(_seq)
			if err != nil {
				fmt.Printf("ParseCtxRequestRpcNo => %+v", err)
				return 0
			}
			return int32(seq)
		}
	}
	return 0
}
