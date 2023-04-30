package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/torresjeff/go-feature-lab-server/model"
	"github.com/torresjeff/go-feature-lab/featurelab"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"net/url"
	"time"
)

func main() {
	var mongoURI string
	flag.StringVar(&mongoURI,
		"mongo",
		"mongodb://mongodb:27017", // defined by docker-compose.yml
		"URI where mongo istance is located. Eg: mongodb://hostname:27017")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	featureLabDAO := model.NewFeatureLabDAO(client, model.DefaultQueryTimeout)

	features := []featurelab.Feature{
		featurelab.NewFeature("FeatureLab", "ShowRecommendations", []featurelab.FeatureAllocation{
			featurelab.NewFeatureAllocation("C", 10),
			featurelab.NewFeatureAllocation("T1", 10),
			featurelab.NewFeatureAllocation("T2", 10),
		}),
		featurelab.NewFeature("FeatureLab", "ChangeBuyButtonColor", []featurelab.FeatureAllocation{
			featurelab.NewFeatureAllocation("C", 32),
			featurelab.NewFeatureAllocation("T1", 68),
		}),
	}

	errAppNotFound := "{\"error\": \"app or feature not found\"}"
	errFeatureNotFound := "{\"error\": \"feature not found or invalid\"}"
	errInvalidCriteria := "{\"error\": \"criteria is empty or not valid\"}"
	errBadRequest := "{\"error\": \"invalid request, please check your request and try again\"}"
	errGeneric := "{\"error\": \"%v\"}"

	app := fiber.New()

	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Fetches all features for an app
	v1.Get("/app/:app/features", func(c *fiber.Ctx) error {
		if c.Params("app") == "FeatureLab" {
			return c.Status(http.StatusOK).JSON(features)
		}

		return c.Status(http.StatusNotFound).Send([]byte(errAppNotFound))
	})

	// Fetches a specific feature for an app
	v1.Get("/app/:app/features/:feature", func(c *fiber.Ctx) error {
		app := c.Params("app")
		feature := c.Params("feature")
		if app == "" || feature == "" {
			return c.Status(http.StatusBadRequest).Send([]byte(errBadRequest))
		}

		featureEntity, err := featureLabDAO.FetchFeature(app, feature)
		if err == model.ErrNotFound {
			return c.Status(http.StatusNotFound).Send([]byte(err.Error()))
		} else if err != nil {
			return c.Status(http.StatusInternalServerError).Send([]byte(fmt.Sprintf(errGeneric, err)))
		}

		allocations := make([]featurelab.FeatureAllocation, len(featureEntity.Allocations))
		for i, allocation := range featureEntity.Allocations {
			allocations[i] = featurelab.NewFeatureAllocation(allocation.Treatment, allocation.Weight)
		}

		result := featurelab.NewFeature(featureEntity.App, featureEntity.Feature, allocations)
		log.Printf("converted entity %+v to %+v\n", featureEntity, result)

		return c.Status(http.StatusOK).JSON(featurelab.NewFeature(featureEntity.App, featureEntity.Feature, allocations))
	})

	treatmentAssigner := featurelab.NewTreatmentAssigner()
	// Calculates the treatment for a particular feature, given the criteria (as query param). Criteria should be URL encoded.
	v1.Get("/app/:app/features/:feature/treatment", func(c *fiber.Ctx) error {
		criteria := c.Query("criteria")
		criteria, err := url.QueryUnescape(criteria)
		if err != nil {
			return c.Status(http.StatusBadRequest).Send([]byte(fmt.Sprintf(errGeneric, err)))
		}
		if criteria == "" {
			return c.Status(http.StatusBadRequest).Send([]byte(errInvalidCriteria))
		}

		if c.Params("app") == "FeatureLab" {
			if c.Params("feature") == "ShowRecommendations" {
				treatment, err := treatmentAssigner.GetTreatmentAssignment(features[0], criteria)
				if err != nil {
					return c.Status(http.StatusInternalServerError).Send([]byte(fmt.Sprintf(errGeneric, err)))
				}
				return c.Status(http.StatusOK).JSON(treatment)
			} else if c.Params("feature") == "ChangeBuyButtonColor" {
				treatment, err := treatmentAssigner.GetTreatmentAssignment(features[1], criteria)
				if err != nil {
					return c.Status(http.StatusInternalServerError).Send([]byte(fmt.Sprintf(errGeneric, err)))
				}
				return c.Status(http.StatusOK).JSON(treatment)
			} else {
				return c.Status(http.StatusNotFound).Send([]byte(errFeatureNotFound))
			}
		}

		return c.Status(http.StatusNotFound).Send([]byte(errAppNotFound))
	})

	log.Fatal(app.Listen(":3000"))

}
