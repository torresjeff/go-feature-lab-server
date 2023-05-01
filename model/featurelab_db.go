package model

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func NewFeatureLabDAO(ctx context.Context, mongoURI string, queryTimeout time.Duration) (dao *FeatureLabDAO, disconnect func()) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))

	disconnect = func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}

	dao = &FeatureLabDAO{
		featureLabCollection: client.Database(FeatureLabDB).Collection(FeatureLabCollection),
		queryTimeout:         queryTimeout,
	}

	return dao, disconnect
}

func (dao *FeatureLabDAO) FetchFeature(app, featureName string) (FeatureEntity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dao.queryTimeout)
	defer cancel()

	var result FeatureEntity
	err := dao.featureLabCollection.
		FindOne(ctx, bson.D{{"app", app}, {"feature", featureName}}).
		Decode(&result)

	if err == mongo.ErrNoDocuments {
		return FeatureEntity{}, ErrNotFound
	} else if err != nil {
		// TODO: error handling, retries
		return FeatureEntity{}, err
	}

	return result, nil
}

func (dao *FeatureLabDAO) FetchFeatures(app string) ([]FeatureEntity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dao.queryTimeout)
	defer cancel()

	cursor, err := dao.featureLabCollection.Find(ctx, bson.D{{"app", app}})
	if err != nil {
		return nil, err
	}

	var results []FeatureEntity
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
