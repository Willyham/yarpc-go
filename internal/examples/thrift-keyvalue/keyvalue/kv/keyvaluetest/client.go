// Code generated by thriftrw-plugin-yarpc
// @generated

package keyvaluetest

import (
	"context"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/internal/examples/thrift-keyvalue/keyvalue/kv/keyvalueclient"
	"github.com/golang/mock/gomock"
)

// MockClient implements a gomock-compatible mock client for service
// KeyValue.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *_MockClientRecorder
}

var _ keyvalueclient.Interface = (*MockClient)(nil)

type _MockClientRecorder struct {
	mock *MockClient
}

// Build a new mock client for service KeyValue.
//
// 	mockCtrl := gomock.NewController(t)
// 	client := keyvaluetest.NewMockClient(mockCtrl)
//
// Use EXPECT() to set expectations on the mock.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &_MockClientRecorder{mock}
	return mock
}

// EXPECT returns an object that allows you to define an expectation on the
// KeyValue mock client.
func (m *MockClient) EXPECT() *_MockClientRecorder {
	return m.recorder
}

// GetValue responds to a GetValue call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().GetValue(gomock.Any(), ...).Return(...)
// 	... := client.GetValue(...)
func (m *MockClient) GetValue(
	ctx context.Context,
	_Key *string,
	opts ...yarpc.CallOption,
) (success string, err error) {

	args := []interface{}{ctx, _Key}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "GetValue", args...)
	success, _ = ret[i].(string)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) GetValue(
	ctx interface{},
	_Key interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Key}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "GetValue", args...)
}

// SetValue responds to a SetValue call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().SetValue(gomock.Any(), ...).Return(...)
// 	... := client.SetValue(...)
func (m *MockClient) SetValue(
	ctx context.Context,
	_Key *string,
	_Value *string,
	opts ...yarpc.CallOption,
) (err error) {

	args := []interface{}{ctx, _Key, _Value}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "SetValue", args...)
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) SetValue(
	ctx interface{},
	_Key interface{},
	_Value interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Key, _Value}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "SetValue", args...)
}
