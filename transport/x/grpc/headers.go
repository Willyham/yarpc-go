// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package grpc

import (
	"fmt"

	"go.uber.org/yarpc/api/transport"

	"go.uber.org/multierr"
	"google.golang.org/grpc/metadata"
)

// these are the same as in transport/http but lowercase
// http2 does all lowercase headers and this should be explicit

const (
	// CallerHeader is the header key for the caller.
	CallerHeader = "rpc-caller"
	// ServiceHeader is the header key for the service.
	ServiceHeader = "rpc-service"
	// ShardKeyHeader is the header key for the shard key.
	ShardKeyHeader = "rpc-shard-key"
	// RoutingKeyHeader is the header key for the routing key.
	RoutingKeyHeader = "rpc-routing-key"
	// RoutingDelegateHeader is the header key for the routing delegate.
	RoutingDelegateHeader = "rpc-routing-delegate"
	// EncodingHeader is the header key for the encoding.
	// This will be removed when we get encoding propagated using content-type.
	EncodingHeader = "rpc-encoding"
)

var (
	reservedHeaders = map[string]bool{
		CallerHeader:          true,
		ServiceHeader:         true,
		ShardKeyHeader:        true,
		RoutingKeyHeader:      true,
		RoutingDelegateHeader: true,
		EncodingHeader:        true,
	}
)

// transportRequestToMetadata will populate all reserved and application headers
// from the Request into a new MD.
func transportRequestToMetadata(request *transport.Request) (metadata.MD, error) {
	md := metadata.New(nil)
	if err := multierr.Combine(
		addToMetadata(md, CallerHeader, request.Caller),
		addToMetadata(md, ServiceHeader, request.Service),
		addToMetadata(md, ShardKeyHeader, request.ShardKey),
		addToMetadata(md, RoutingKeyHeader, request.RoutingKey),
		addToMetadata(md, RoutingDelegateHeader, request.RoutingDelegate),
		addToMetadata(md, EncodingHeader, string(request.Encoding)),
	); err != nil {
		return md, err
	}
	return md, addApplicationHeaders(md, request.Headers)
}

// metadataToTransportRequest will populate the Request with all reserved and application
// headers into a new Request, only not setting the Body field.
func metadataToTransportRequest(md metadata.MD) (*transport.Request, error) {
	request := &transport.Request{
		// NewHeadersWithCapacity handles if capacity <= 0
		Headers: transport.NewHeadersWithCapacity(md.Len() - len(reservedHeaders)),
	}
	for header := range md {
		header = transport.CanonicalizeHeaderKey(header)
		value, err := getFromMetadata(md, header)
		if err != nil {
			return nil, err
		}
		switch header {
		case CallerHeader:
			request.Caller = value
		case ServiceHeader:
			request.Service = value
		case ShardKeyHeader:
			request.ShardKey = value
		case RoutingKeyHeader:
			request.RoutingKey = value
		case RoutingDelegateHeader:
			request.RoutingDelegate = value
		case EncodingHeader:
			request.Encoding = transport.Encoding(value)
		default:
			request.Headers = request.Headers.With(header, value)
		}
	}
	return request, nil
}

// addApplicationHeaders adds the headers to md.
func addApplicationHeaders(md metadata.MD, headers transport.Headers) error {
	for header, value := range headers.Items() {
		header = transport.CanonicalizeHeaderKey(header)
		if _, ok := reservedHeaders[header]; ok {
			return fmt.Errorf("cannot use reserved header in application headers: %s", header)
		}
		if err := addToMetadata(md, header, value); err != nil {
			return err
		}
	}
	return nil
}

// getApplicationHeaders returns the headers from md without any reserved headers.
func getApplicationHeaders(md metadata.MD) (transport.Headers, error) {
	headers := transport.NewHeadersWithCapacity(md.Len())
	for header := range md {
		header = transport.CanonicalizeHeaderKey(header)
		if _, ok := reservedHeaders[header]; ok {
			continue
		}
		value, err := getFromMetadata(md, header)
		if err != nil {
			return headers, err
		}
		headers = headers.With(header, value)
	}
	return headers, nil
}

// add to md
// return error if key already in md
func addToMetadata(md metadata.MD, key string, value string) error {
	if value == "" {
		return nil
	}
	if _, ok := md[key]; ok {
		return fmt.Errorf("duplicate key: %s", key)
	}
	md[key] = []string{value}
	return nil
}

// get from md
// return "" if not present
// return error if more than one value
func getFromMetadata(md metadata.MD, key string) (string, error) {
	values, ok := md[key]
	if !ok {
		return "", nil
	}
	switch len(values) {
	case 0:
		return "", nil
	case 1:
		return values[0], nil
	default:
		return "", fmt.Errorf("key has more than one value: %s", key)
	}
}
