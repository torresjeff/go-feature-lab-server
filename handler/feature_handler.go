package handler

import (
	"github.com/torresjeff/go-feature-lab-server/model"
	"github.com/torresjeff/go-feature-lab/featurelab"
)

type FeatureHandler struct {
	dao               *model.FeatureLabDAO
	treatmentAssigner featurelab.TreatmentAssigner
}

func NewFeatureHandler(dao *model.FeatureLabDAO, ta featurelab.TreatmentAssigner) *FeatureHandler {
	return &FeatureHandler{
		dao:               dao,
		treatmentAssigner: ta,
	}
}

func (handler *FeatureHandler) FetchFeatures(app string) ([]featurelab.Feature, error) {
	featureEntities, err := handler.dao.FetchFeatures(app)
	if err != nil {
		return nil, err
	}

	return model.ToFeatures(featureEntities), nil
}

func (handler *FeatureHandler) FetchFeature(app, feature string) (featurelab.Feature, error) {
	featureEntity, err := handler.dao.FetchFeature(app, feature)
	if err != nil {
		return featurelab.Feature{}, err
	}

	return model.ToFeature(featureEntity), nil
}

func (handler *FeatureHandler) GetTreatment(app, feature, criteria string) (featurelab.TreatmentAssignment, error) {
	featureObj, err := handler.FetchFeature(app, feature)
	if err != nil {
		return featurelab.TreatmentAssignment{}, err
	}

	return handler.treatmentAssigner.GetTreatmentAssignment(featureObj, criteria)
}
