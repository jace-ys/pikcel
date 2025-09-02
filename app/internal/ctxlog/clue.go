package ctxlog

import (
	"context"
	"io"

	"goa.design/clue/log"
)

func NewContext(ctx context.Context, w io.Writer, debug bool) context.Context {
	format := log.FormatJSON
	if log.IsTerminal() {
		format = log.FormatTerminal
	}

	opts := []log.LogOption{
		log.WithOutput(w),
		log.WithFormat(format),
		log.WithFileLocation(),
		log.WithFunc(log.Span),
	}

	if debug {
		opts = append(opts, log.WithDebug())
	}

	return log.Context(ctx, opts...)
}

func KV(key string, val any) log.KV {
	return log.KV{K: key, V: val}
}

func With(ctx context.Context, kv ...log.Fielder) context.Context {
	return log.With(ctx, kv...)
}

func Print(ctx context.Context, msg string, kv ...log.Fielder) {
	log.Print(ctx, buildKVs(msg, kv...)...)
}

func Debug(ctx context.Context, msg string, kv ...log.Fielder) {
	log.Debug(ctx, buildKVs(msg, kv...)...)
}

func Info(ctx context.Context, msg string, kv ...log.Fielder) {
	log.Info(ctx, buildKVs(msg, kv...)...)
}

func Error(ctx context.Context, msg string, err error, kv ...log.Fielder) {
	log.Error(ctx, err, buildKVs(msg, kv...)...)
}

func Fatal(ctx context.Context, msg string, err error, kv ...log.Fielder) {
	log.Fatal(ctx, err, buildKVs(msg, kv...)...)
}

func buildKVs(msg string, kv ...log.Fielder) []log.Fielder {
	kvs := make([]log.Fielder, 0, len(kv)+1)
	kvs = append(kvs, KV(log.MessageKey, msg))
	kvs = append(kvs, kv...)
	return kvs
}
