package utils

import (
	"context"
	"errors"

	"google.golang.org/api/idtoken"
)

type GoogleIdentityClaims struct {
	Subject string
	Email   string
	Name    string
	Picture string
}

func VerifyGoogleIDToken(ctx context.Context, token, audience string) (*GoogleIdentityClaims, error) {
	if token == "" {
		return nil, errors.New("missing google id token")
	}

	payload, err := idtoken.Validate(ctx, token, audience)
	if err != nil {
		return nil, err
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	if payload.Subject == "" || email == "" {
		return nil, errors.New("google id token missing required claims")
	}

	return &GoogleIdentityClaims{
		Subject: payload.Subject,
		Email:   email,
		Name:    name,
		Picture: picture,
	}, nil
}
