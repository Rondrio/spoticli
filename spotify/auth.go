package spotify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
)

const (
	authorisationEndpoint = "https://accounts.spotify.com/authorize?"
	accessTokenEndpoint   = "https://accounts.spotify.com/api/token"

	redirectURI  = "http://localhost:9999/callback"
	responseType = "code"
	state        = "state"
	grantType    = "authorization_code"

	// SCOPES

	//images
	SCOPE_UGC_IMAGE_UPLOAD = scope("ugc-image-upload ")

	// spoitfy connect
	SCOPE_USER_READ_PLAYBACK_STATE    = scope("user-read-playback-state ")
	SCOPE_USER_MODIFY_PLAYBACK_STATE  = scope("user-modify-playback-state ")
	SCOPE_USER_READ_CURRENTLY_PLAYING = scope("user-read-currently-playing ")

	// playback
	SCOPE_APP_REMOTE_CONTROL = scope("app-remote-control ")
	SCOPE_STREAMING          = scope("streaming ")

	// follow
	SCOPE_USER_FOLLOW_MODIFY = scope("user-follow-modify ")
	SCOPE_USER_FOLLOW_READ   = scope("user-follow-read ")

	// listening history
	SCOPE_READ_PLAYBACK_POSITION    = scope("user-read-playback-position ")
	SCOPE_USER_TOP_READ             = scope("user-top-read ")
	SCOPE_USER_READ_RECENTLY_PLAYED = scope("user-read-recently-played ")

	// library
	SCOPE_USER_LIBRARY_MODIFY = scope("user-library-modify ")
	SCOPE_USER_LIBRARY_READ   = scope("user-library-read ")

	// users
	SCOPE_USER_READ_EMAIL   = scope("user-read-email ")
	SCOPE_USER_READ_PRIVATE = scope("user-read-private ")

	// open access
	SCOPE_USER_SOA_LINK            = scope("user-soa-link ")
	SCOPE_USER_SOA_UNLINK          = scope("user-soa-unlink ")
	SCOPE_USER_MANAGE_ENTITLEMENTS = scope("user-manage-entitlements ")
	SCOPE_USER_MANAGE_PARTNER      = scope("user-manage-partner ")
	SCOPE_USER_CREATE_PARTNER      = scope("user-create-partner ")
)

type scope string

type accessToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func (cl *Client) Authorize(sc scope) error {
	// build query
	url, err := url.Parse(authorisationEndpoint)
	if err != nil {
		return err
	}

	query := url.Query()
	query.Add("response_type", responseType)
	query.Add("client_id", cl.clientId)
	query.Add("scope", string(sc))
	query.Add("redirect_uri", redirectURI)
	query.Add("state", state)

	url.RawQuery = query.Encode()

	// start local http server to handle redirect

	srv := &http.Server{Addr: ":9999"}
	callback := callbackServer{
		returnChan: make(chan string),
	}

	http.HandleFunc("/callback", callback.handleFunc)
	go srv.ListenAndServe()

	// open link in browser
	open(url.String())

	authCode := <-callback.returnChan
	if authCode == "" {
		return errors.New("authentication failed")
	}

	srv.Shutdown(context.Background())

	// request access token
	cl.accessToken, err = cl.requestAccessToken(authCode)
	if err != nil {
		return err
	}

	return err
}

// requestAccessToken uses the authCode to make a request to the api/token endpoint and receive an accessToken
func (c *Client) requestAccessToken(authCode string) (*accessToken, error) {

	var token accessToken

	req, err := http.NewRequest("POST", accessTokenEndpoint, nil)
	if err != nil {
		return nil, err
	}

	// add params
	query := req.URL.Query()
	query.Add("grant_type", grantType)
	query.Add("code", authCode)
	query.Add("redirect_uri", redirectURI)
	req.URL.RawQuery = query.Encode()

	clientCreds := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.clientId, c.clientSecret)))

	// add headers
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", clientCreds))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// callbackServer is a local server to receive the authorization code
type callbackServer struct {
	returnChan chan string
}

func (cs callbackServer) handleFunc(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if query.Get("error") != "" {
		cs.returnChan <- ""
		return
	}

	cs.returnChan <- query.Get("code")
}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
