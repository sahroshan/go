package pubnub

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pubnub/go/utils"
)

const REMOVE_CHANNEL_CHANNEL_GROUP = "/v1/channel-registration/sub-key/%s/channel-group/%s"

var emptyRemoveChannelFromChannelGroupResponse *RemoveChannelFromChannelGroupResponse

type RemoveChannelFromChannelGroupBuilder struct {
	opts *removeChannelOpts
}

func newRemoveChannelFromChannelGroupBuilder(
	pubnub *PubNub) *RemoveChannelFromChannelGroupBuilder {
	builder := RemoveChannelFromChannelGroupBuilder{
		opts: &removeChannelOpts{
			pubnub: pubnub,
		},
	}

	return &builder
}

func newRemoveChannelFromChannelGroupBuilderWithContext(
	pubnub *PubNub, context Context) *RemoveChannelFromChannelGroupBuilder {
	builder := RemoveChannelFromChannelGroupBuilder{
		opts: &removeChannelOpts{
			pubnub: pubnub,
			ctx:    context,
		},
	}

	return &builder
}

func (b *RemoveChannelFromChannelGroupBuilder) Channels(
	ch []string) *RemoveChannelFromChannelGroupBuilder {
	b.opts.Channels = ch
	return b
}

func (b *RemoveChannelFromChannelGroupBuilder) Group(
	cg string) *RemoveChannelFromChannelGroupBuilder {
	b.opts.Group = cg
	return b
}

func (b *RemoveChannelFromChannelGroupBuilder) Execute() (
	*RemoveChannelFromChannelGroupResponse, StatusResponse, error) {
	rawJson, status, err := executeRequest(b.opts)
	if err != nil {
		return emptyRemoveChannelFromChannelGroupResponse, status, err
	}

	return newRemoveChannelFromChannelGroupResponse(rawJson, status)
}

type removeChannelOpts struct {
	pubnub *PubNub

	Channels []string

	Group string

	Transport http.RoundTripper

	ctx Context
}

func (o *removeChannelOpts) config() Config {
	return *o.pubnub.Config
}

func (o *removeChannelOpts) client() *http.Client {
	return o.pubnub.GetClient()
}

func (o *removeChannelOpts) context() Context {
	return o.ctx
}

func (o *removeChannelOpts) validate() error {
	if o.config().SubscribeKey == "" {
		return newValidationError(o, StrMissingSubKey)
	}

	if len(o.Channels) == 0 {
		return newValidationError(o, StrMissingChannel)
	}

	if o.Group == "" {
		return newValidationError(o, StrMissingChannelGroup)
	}

	return nil
}

func (o *removeChannelOpts) buildPath() (string, error) {
	return fmt.Sprintf(REMOVE_CHANNEL_CHANNEL_GROUP,
		o.pubnub.Config.SubscribeKey,
		utils.UrlEncode(o.Group)), nil
}

func (o *removeChannelOpts) buildQuery() (*url.Values, error) {
	q := defaultQuery(
		utils.UrlEncode(o.pubnub.Config.Uuid))

	var channels []string

	for _, ch := range o.Channels {
		channels = append(channels, utils.UrlEncode(ch))
	}

	q.Set("remove", strings.Join(channels, ","))

	return q, nil
}

func (o *removeChannelOpts) buildBody() ([]byte, error) {
	return []byte{}, nil
}

func (o *removeChannelOpts) httpMethod() string {
	return "GET"
}

func (o *removeChannelOpts) isAuthRequired() bool {
	return true
}

func (o *removeChannelOpts) requestTimeout() int {
	return o.pubnub.Config.NonSubscribeRequestTimeout
}

func (o *removeChannelOpts) connectTimeout() int {
	return o.pubnub.Config.ConnectTimeout
}

func (o *removeChannelOpts) operationType() OperationType {
	return PNRemoveChannelFromChannelGroupOperation
}

type RemoveChannelFromChannelGroupResponse struct {
}

func newRemoveChannelFromChannelGroupResponse(jsonBytes []byte,
	status StatusResponse) (*RemoveChannelFromChannelGroupResponse,
	StatusResponse, error) {
	return emptyRemoveChannelFromChannelGroupResponse, status, nil
}