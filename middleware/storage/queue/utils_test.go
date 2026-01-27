package queue

import (
	"errors"
	"testing"
)

func TestHandle(t *testing.T) {
	type TestMsg struct {
		Content string
	}

	tests := []struct {
		name    string
		handler Handler[TestMsg]
		msg     *Message[TestMsg]
		wantErr bool
		wantRsp ProcessRsp
	}{
		{
			name: "成功不重试",
			handler: func(msg *Message[TestMsg]) (ProcessRsp, error) {
				return ProcessRsp{Retry: false}, nil
			},
			msg:     &Message[TestMsg]{Body: TestMsg{Content: "test"}},
			wantErr: false,
			wantRsp: ProcessRsp{Retry: false},
		},
		{
			name: "成功需要重试",
			handler: func(msg *Message[TestMsg]) (ProcessRsp, error) {
				return ProcessRsp{Retry: true, SkipAttempts: true}, nil
			},
			msg:     &Message[TestMsg]{Body: TestMsg{Content: "test"}},
			wantErr: false,
			wantRsp: ProcessRsp{Retry: true, SkipAttempts: true},
		},
		{
			name: "处理错误",
			handler: func(msg *Message[TestMsg]) (ProcessRsp, error) {
				return ProcessRsp{}, errors.New("handler error")
			},
			msg:     &Message[TestMsg]{Body: TestMsg{Content: "test"}},
			wantErr: true,
		},
		{
			name: "Handler panic",
			handler: func(msg *Message[TestMsg]) (ProcessRsp, error) {
				panic("handler panic")
			},
			msg:     &Message[TestMsg]{Body: TestMsg{Content: "test"}},
			wantErr: false, // CachePanicWithHandle 不会返回错误，而是记录日志
		},
		{
			name: "Handler panic with int",
			handler: func(msg *Message[TestMsg]) (ProcessRsp, error) {
				panic(123)
			},
			msg:     &Message[TestMsg]{Body: TestMsg{Content: "test"}},
			wantErr: false, // CachePanicWithHandle 不会返回错误，而是记录日志
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := Handle(tt.handler, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && rsp.Retry != tt.wantRsp.Retry {
				t.Errorf("Handle() rsp.Retry = %v, want %v", rsp.Retry, tt.wantRsp.Retry)
			}
			if !tt.wantErr && rsp.SkipAttempts != tt.wantRsp.SkipAttempts {
				t.Errorf("Handle() rsp.SkipAttempts = %v, want %v", rsp.SkipAttempts, tt.wantRsp.SkipAttempts)
			}
		})
	}
}
