package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/torresjeff/go-feature-lab-server/handler"
	"github.com/torresjeff/go-feature-lab-server/model"
	"github.com/torresjeff/go-feature-lab/featurelab"
	"log"
	"net/http"
	"strings"
	"time"
)

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

	errNotFound := "{\"error\": \"app or feature not found\"}"
	errBadRequest := "{\"error\": \"invalid request, please check your request and try again\"}"
	errGeneric := "{\"error\": \"%v\"}"

	featureHandler := handler.NewFeatureHandler(featureLabDAO, featurelab.NewTreatmentAssigner())

	app := fiber.New()

	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Fetches all features for an app
	v1.Get("/app/:app/features", func(c *fiber.Ctx) error {
		app := strings.TrimSpace(c.Params("app"))
		if app == "" {
			return c.Status(http.StatusBadRequest).Send([]byte(errBadRequest))
		}

		features, err := featureHandler.FetchFeatures(app)
		if err != nil {
			return c.Status(http.StatusInternalServerError).Send([]byte(fmt.Sprintf(errGeneric, err)))
		}

		return c.Status(http.StatusOK).JSON(features)
	})

	// Fetches a specific feature for an app
	v1.Get("/app/:app/features/:feature", func(c *fiber.Ctx) error {
		app := strings.TrimSpace(c.Params("app"))
		feature := strings.TrimSpace(c.Params("feature"))
		if app == "" || feature == "" {
			return c.Status(http.StatusBadRequest).Send([]byte(errBadRequest))
		}

		featureEntity, err := featureLabDAO.FetchFeature(app, feature)
		if err == model.ErrNotFound {
			return c.Status(http.StatusNotFound).Send([]byte(err.Error()))
		} else if err != nil {
			return c.Status(http.StatusInternalServerError).Send([]byte(fmt.Sprintf(errGeneric, err)))
		}

		return c.Status(http.StatusOK).JSON(model.ToFeature(featureEntity))
	})

	// Calculates the treatment for a particular feature, given the criteria (as query param). Criteria should be URL encoded.
	v1.Get("/app/:app/features/:feature/treatment/:criteria", func(c *fiber.Ctx) error {
		app := strings.TrimSpace(c.Params("app"))
		feature := strings.TrimSpace(c.Params("feature"))
		criteria := strings.TrimSpace(c.Params("criteria"))
		if app == "" || feature == "" || criteria == "" {
			return c.Status(http.StatusBadRequest).Send([]byte(errBadRequest))
		}

		treatment, err := featureHandler.GetTreatment(app, feature, criteria)
		if err == model.ErrNotFound {
			return c.Status(http.StatusNotFound).Send([]byte(errNotFound))
		} else if err != nil {
			return c.Status(http.StatusInternalServerError).Send([]byte(fmt.Sprintf(errGeneric, err)))
		}

		return c.Status(http.StatusOK).JSON(treatment)
	})

	log.Fatal(app.Listen(":3000"))

}
