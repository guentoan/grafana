package accesscontrol_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/accesscontrol"
	"github.com/grafana/grafana/pkg/services/sqlstore"
)

type filterDatasourcesTestCase struct {
	desc        string
	sqlID       string
	prefix      string
	attribute   string
	actions     []string
	permissions map[string][]string

	expectedDataSources []string
	expectErr           bool
}

func TestFilter_Datasources(t *testing.T) {
	tests := []filterDatasourcesTestCase{
		{
			desc:      "expect all data sources to be returned",
			sqlID:     "data_source.id",
			prefix:    "datasources",
			attribute: accesscontrol.ScopeAttributeID,
			actions:   []string{"datasources:read"},
			permissions: map[string][]string{
				"datasources:read": {"datasources:*"},
			},
			expectedDataSources: []string{"ds:1", "ds:2", "ds:3", "ds:4", "ds:5", "ds:6", "ds:7", "ds:8", "ds:9", "ds:10"},
		},
		{
			desc:      "expect all data sources for wildcard id scope to be returned",
			sqlID:     "data_source.id",
			prefix:    "datasources",
			attribute: accesscontrol.ScopeAttributeID,
			actions:   []string{"datasources:read"},
			permissions: map[string][]string{
				"datasources:read": {"datasources:id:*"},
			},
			expectedDataSources: []string{"ds:1", "ds:2", "ds:3", "ds:4", "ds:5", "ds:6", "ds:7", "ds:8", "ds:9", "ds:10"},
		},
		{
			desc:      "expect all data sources for wildcard scope to be returned",
			sqlID:     "data_source.id",
			prefix:    "datasources",
			attribute: accesscontrol.ScopeAttributeID,
			actions:   []string{"datasources:read"},
			permissions: map[string][]string{
				"datasources:read": {"*"},
			},
			expectedDataSources: []string{"ds:1", "ds:2", "ds:3", "ds:4", "ds:5", "ds:6", "ds:7", "ds:8", "ds:9", "ds:10"},
		},
		{
			desc:                "expect no data sources to be returned",
			sqlID:               "data_source.id",
			prefix:              "datasources",
			attribute:           accesscontrol.ScopeAttributeID,
			actions:             []string{"datasources:read"},
			permissions:         map[string][]string{},
			expectedDataSources: []string{},
		},
		{
			desc:      "expect data sources with id 3, 7 and 8 to be returned",
			sqlID:     "data_source.id",
			prefix:    "datasources",
			attribute: accesscontrol.ScopeAttributeID,
			actions:   []string{"datasources:read"},
			permissions: map[string][]string{
				"datasources:read": {"datasources:id:3", "datasources:id:7", "datasources:id:8"},
			},
			expectedDataSources: []string{"ds:3", "ds:7", "ds:8"},
		},
		{
			desc:    "expect no data sources to be returned for malformed scope",
			sqlID:   "data_source.id",
			prefix:  "datasources",
			actions: []string{"datasources:read"},
			permissions: map[string][]string{
				"datasources:read": {"datasources:id:1*"},
			},
		},
		{
			desc:      "expect error if sqlID is not in the accept list",
			sqlID:     "other.id",
			prefix:    "datasources",
			attribute: accesscontrol.ScopeAttributeID,
			actions:   []string{"datasources:read"},
			permissions: map[string][]string{
				"datasources:read": {"datasources:id:3", "datasources:id:7", "datasources:id:8"},
			},
			expectErr: true,
		},
		{
			desc:      "expect data sources that users has several actions for",
			sqlID:     "data_source.id",
			prefix:    "datasources",
			attribute: accesscontrol.ScopeAttributeID,
			actions:   []string{"datasources:read", "datasources:write"},
			permissions: map[string][]string{
				"datasources:read":  {"datasources:id:3", "datasources:id:7", "datasources:id:8"},
				"datasources:write": {"datasources:id:3", "datasources:id:8"},
			},
			expectedDataSources: []string{"ds:3", "ds:8"},
			expectErr:           false,
		},
		{
			desc:      "expect data sources that users has several actions for",
			sqlID:     "data_source.id",
			prefix:    "datasources",
			attribute: accesscontrol.ScopeAttributeID,
			actions:   []string{"datasources:read", "datasources:write"},
			permissions: map[string][]string{
				"datasources:read":  {"datasources:id:3", "datasources:id:7", "datasources:id:8"},
				"datasources:write": {"datasources:*", "datasources:id:8"},
			},
			expectedDataSources: []string{"ds:3", "ds:7", "ds:8"},
			expectErr:           false,
		},
		{
			desc:      "expect no data sources when scopes does not match",
			sqlID:     "data_source.id",
			prefix:    "datasources",
			attribute: accesscontrol.ScopeAttributeID,
			actions:   []string{"datasources:read", "datasources:write"},
			permissions: map[string][]string{
				"datasources:read":  {"datasources:id:3", "datasources:id:7", "datasources:id:8"},
				"datasources:write": {"datasources:id:10"},
			},
			expectedDataSources: []string{},
			expectErr:           false,
		},
		{
			desc:      "expect to not crash if duplicates in the scope",
			sqlID:     "data_source.id",
			prefix:    "datasources",
			attribute: accesscontrol.ScopeAttributeID,
			actions:   []string{"datasources:read", "datasources:write"},
			permissions: map[string][]string{
				"datasources:read":  {"datasources:id:3", "datasources:id:7", "datasources:id:8", "datasources:id:3", "datasources:id:8"},
				"datasources:write": {"datasources:id:3", "datasources:id:7"},
			},
			expectedDataSources: []string{"ds:3", "ds:7"},
			expectErr:           false,
		},
	}

	// set sqlIDAcceptList before running tests
	restore := accesscontrol.SetAcceptListForTest(map[string]struct{}{
		"data_source.id": {},
	})
	defer restore()

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			store := sqlstore.InitTestDB(t)

			sess := store.NewSession(context.Background())
			defer sess.Close()

			// seed 10 data sources
			for i := 1; i <= 10; i++ {
				err := store.AddDataSource(context.Background(), &models.AddDataSourceCommand{Name: fmt.Sprintf("ds:%d", i)})
				require.NoError(t, err)
			}

			baseSql := `SELECT data_source.* FROM data_source WHERE`
			acFilter, err := accesscontrol.Filter(
				&models.SignedInUser{
					OrgId:       1,
					Permissions: map[int64]map[string][]string{1: tt.permissions},
				},
				tt.sqlID,
				tt.prefix,
				"id",
				tt.actions...,
			)

			if !tt.expectErr {
				require.NoError(t, err)
				var datasources []models.DataSource
				err = sess.SQL(baseSql+acFilter.Where, acFilter.Args...).Find(&datasources)
				require.NoError(t, err)

				assert.Len(t, datasources, len(tt.expectedDataSources))
				for i, ds := range datasources {
					assert.Equal(t, tt.expectedDataSources[i], ds.Name)
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}
