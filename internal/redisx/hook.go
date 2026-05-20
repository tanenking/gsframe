package redisx

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
)

func (p Prefix) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (p Prefix) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		p.AssembleCMD(cmd)
		return next(ctx, cmd)
	}
}

func (p Prefix) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		for _, cmd := range cmds {
			p.AssembleCMD(cmd)
		}
		return next(ctx, cmds)
	}
}

var _ redis.Hook = Prefix("")

// AssembleCMD modifies the command arguments to add the prefix to keys.
func (p Prefix) AssembleCMD(cmd redis.Cmder) redis.Cmder {
	args := cmd.Args()
	lArgs := len(args)
	if lArgs < 2 {
		return cmd
	}

	switch c := strings.ToLower(cmd.Name()); c {
	case "mset", "msetnx":
		// MSET key1 value1 [key2 value2 ...] -- https://redis.io/commands/mset/
		// MSETNX key1 value1 [key2 value2 ...] -- https://redis.io/commands/msetnx/
		for i := 1; i < lArgs; i += 2 {
			if key, ok := args[i].(string); ok {
				args[i] = p.MakeKey(key)
			}
		}
	case "mget", "exists", "del", "unlink", "touch":
		// MGET key [key ...] -- https://redis.io/commands/mget/
		// EXISTS key [key ...] -- https://redis.io/commands/exists/
		// DEL key [key ...] -- https://redis.io/commands/del/
		// UNLINK key [key ...] -- https://redis.io/commands/unlink/
		// TOUCH key [key ...] -- https://redis.io/commands/touch/
		for i := 1; i < lArgs; i++ {
			if key, ok := args[i].(string); ok {
				args[i] = p.MakeKey(key)
			}
		}
	case "rename", "renamenx":
		// RENAME key newkey -- https://redis.io/commands/rename/
		// RENAMENX key newkey -- https://redis.io/commands/renamenx/
		if key, ok := args[1].(string); ok {
			args[1] = p.MakeKey(key)
		}
		if key, ok := args[2].(string); ok {
			args[2] = p.MakeKey(key)
		}
	case "rpoplpush", "brpoplpush":
		// RPOPLPUSH source destination -- https://redis.io/commands/rpoplpush/
		// BRPOPLPUSH source destination timeout -- https://redis.io/commands/brpoplpush/
		if key, ok := args[1].(string); ok {
			args[1] = p.MakeKey(key)
		}
		if key, ok := args[2].(string); ok {
			args[2] = p.MakeKey(key)
		}
	case "sdiffstore", "sinterstore", "sunionstore":
		// SDIFFSTORE destination key [key ...] -- https://redis.io/commands/sdiffstore/
		// SINTERSTORE destination key [key ...] -- https://redis.io/commands/sinterstore/
		// SUNIONSTORE destination key [key ...] -- https://redis.io/commands/sunionstore/
		for i := 1; i < lArgs; i++ {
			if key, ok := args[i].(string); ok {
				args[i] = p.MakeKey(key)
			}
		}
	case "zinterstore", "zunionstore":
		// ZINTERSTORE destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX] -- https://redis.io/commands/zinterstore/
		// ZUNIONSTORE destination numkeys key [key ...] [WEIGHTS weight [weight ...]] [AGGREGATE SUM|MIN|MAX] -- https://redis.io/commands/zunionstore/
		if key, ok := args[1].(string); ok {
			args[1] = p.MakeKey(key)
		}
		if numKeys, ok := args[2].(int64); ok {
			for i := 3; i < 3+int(numKeys) && i < lArgs; i++ {
				if key, ok := args[i].(string); ok {
					args[i] = p.MakeKey(key)
				}
			}
		}
	case "eval", "evalsha":
		// EVAL script numkeys key [key ...] arg [arg ...] -- https://redis.io/commands/eval/
		// EVALSHA sha1 numkeys key [key ...] arg [arg ...] -- https://redis.io/commands/evalsha/
		numKeys := 0
		if numKeysArg, ok := args[2].(int64); ok {
			numKeys = int(numKeysArg)
		} else if numKeysArg, ok := args[2].(int); ok {
			numKeys = numKeysArg
		}
		if numKeys > 0 {
			for i := 3; i < 3+numKeys && i < lArgs; i++ {
				if key, ok := args[i].(string); ok {
					args[i] = p.MakeKey(key)
				}
			}
		}
	case "script":
		// SCRIPT LOAD script  -- https://redis.io/commands/script-load/
		// SCRIPT EXISTS sha1 [sha1 ...] -- https://redis.io/commands/script-exists/
		// SCRIPT FLUSH -- https://redis.io/commands/script-flush/
		// SCRIPT KILL -- https://redis.io/commands/script-kill/
		// Do not modify keys for SCRIPT commands
	default:
		// Default case for commands with a single key as the second argument
		if lArgs > 1 {
			if key, ok := args[1].(string); ok {
				args[1] = p.MakeKey(key)
			}
		}
	}

	return cmd
}
