package model

import "github.com/torresjeff/go-feature-lab/featurelab"

func ToFeatureEntity(feature featurelab.Feature) FeatureEntity {
	return FeatureEntity{
		App:         feature.App,
		Feature:     feature.Name,
		Allocations: ToFeatureAllocationEntities(feature.Allocations),
	}
}

func ToFeatureAllocationEntities(allocations []featurelab.FeatureAllocation) []FeatureAllocationEntity {
	featureAllocationEntities := make([]FeatureAllocationEntity, len(allocations))
	for i, allocation := range allocations {
		featureAllocationEntities[i] = FeatureAllocationEntity{
			Treatment: allocation.Treatment,
			Weight:    allocation.Weight,
		}
	}

	return featureAllocationEntities
}

func ToFeature(featureEntity FeatureEntity) featurelab.Feature {
	return featurelab.NewFeature(featureEntity.App, featureEntity.Feature, ToFeatureAllocations(featureEntity.Allocations))
}

func ToFeatures(featureEntities []FeatureEntity) []featurelab.Feature {
	features := make([]featurelab.Feature, len(featureEntities))
	for i, feature := range featureEntities {
		features[i] = featurelab.NewFeature(feature.App, feature.Feature, ToFeatureAllocations(feature.Allocations))
	}

	return features
}

func ToFeatureAllocations(allocationEntities []FeatureAllocationEntity) []featurelab.FeatureAllocation {
	allocations := make([]featurelab.FeatureAllocation, len(allocationEntities))
	for i, allocation := range allocationEntities {
		allocations[i] = featurelab.NewFeatureAllocation(allocation.Treatment, allocation.Weight)
	}

	return allocations
}
