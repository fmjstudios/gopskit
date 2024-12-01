package keycloak

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
)

func (kc *Client) CreateRealm(realm *gocloak.RealmRepresentation) error {
	_, err := kc.api.CreateRealm(context.Background(), kc.auth.JWT.AccessToken, *realm)
	if err != nil {
		handled := kc.handleAuthorizationError(err)
		if handled != nil {
			return handled
		}

		return err
	}

	return nil
}

func (kc *Client) CreateGroup(group gocloak.Group) error {
	_, err := kc.api.CreateGroup(context.Background(), kc.auth.JWT.AccessToken, kc.realm, group)
	if err != nil {
		handled := kc.handleAuthorizationError(err)
		if handled != nil {
			return handled
		}

		return err
	}

	return nil
}

func (kc *Client) CreateClient(client gocloak.Client) error {
	_, err := kc.api.CreateClient(context.Background(), kc.auth.JWT.AccessToken, kc.realm, client)
	if err != nil {
		handled := kc.handleAuthorizationError(err)
		if handled != nil {
			return handled
		}

		return err
	}

	return nil
}
