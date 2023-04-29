package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/torresjeff/go-feature-lab/featurelab"
	"log"
	"net/http"
	"net/url"
)

func main() {
	features := []featurelab.Feature{
		featurelab.NewFeature("ShowRecommendations", []featurelab.FeatureAllocation{
			featurelab.NewFeatureAllocation("C", 10),
			featurelab.NewFeatureAllocation("T1", 10),
			featurelab.NewFeatureAllocation("T2", 10),
		}),
		featurelab.NewFeature("ChangeBuyButtonColor", []featurelab.FeatureAllocation{
			featurelab.NewFeatureAllocation("C", 32),
			featurelab.NewFeatureAllocation("T1", 68),
		}),
	}

	errAppNotFound := "{\"error\": \"app not found or invalid\"}"
	errFeatureNotFound := "{\"error\": \"feature not found or invalid\"}"
	errInvalidCriteria := "{\"error\": \"criteria is empty or not valid\"}"
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
		if c.Params("app") == "FeatureLab" {
			if c.Params("feature") == "ShowRecommendations" {
				return c.Status(http.StatusOK).JSON(features[0])
			} else if c.Params("feature") == "ChangeBuyButtonColor" {
				return c.Status(http.StatusOK).JSON(features[1])
			} else {
				return c.Status(http.StatusNotFound).Send([]byte(errFeatureNotFound))
			}
		}

		return c.Status(http.StatusNotFound).Send([]byte(errAppNotFound))
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
