package model

import (
	"context"
	"errors"
	"fmt"
	"github.com/torresjeff/go-feature-lab/featurelab"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const FeatureLabDB = "featurelabdb"
const FeatureLabCollection = "featurelab"
const DefaultQueryTimeout = 3 * time.Second

var ErrDuplicateEntry = errors.New("entry already exists")
var ErrNoEntry = errors.New("no entry exists")

type FeatureAllocationEntity struct {
	Treatment string `bson:"treatment"`
	Weight    uint32 `bson:"weight"`
}

type FeatureEntity struct {
	Id          string                    `bson:"_id"`
	CreatedTime string                    `bson:"createdTime"`
	UpdatedTime string                    `bson:"updatedTime"`
	App         string                    `bson:"app"`
	Feature     string                    `bson:"feature"`
	Allocations []FeatureAllocationEntity `bson:"allocations"`
}
type FeatureLabDAO struct {
	featureLabCollection *mongo.Collection
	queryTimeout         time.Duration
}

func getKey(app, feature string) string {
	return fmt.Sprintf("%s:%s", app, feature)
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

func (dao *FeatureLabDAO) CreateFeature(feature featurelab.Feature) (FeatureEntity, error) {
	_, found, _ := dao.FetchFeature(feature.App, feature.Name)
	if found {
		return FeatureEntity{}, ErrDuplicateEntry
	}

	featureEntity := ToFeatureEntity(feature)
	featureEntity.Id = getKey(featureEntity.App, featureEntity.Feature)
	now := time.Now().Format(time.RFC3339)
	featureEntity.CreatedTime = now
	featureEntity.UpdatedTime = now

	ctx, cancel := context.WithTimeout(context.Background(), dao.queryTimeout)
	defer cancel()

	_, err := dao.featureLabCollection.InsertOne(ctx, featureEntity)
	if err != nil {
		// TODO: retries?
		return FeatureEntity{}, err
	}

	return featureEntity, nil
}

func (dao *FeatureLabDAO) UpdateFeature(feature featurelab.Feature) (FeatureEntity, error) {
	entityToUpdate, found, _ := dao.FetchFeature(feature.App, feature.Name)
	if !found {
		return FeatureEntity{}, ErrNoEntry
	}

	entityToUpdate.UpdatedTime = time.Now().Format(time.RFC3339)
	// The only thing we can really update is the allocations, app and feature name can't change as they make up the ID
	entityToUpdate.Allocations = ToFeatureAllocationEntities(feature.Allocations)

	ctx, cancel := context.WithTimeout(context.Background(), dao.queryTimeout)
	defer cancel()

	result, err := dao.featureLabCollection.ReplaceOne(ctx, bson.D{{"_id", entityToUpdate.Id}}, entityToUpdate)
	if err != nil {
		return FeatureEntity{}, err
	}
	if result.MatchedCount == 0 {
		return FeatureEntity{}, ErrNoEntry
	}

	return entityToUpdate, nil

}

// FetchFeature looks up a feature using app and feature name. It returns a FeatureEntity, a boolean that specifies whether the record
// was found, and an error in case any happened.
func (dao *FeatureLabDAO) FetchFeature(app, featureName string) (FeatureEntity, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dao.queryTimeout)
	defer cancel()

	var result FeatureEntity
	err := dao.featureLabCollection.
		FindOne(ctx, bson.D{{"_id", getKey(app, featureName)}}).
		Decode(&result)

	if err == mongo.ErrNoDocuments {
		return FeatureEntity{}, false, nil
	} else if err != nil {
		// TODO: error handling, retries
		return FeatureEntity{}, false, err
	}

	return result, true, nil
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
