package model

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

const FeatureLabDB = "featurelabdb"
const FeatureLabCollection = "featurelab"
const DefaultQueryTimeout = 3 * time.Second

var ErrNotFound = errors.New("application or feature not found")

type FeatureAllocationEntity struct {
	Treatment string `json:"treatment"`
	Weight    uint32 `json:"weight"`
}

type FeatureEntity struct {
	App         string                    `json:"app"`
	Feature     string                    `json:"feature"`
	Allocations []FeatureAllocationEntity `json:"allocations"`
}
type FeatureLabDAO struct {
	featureLabCollection *mongo.Collection
	queryTimeout         time.Duration
}

func NewFeatureLabDAO(mongo *mongo.Client, queryTimeout time.Duration) *FeatureLabDAO {
	return &FeatureLabDAO{
		featureLabCollection: mongo.Database(FeatureLabDB).Collection(FeatureLabCollection),
		queryTimeout:         queryTimeout,
	}
}

func (dao *FeatureLabDAO) FetchFeature(app, featureName string) (*FeatureEntity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dao.queryTimeout)
	defer cancel()
	var result FeatureEntity
	err := dao.featureLabCollection.
		FindOne(ctx, bson.D{{"app", app}, {"feature", featureName}}).
		Decode(&result)

	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	} else if err != nil {
		// TODO: error handling, retries
		return nil, err
	}

	return &result, nil
}
