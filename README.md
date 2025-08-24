# TFL Tasks

* Supports a Siri shortcut to find the nearest Santander docks & bikes, and reads the results out.

## About

This app uses the TFL API for Santander Cycles to find the nearest docks and bikes, and presents them for Siri to read.
Basically, automate my morning commute.

## Get closest dock to a location
e.g. `GET /cycles/nearby/siri?coords=51.51381379567755,-0.08370914455887864`

Useful for finding the nearest dock to your current location. See the iOS shortcut I use below to inject my live location.

## Query specific docks
e.g. `GET /cycles/stations/siri?ids=BikePoints_26,BikePoints_79,BikePoints_102,BikePoints_14`

This is useful if you have a set of preferred docks you want to check before you leave, for example the nearest docks to your home and work. Get the IDs from the TFL API and add them in a shortcut or link.

You will receive the results in the order you provided the IDs, so you should think of the order as your preference.

## Requirements
* To deploy on AWS, you'll need the aws cli installed and configured to a role with Lambda deployment & IAM role managemenet permissions.
* TFL_APP_KEY - you can get one free from https://api-portal.tfl.gov.uk/signup

## First time deployment
Make sure you have the aws cli installed and configured as described above.

* `make lambda/role/create` -- create the IAM role for the Lambda function
* `make lambda/build` -- build the zip artefact
* `make lambda/create` -- create the Lambda function in your AWS account
This will succeed but the function won't work yet, as we need to configure env vars next.

* `TFL_APP_KEY=${YOUR_KEY_HERE} make lambda/config` -- set the TFL API key as an env var on the Lambda function
* `make lambda/public-url` -- this will give you a publicly accessible URL for the function

## Subsequent deployments
* `make lambda/build` -- make sure you rebuild
* `make lambda/update` -- redeploys the function

## iOS Shortcuts
All of these endpoints work well as bookmarks in the browser.

However I quite like to trigger these with Siri, and have the results read aloud to me as I'm biking.

The below assumes you have a basic understanding of the iOS Shortcuts App.

### Example: Nearest Available Dock & Bike
Step 1) Get current location
Step 2) Get Contents of URL
	* The format of the URL will be "${YOUR_LAMBDA_URL}/cycles/nearby/siri?coords=LATITUDE,LONGITUDE"
	* LATITUDE & LONGITUDE are the variables Shortcuts provides from step 1
Step 3) Speak Contents of URL

### Example: Query specific docks
* Find the specific docks you need here: https://api.tfl.gov.uk/BikePoint
* Make a note of their `id` field

Step 1) Get Contents of URL
	* The format of the URL will be "${YOUR_LAMBDA_URL}/cycles/stations/siri?ids=BikePoints_89,BikePoints_26
	* You will receive the results in the order you provided them
Step 2) Speak Contents of URL

