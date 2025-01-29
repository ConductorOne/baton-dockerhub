package connector

import (
	"context"
	"fmt"
	"io"

	"github.com/conductorone/baton-dockerhub/pkg/dockerhub"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type DockerHub struct {
	client *dockerhub.Client
	orgs   []string
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (dh *DockerHub) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		orgBuilder(dh.client, dh.orgs),
		repositoryBuilder(dh.client),
		userBuilder(dh.client),
		teamBuilder(dh.client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (dh *DockerHub) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (dh *DockerHub) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "DockerHub",
		Description: "Connector syncing DockerHub organizations, their members, teams, and repositories to Baton",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (dh *DockerHub) Validate(ctx context.Context) (annotations.Annotations, error) {
	// get the scope of used credentials
	_, _, err := dh.client.ListOrganizations(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("dockerhub-connector: validate: failed to list organizations: %w", err)
	}

	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, username, accessToken, password string, orgs []string) (*DockerHub, error) {
	l := ctxzap.Extract(ctx)
	l.Debug("creating client")
	hubClient, err := dockerhub.NewClient(ctx, username, password, accessToken)
	if err != nil {
		l.Error("error creating client", zap.Error(err))
		return nil, err
	}

	err = hubClient.SetCurrentUser(ctx, username)
	if err != nil {
		l.Error("error setting current user", zap.Error(err))
		return nil, err
	}

	return &DockerHub{
		client: hubClient,
		orgs:   orgs,
	}, nil
}
