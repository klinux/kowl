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

func (s *Service) CreateConnector(ctx context.Context, clusterName string, req con.CreateConnectorRequest) (con.ConnectorInfo, *rest.Error) {
	c, restErr := s.getConnectClusterByName(clusterName)
	if restErr != nil {
		return con.ConnectorInfo{}, restErr
	}

	cInfo, err := c.Client.CreateConnector(ctx, req)
	if err != nil {
		return con.ConnectorInfo{}, &rest.Error{
			Err:          fmt.Errorf("failed to create connector: %w", err),
			Status:       http.StatusOK,
			Message:      fmt.Sprintf("Failed to create Connector: %v", err.Error()),
			InternalLogs: []zapcore.Field{zap.String("cluster_name", clusterName)},
			IsSilent:     false,
		}
	}

	return cInfo, nil
}
