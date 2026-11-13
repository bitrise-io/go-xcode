package appstoreconnectclient

type AuthClient struct {
}

func NewAuthClient() *AuthClient {
	return &AuthClient{}
}

func (c *AuthClient) Login() error {
	return nil
}
