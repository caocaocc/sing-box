package rule

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/logger"
	"github.com/sagernet/sing/service"

	"github.com/stretchr/testify/require"
)

type remoteRuleSetTestHTTPClientManager struct {
	transport adapter.HTTPTransport
}

func (m *remoteRuleSetTestHTTPClientManager) ResolveTransport(context.Context, logger.ContextLogger, option.HTTPClientOptions) (adapter.HTTPTransport, error) {
	return m.transport, nil
}

func (m *remoteRuleSetTestHTTPClientManager) DefaultTransport() adapter.HTTPTransport {
	return m.transport
}

func (m *remoteRuleSetTestHTTPClientManager) ResetNetwork() {}

type remoteRuleSetTestTransport struct {
	access    sync.Mutex
	roundTrip func(*http.Request) (*http.Response, error)
}

func (t *remoteRuleSetTestTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	t.access.Lock()
	roundTrip := t.roundTrip
	t.access.Unlock()
	return roundTrip(request)
}

func (t *remoteRuleSetTestTransport) CloseIdleConnections() {}

func (t *remoteRuleSetTestTransport) Reset() {}

func (t *remoteRuleSetTestTransport) SetRoundTrip(roundTrip func(*http.Request) (*http.Response, error)) {
	t.access.Lock()
	t.roundTrip = roundTrip
	t.access.Unlock()
}

func TestRemoteRuleSetStartsAfterInitialFetchFailureAndUpdatesInPlace(t *testing.T) {
	t.Parallel()

	transport := &remoteRuleSetTestTransport{}
	transport.SetRoundTrip(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("blocked")
	})
	ctx := service.ContextWith[adapter.HTTPClientManager](context.Background(), &remoteRuleSetTestHTTPClientManager{
		transport: transport,
	})
	ruleSet, err := NewRemoteRuleSet(ctx, log.NewNOPFactory().NewLogger("router"), option.RuleSet{
		Type:   C.RuleSetTypeRemote,
		Tag:    "lazy-remote",
		Format: C.RuleSetFormatSource,
		RemoteOptions: option.RemoteRuleSet{
			URL: "https://example.com/rules.json",
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, ruleSet.Close())
	})

	startContext := adapter.NewHTTPStartContext()
	err = ruleSet.StartContext(ctx, startContext)
	startContext.Close()
	require.NoError(t, err)
	require.False(t, ruleSet.Match(&adapter.InboundContext{Domain: "example.com"}))

	ruleSet.IncRef()
	defer ruleSet.DecRef()
	transport.SetRoundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(`{"version":4,"rules":[{"domain":["example.com"]}]}`)),
			Header:     make(http.Header),
		}, nil
	})
	ruleSet.updateOnce()

	require.True(t, ruleSet.Match(&adapter.InboundContext{Domain: "example.com"}))
	require.False(t, ruleSet.Match(&adapter.InboundContext{Domain: "example.org"}))
}
