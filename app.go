package main

import (
	"context"
	"flag"
	"github.com/gofiber/fiber/v2"
	"github.com/torresjeff/go-feature-lab-server/handler"
	"github.com/torresjeff/go-feature-lab-server/model"
	"github.com/torresjeff/go-feature-lab/featurelab"
	"log"
	"net/http"
	"strings"
	"time"
)

var ErrInternalServer = featurelab.NewError(featurelab.ErrInternalServerError, "an unexpected error occurred, please try again later")

func main() {
	var mongoURI string
	flag.StringVar(&mongoURI,
		"mongo",
		"mongodb://mongodb:27017", // defined by docker-compose.yml
		"URI where mongo istance is located. Eg: mongodb://hostname:27017")
	flag.Parse()

	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	featureLabDAO, disconnect := model.NewFeatureLabDAO(dbCtx, mongoURI, model.DefaultQueryTimeout)
	defer disconnect()

	featureHandler := handler.NewFeatureHandler(featureLabDAO, featurelab.NewTreatmentAssigner())

	app := fiber.New()

	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Fetches all features for an app
	v1.Get("/app/:app/features", func(c *fiber.Ctx) error {
		app := strings.TrimSpace(c.Params("app"))
		if app == "" {
			return c.Status(http.StatusBadRequest).JSON(
				featurelab.NewError(featurelab.ErrBadRequest, "invalid app name"))
		}

		features, err := featureHandler.FetchFeatures(app)
		if err != nil {
			log.Printf("error: %s\n", err)
			return c.Status(http.StatusInternalServerError).JSON(ErrInternalServer)
		}

		return c.Status(http.StatusOK).JSON(features)
	})

	// Fetches a specific feature for an app
	v1.Get("/app/:app/features/:feature", func(c *fiber.Ctx) error {
		app := strings.TrimSpace(c.Params("app"))
		feature := strings.TrimSpace(c.Params("feature"))
		if app == "" || feature == "" {
			return c.Status(http.StatusBadRequest).JSON(
				featurelab.NewError(featurelab.ErrBadRequest, "invalid app and/or feature"))
		}

		featureEntity, err := featureLabDAO.FetchFeature(app, feature)
		if err == model.ErrNotFound {
			return c.Status(http.StatusNotFound).JSON(featurelab.NewError(featurelab.ErrNotFound, "app and/or feature doesn't exist"))
		} else if err != nil {
			log.Printf("error: %s\n", err)
			return c.Status(http.StatusInternalServerError).JSON(ErrInternalServer)
		}

		return c.Status(http.StatusOK).JSON(model.ToFeature(featureEntity))
	})

	// Calculates the treatment for a particular feature, given the criteria
	v1.Get("/app/:app/features/:feature/treatment/:criteria", func(c *fiber.Ctx) error {
		app := strings.TrimSpace(c.Params("app"))
		feature := strings.TrimSpace(c.Params("feature"))
		criteria := strings.TrimSpace(c.Params("criteria"))
		if app == "" || feature == "" || criteria == "" {
			return c.Status(http.StatusBadRequest).JSON(
				featurelab.NewError(featurelab.ErrBadRequest, "invalid app, feature and/or criteria"))
		}

		treatment, err := featureHandler.GetTreatment(app, feature, criteria)
		if err == model.ErrNotFound {
			return c.Status(http.StatusNotFound).JSON(featurelab.NewError(featurelab.ErrNotFound, "app and/or feature doesn't exist"))
		} else if err != nil {
			log.Printf("error: %s\n", err)
			return c.Status(http.StatusInternalServerError).JSON(ErrInternalServer)
		}

		return c.Status(http.StatusOK).JSON(treatment)
	})

	log.Fatal(app.Listen(":3000"))

}
