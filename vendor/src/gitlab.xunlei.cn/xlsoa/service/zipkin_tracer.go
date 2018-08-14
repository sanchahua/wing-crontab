package service

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

var ZIPKIN_TRACER_HEADERS_TO_PROPAGATE = []string{
	"x-request-id",
	"x-ot-span-context",
	"x-b3-traceid",
	"x-b3-spanid",
	"x-b3-parentspanid",
	"x-b3-sampled",
	"x-b3-flags",
}

type zipkinTracer struct {
}

func NewZipkinTracer() *zipkinTracer {
	return &zipkinTracer{}
}

func (z *zipkinTracer) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	var meta = map[string]string{}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for _, v := range ZIPKIN_TRACER_HEADERS_TO_PROPAGATE {
			meta[v] = retrieveFromMeta(md, v)
		}
	}

	return meta, nil
}
