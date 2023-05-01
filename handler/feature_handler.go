package handler

import (
	"errors"
	"github.com/torresjeff/go-feature-lab-server/model"
	"github.com/torresjeff/go-feature-lab/featurelab"
)

var ErrNotFound = errors.New("application and/or feature not found")

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

func (handler *FeatureHandler) CreateFeature(feature featurelab.Feature) (featurelab.Feature, error) {
	featureEntity, err := handler.dao.CreateFeature(feature)
	if err != nil {
		return featurelab.Feature{}, err
	}

	return model.ToFeature(featureEntity), nil
}

func (handler *FeatureHandler) UpdateFeature(feature featurelab.Feature) (featurelab.Feature, error) {
	featureEntity, err := handler.dao.UpdateFeature(feature)
	if err != nil {
		return featurelab.Feature{}, err
	}

	return model.ToFeature(featureEntity), nil
}

func (handler *FeatureHandler) FetchFeatures(app string) ([]featurelab.Feature, error) {
	featureEntities, err := handler.dao.FetchFeatures(app)
	if err != nil {
		return nil, err
	}

	return model.ToFeatures(featureEntities), nil
}

func (handler *FeatureHandler) FetchFeature(app, feature string) (featurelab.Feature, bool, error) {
	featureEntity, found, err := handler.dao.FetchFeature(app, feature)

	if err != nil {
		return featurelab.Feature{}, false, err
	} else if !found {
		return featurelab.Feature{}, false, nil
	}

	result := model.ToFeature(featureEntity)
	return result, true, nil
}

func (handler *FeatureHandler) GetTreatment(app, feature, criteria string) (featurelab.TreatmentAssignment, error) {
	featureObj, found, err := handler.FetchFeature(app, feature)
	if err != nil {
		return featurelab.TreatmentAssignment{}, err
	} else if !found {
		return featurelab.TreatmentAssignment{}, ErrNotFound
	}

	return handler.treatmentAssigner.GetTreatmentAssignment(featureObj, criteria)
}
