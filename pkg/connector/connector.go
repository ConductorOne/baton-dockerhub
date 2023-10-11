package connector

import (
	"context"
	"fmt"
	"io"

	"github.com/conductorone/baton-dockerhub/pkg/dockerhub"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
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
		Description: "Connector syncing DockerHub organization members their teams, and their roles to Baton",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (dh *DockerHub) Validate(ctx context.Context) (annotations.Annotations, error) {
	// get the scope of used credentials
	err := dh.client.GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("dockerhub-connector: failed to get current user: %w", err)
	}

	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, username, password string, orgs []string) (*DockerHub, error) {
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, ctxzap.Extract(ctx)))
	if err != nil {
		return nil, err
	}

	hubClient, err := dockerhub.NewClient(ctx, httpClient, username, password)
	if err != nil {
		return nil, err
	}

	err = hubClient.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	return &DockerHub{
		client: hubClient,
		orgs:   orgs,
	}, nil
}
