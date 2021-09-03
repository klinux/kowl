package connect

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudhut/common/rest"
	con "github.com/klinux/connect-client"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (s *Service) PutConnectorConfig(ctx context.Context, clusterName string, connectorName string, req con.PutConnectorConfigOptions) (con.ConnectorInfo, *rest.Error) {
	c, restErr := s.getConnectClusterByName(clusterName)
	if restErr != nil {
		return con.ConnectorInfo{}, restErr
	}

	cInfo, err := c.Client.PutConnectorConfig(ctx, connectorName, req)
	if err != nil {
		return con.ConnectorInfo{}, &rest.Error{
			Err:          fmt.Errorf("failed to patch connector config: %w", err),
			Status:       http.StatusOK,
			Message:      fmt.Sprintf("Failed to patch Connector config: %v", err.Error()),
			InternalLogs: []zapcore.Field{zap.String("cluster_name", clusterName), zap.String("connector_name", connectorName)},
			IsSilent:     false,
		}
	}

	return cInfo, nil
}
