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

const App = "app"
const Feature = "feature"

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

	// Create feature
	v1.Post("/app/:app/features", func(c *fiber.Ctx) error {
		var feature featurelab.Feature
		err := parseBody(c, &feature)
		if err != nil {
			return err
		}

		feature.App = strings.TrimSpace(c.Params(App))
		feature.Name = strings.TrimSpace(feature.Name)

		log.Printf("Got request to create feature: %s:%s\n", feature.App, feature.Name)

		if feature.App == "" || feature.Name == "" {
			return sendError(c, featurelab.NewError(featurelab.ErrBadRequest, "invalid app and/or feature"))
		}

		result, err := featureHandler.CreateFeature(feature)
		if err == model.ErrDuplicateEntry {
			return sendError(c, featurelab.NewError(featurelab.ErrBadRequest, "feature already exists"))
		} else if err != nil {
			log.Printf("Error creating feature: %s\n", err)
			return sendError(c, ErrInternalServer)
		}

		c.Set("Location", fmt.Sprintf("/app/%s/features/%s", result.App, result.Name))
		return c.Status(http.StatusCreated).JSON(result)
	})

	// Update feature
	v1.Put("/app/:app/features/:feature", func(c *fiber.Ctx) error {

		app := strings.TrimSpace(c.Params(App))
		feature := strings.TrimSpace(c.Params(Feature))

		log.Printf("Got request to update feature: %s:%s\n", app, feature)

		if app == "" || feature == "" {
			return sendError(c, featurelab.NewError(featurelab.ErrBadRequest, "invalid app and/or feature"))
		}

		var f featurelab.Feature
		err := parseBody(c, &f)
		f.App = app
		f.Name = feature
		if err != nil {
			return err
		} else if f.Allocations == nil {
			return sendError(c, featurelab.NewError(featurelab.ErrBadRequest, "invalid allocations, allocations must not be null"))
		}

		_, err = featureHandler.UpdateFeature(f)
		if err == model.ErrNoEntry {
			return sendError(c, featurelab.NewError(featurelab.ErrNotFound, "feature not found"))
		} else if err != nil {
			log.Printf("Error updating feature: %s\n", err)
			return sendError(c, ErrInternalServer)
		}

		return c.Status(http.StatusNoContent).Send(nil)
	})

	// Fetches all features for an app
	v1.Get("/app/:app/features", func(c *fiber.Ctx) error {
		app := strings.TrimSpace(c.Params(App))

		log.Printf("Got request to fetch features for app: %s\n", app)

		if app == "" {
			return sendError(c, featurelab.NewError(featurelab.ErrBadRequest, "invalid app name"))
		}

		features, err := featureHandler.FetchFeatures(app)
		if err != nil {
			log.Printf("error: %s\n", err)
			return sendError(c, ErrInternalServer)
		}

		return c.Status(http.StatusOK).JSON(features)
	})

	// Fetches a specific feature for an app
	v1.Get("/app/:app/features/:feature", func(c *fiber.Ctx) error {
		app := strings.TrimSpace(c.Params(App))
		feature := strings.TrimSpace(c.Params(Feature))

		log.Printf("Got request to fetch feature: %s:%s\n", app, feature)

		if app == "" || feature == "" {
			return sendError(c, featurelab.NewError(featurelab.ErrBadRequest, "invalid app and/or feature"))
		}

		result, found, err := featureHandler.FetchFeature(app, feature)
		if !found {
			return sendError(c, featurelab.NewError(featurelab.ErrNotFound, "app and/or feature doesn't exist"))
		} else if err != nil {
			return sendError(c, ErrInternalServer)
		}

		return c.Status(http.StatusOK).JSON(result)
	})

	// Calculates the treatment for a particular feature, given the criteria
	v1.Get("/app/:app/features/:feature/treatment/:criteria", func(c *fiber.Ctx) error {
		app := strings.TrimSpace(c.Params(App))
		feature := strings.TrimSpace(c.Params(Feature))
		criteria := strings.TrimSpace(c.Params("criteria"))

		log.Printf("Got request to calculate treatment for feature: %s:%s, criteria: %s\n", app, feature, criteria)

		if app == "" || feature == "" || criteria == "" {
			return sendError(c, featurelab.NewError(featurelab.ErrBadRequest, "invalid app, feature and/or criteria"))
		}

		treatment, err := featureHandler.GetTreatment(app, feature, criteria)
		if err == handler.ErrNotFound {
			return sendError(c, featurelab.NewError(featurelab.ErrNotFound, "app and/or feature doesn't exist"))
		} else if err != nil {
			log.Printf("error: %s\n", err)
			return sendError(c, ErrInternalServer)
		}

		return c.Status(http.StatusOK).JSON(treatment)
	})

	log.Fatal(app.Listen(":3000"))

}

func parseBody(c *fiber.Ctx, out interface{}) error {
	err := c.BodyParser(out)
	if err == fiber.ErrUnprocessableEntity {
		return sendError(c, featurelab.NewError(featurelab.ErrBadRequest, "invalid request (did you forget to set Content-Type header?)"))
	} else if err != nil {
		log.Printf("error parsing request body: %s", err)
		return sendError(c, featurelab.NewError(featurelab.ErrBadRequest, "invalid feature JSON"))
	}

	return nil
}

func sendError(c *fiber.Ctx, err featurelab.Error) error {
	return c.Status(int(err.Code)).JSON(err)
}
