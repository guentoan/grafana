package publicdashboard

import (
	"context"
	"fmt"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/accesscontrol"
	publicdashboardStore "github.com/grafana/grafana/pkg/services/publicdashboards/database"
	publicdashboardModels "github.com/grafana/grafana/pkg/services/publicdashboards/models"
	"github.com/grafana/grafana/pkg/services/user"
)

var _ Watcher = (*watcher)(nil)

type watcher struct {
	log                  log.Logger
	webhooks             *WebhooksAPI
	publicDashboardStore *publicdashboardStore.PublicDashboardStoreImpl
	userService          user.Service
	accessControlService accesscontrol.Service
}

func ProvideWatcher(userService user.Service, webhooks *WebhooksAPI, publicDashboardStore *publicdashboardStore.PublicDashboardStoreImpl, accessControlService accesscontrol.Service) *watcher {
	return &watcher{
		log:                  log.New("k8s.publicdashboard.service-watcher"),
		webhooks:             webhooks,
		publicDashboardStore: publicDashboardStore,
		userService:          userService,
		accessControlService: accessControlService,
	}
}

func (w *watcher) Add(ctx context.Context, obj *PublicDashboard) error {
	//convert to dto
	pdModel, err := k8sObjectToModel(obj)
	if err != nil {
		return err
	}

	fmt.Printf("%#v", pdModel)

	// convert to cmd
	cmd := publicdashboardModels.SavePublicDashboardCommand{
		PublicDashboard: *pdModel,
	}

	// call service
	_, err = w.publicDashboardStore.Create(ctx, cmd)
	if err != nil {
		return err
	}

	return nil
}

func (w *watcher) Update(ctx context.Context, oldObj, newObj *PublicDashboard) error {
	// TODO
	return nil
}

func (w *watcher) Delete(ctx context.Context, obj *PublicDashboard) error {
	// TODO
	return nil
}
