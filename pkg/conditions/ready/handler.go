package ready

import (
	"context"

	"github.com/giantswarm/conditions/pkg/conditions"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/conditions-handler/pkg/internal"
	"github.com/giantswarm/conditions-handler/pkg/key"
)

type HandlerConfig struct {
	CtrlClient ctrl.Client
	Logger     micrologger.Logger

	UpdateStatusOnConditionChange bool
	ConditionsToSummarize         []capi.ConditionType
	IgnoreOptions                 []conditions.CheckOption
}

type Handler struct {
	ctrlClient            ctrl.Client
	internalHandler       *internal.Handler
	logger                micrologger.Logger
	conditionsToSummarize []capi.ConditionType
	ignoreOptions         []conditions.CheckOption
}

func NewHandler(config HandlerConfig) (*Handler, error) {
	h := &Handler{
		ctrlClient:            config.CtrlClient,
		logger:                config.Logger,
		conditionsToSummarize: config.ConditionsToSummarize,
		ignoreOptions:         config.IgnoreOptions,
	}

	internalHandlerConfig := internal.HandlerConfig{
		CtrlClient:                    config.CtrlClient,
		Logger:                        config.Logger,
		UpdateStatusOnConditionChange: config.UpdateStatusOnConditionChange,
		ConditionType:                 capi.ReadyCondition,
		EnsureCreatedFunc:             h.ensureCreated,
	}

	internalHandler, err := internal.NewHandler(internalHandlerConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	h.internalHandler = internalHandler

	return h, nil
}

func (h *Handler) EnsureCreated(ctx context.Context, object interface{}) error {
	obj, err := key.ToObjectWithConditions(object)
	if err != nil {
		return microerror.Mask(err)
	}

	return h.internalHandler.EnsureCreated(ctx, obj)
}

func (h *Handler) EnsureDeleted(_ context.Context, _ interface{}) error {
	return nil
}

func (h *Handler) ensureCreated(_ context.Context, object conditions.Object) error {
	update(object, h.conditionsToSummarize, h.ignoreOptions...)
	return nil
}
