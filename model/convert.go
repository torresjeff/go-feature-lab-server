package model

import "github.com/torresjeff/go-feature-lab/featurelab"

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
