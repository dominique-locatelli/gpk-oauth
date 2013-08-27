package oauth

import (
	"bytes"
	"ericaro.net/gopack/protocol"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func init() {
	// Special function called when the package is loaded:
	protocol.RegisterClient("oauth", NewOAuthClient) // Register oauth client
}

func NewOAuthClient(name string, u url.URL, token *protocol.Token) (c protocol.Client, err error) {

	// Remove 'oauth' prefix:
	urlString := u.String()
	urlString = strings.TrimLeft(urlString, "oauth:")
	newUrl, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	// Initialize OAuth token:
	oToken := &OAuthToken{}
	err = oToken.UnmarshalJSON(*token)
	if err != nil {
		return nil, err
	}

	c = &OAuthClient{
		name:       name,
		url:        *newUrl,
		oauthToken: oToken,
	}

	return c, nil
}

type OAuthClient struct {
	url        url.URL
	name       string
	oauthToken *OAuthToken
}

func (c *OAuthClient) Fetch(pid protocol.PID) (r io.ReadCloser, err error) {
	v := &url.Values{}
	pid.InParameter(v)

	u := &url.URL{
		Path:     protocol.FETCH,
		RawQuery: v.Encode(),
	}
	remote := c.Path()
	resp, err := doOAuthGET(remote.ResolveReference(u))
	if err != nil {
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	return resp.Body, nil
}

func (c *OAuthClient) Push(pid protocol.PID, r io.Reader) (err error) {
	v := &url.Values{}
	pid.InParameter(v)
	//query url
	u := &url.URL{
		Path:     protocol.PUSH,
		RawQuery: v.Encode(),
	}

	buf := new(bytes.Buffer)
	io.Copy(buf, r)

	var client http.Client
	remote := c.Path()
	req, err := http.NewRequest("POST", remote.ResolveReference(u).String(), buf)
	if err != nil {
		return
	}
	req.ContentLength = int64(buf.Len()) // fuck I can't do that, I need to compute the length first
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}
	return
}

func (c *OAuthClient) PushExecutables(pid protocol.PID, r io.Reader) (err error) {
	//[DL] TODO
	return
}

func (c *OAuthClient) Search(query string, start int) (result []protocol.PID) {
	//[DL] TODO
	return
}

func (c *OAuthClient) Name() string {
	return c.name
}

func (c *OAuthClient) Path() url.URL {
	return c.url
}

func (c *OAuthClient) Token() *protocol.Token {
	bytes, _ := c.oauthToken.MarshalJSON()
	gpkToken := protocol.Token(bytes)
	return &gpkToken
}

func doOAuthGET(url *url.URL) (resp *http.Response, err error) {
	//[DL] TODO
	return
}

func doOAuthPOST(url url.URL, buf *bytes.Buffer) (resp *http.Response, err error) {
	//[DL] TODO
	return
}
