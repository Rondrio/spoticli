package spotify

type Client struct {
	clientId     string
	clientSecret string
	accessToken  *accessToken
}

func NewClient(clientId string, clientSecret string) *Client {
	return &Client{
		clientId:     clientId,
		clientSecret: clientSecret,
	}
}
