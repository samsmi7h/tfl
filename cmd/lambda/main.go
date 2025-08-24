package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"tfl"
	"tfl/pkg"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(cfg tfl.Config) func(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	return func(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
		switch req.RawPath {
		case "/cycles/nearby/siri":
			coords := req.QueryStringParameters["coords"]
			message, err := handlerCyclesNearBySiri(cfg, coords)
			if err != nil {
				return events.LambdaFunctionURLResponse{}, err
			}

			return events.LambdaFunctionURLResponse{
				StatusCode: 200,
				Body:       message,
			}, nil
		case "/cycles/stations/siri":
			ids := req.QueryStringParameters["ids"]
			message, err := handlerCycleStationsQueryByIDs(cfg, ids)
			if err != nil {
				return events.LambdaFunctionURLResponse{}, err
			}

			return events.LambdaFunctionURLResponse{
				StatusCode: 200,
				Body:       message,
			}, nil
		}

		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusNotFound,
		}, nil
	}
}

func handlerCycleStationsQueryByIDs(cfg tfl.Config, idsStr string) (string, error) {
	ids := strings.Split(idsStr, ",")
	if len(ids) == 0 {
		return "", fmt.Errorf("invalid ids parameter")
	}

	fmt.Println("Request: /cycles/stations?ids=", idsStr)
	message, err := tfl.HandleQueryStationAvailability(cfg, ids)
	if err != nil {
		return "", fmt.Errorf("error handling query station availability: %w", err)
	}

	return message, nil
}

func handlerCyclesNearBySiri(cfg tfl.Config, coordStr string) (string, error) {
	parts := strings.Split(coordStr, ",")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid coords parameter")
	}

	fmt.Println("Request: /cycles/nearby?coords=", coordStr)

	lat, lng, err := pkg.ParseCoords(coordStr)
	if err != nil {
		return "", fmt.Errorf("invalid coords parameter: %w", err)
	}

	message, err := tfl.HandleGetNearbyCycleStations(cfg, lat, lng)
	if err != nil {
		return "", err
	}

	return message, nil
}

func main() {
	cfg, err := tfl.Init()
	if err != nil {
		fmt.Println("Failed to initialize config:", err)
		return
	}

	lambda.Start(handler(cfg))
}
