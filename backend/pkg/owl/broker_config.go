package owl

import (
	"context"
	"fmt"
	"github.com/cloudhut/common/rest"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kmsg"
	"net/http"
)

type BrokerConfigs struct {
	BrokerID int32               `json:"brokerId"`
	Configs  []BrokerConfigEntry `json:"configs"`
}

type BrokerConfigEntry struct {
	Name   string  `json:"name"`
	Value  *string `json:"value"` // If value is sensitive this will be nil
	Source string  `json:"source"`
	Type   string  `json:"type"`
	// IsExplicitlySet indicates whether this config's value was explicitly configured. It could still be the default value.
	IsExplicitlySet bool                  `json:"isExplicitlySet"`
	IsDefaultValue  bool                  `json:"isDefaultValue"`
	IsReadOnly      bool                  `json:"isReadOnly"`
	IsSensitive     bool                  `json:"isSensitive"`
	Documentation   *string               `json:"documentation"` // Will be nil for Kafka <v2.6.0
	Synonyms        []BrokerConfigSynonym `json:"synonyms"`
}

type BrokerConfigSynonym struct {
	Name   string  `json:"name"`
	Value  *string `json:"value"`
	Source string  `json:"source"`
}

func (s *Service) GetBrokerConfig(ctx context.Context, brokerID int32) ([]BrokerConfigEntry, *rest.Error) {
	res, err := s.kafkaSvc.DescribeBrokerConfig(ctx, brokerID, nil)
	if err != nil {
		return nil, &rest.Error{
			Err:      fmt.Errorf("failed to request broker config: %w", err),
			Status:   http.StatusServiceUnavailable,
			Message:  fmt.Sprintf("Failed to request broker's config: %v", err.Error()),
			IsSilent: false,
		}
	}

	// Resources should always be of length = 1
	for _, resource := range res.Resources {
		err := kerr.TypedErrorForCode(resource.ErrorCode)
		if err != nil {
			return nil, &rest.Error{
				Err:      err,
				Status:   http.StatusServiceUnavailable,
				Message:  fmt.Sprintf("Failed to describe broker config resource: %v", err.Error()),
				IsSilent: false,
			}
		}

		configEntries := make([]BrokerConfigEntry, len(resource.Configs))
		for j, cfg := range resource.Configs {
			isDefaultValue := false
			innerEntries := make([]BrokerConfigSynonym, len(cfg.ConfigSynonyms))
			for j, innerCfg := range cfg.ConfigSynonyms {
				innerEntries[j] = BrokerConfigSynonym{
					Name:   innerCfg.Name,
					Value:  innerCfg.Value,
					Source: innerCfg.Source.String(),
				}
				if innerCfg.Source == kmsg.ConfigSourceDefaultConfig {
					isDefaultValue = derefString(cfg.Value) == derefString(innerCfg.Value)
				}
			}

			isExplicitlySet := !cfg.IsDefault ||
				cfg.Source == kmsg.ConfigSourceStaticBrokerConfig ||
				cfg.Source == kmsg.ConfigSourceDynamicBrokerConfig
			configEntries[j] = BrokerConfigEntry{
				Name:            cfg.Name,
				Value:           cfg.Value,
				Source:          cfg.Source.String(),
				Type:            cfg.ConfigType.String(),
				IsExplicitlySet: isExplicitlySet,
				IsDefaultValue:  isDefaultValue,
				IsReadOnly:      cfg.ReadOnly,
				IsSensitive:     cfg.IsSensitive,
				Documentation:   cfg.Documentation,
				Synonyms:        innerEntries,
			}
		}

		return configEntries, nil
	}

	return nil, &rest.Error{
		Err:      fmt.Errorf("broker describe config response was empty"),
		Status:   http.StatusInternalServerError,
		Message:  "Broker config response was empty",
		IsSilent: false,
	}
}
