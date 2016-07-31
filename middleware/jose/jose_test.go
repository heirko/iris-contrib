package iris_jose

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/kataras/iris"
)

type Claim struct {
	Name string
}

type Response struct {
	Text string
}

func TestBasicJose(t *testing.T) {
	var (
		privateKey, _    = rsa.GenerateKey(rand.Reader, 2048)
		api              = iris.New()
		myJoseMiddleware = New(Config{
			KeysGetter: func() (interface{}, interface{}, error) {
				return privateKey, &privateKey.PublicKey, nil
			},
		})
	)

	securedPingHandler := func(ctx *iris.Context) {
		claimString := myJoseMiddleware.Get(ctx)

		response := Response{Text: "Iauthenticated " + string(claimString)}
		log.Printf("literal claim string : %s", claimString)
		claim := Claim{}
		json.Unmarshal(claimString, &claim)
		log.Printf("claim : %s", claim.Name)
		ctx.JSON(iris.StatusOK, response)
	}

	api.Get("/secured/ping", myJoseMiddleware.Serve, securedPingHandler)

	e := api.Tester(t)
	log.Printf("Test no 1 : should be unauthorized")
	e.GET("/secured/ping").Expect().Status(iris.StatusUnauthorized)

	log.Printf("Test no 2 : should be authorized")

	claim := Claim{Name: " with Jose"}
	token := myJoseMiddleware.NewTokenWithClaim(claim)
	e.GET("/secured/ping").WithHeader("Authorization", "Bearer "+token).
		Expect().Status(iris.StatusOK).Body().
		Contains("Iauthenticated").Contains("with Jose")
}

func ExampleUsingJose(t *testing.T) {
	var (
		privateKey, _    = rsa.GenerateKey(rand.Reader, 2048)
		api              = iris.New()
		myJoseMiddleware = New(Config{
			KeysGetter: func() (interface{}, interface{}, error) {
				return privateKey, &privateKey.PublicKey, nil
			},
		})
	)

	securedPingHandler := func(ctx *iris.Context) {
		claimString := myJoseMiddleware.Get(ctx)

		response := Response{Text: "Iauthenticated " + string(claimString)}
		log.Printf("claim string : %s", claimString)
		claim := Claim{}
		json.Unmarshal(claimString, &claim)
		fmt.Println("claim : %s", claim.Name)
		// Output:
		// claim :  with Jose
		ctx.JSON(iris.StatusOK, response)
	}

	// assign middleware and callback
	api.Get("/secured/ping", myJoseMiddleware.Serve, securedPingHandler)

	e := api.Tester(t)

	log.Printf("Test no 2 : should be authorized")

	// Create a new token
	claim := Claim{Name: " with Jose"}
	token := myJoseMiddleware.NewTokenWithClaim(claim)
	e.GET("/secured/ping").WithHeader("Authorization", "Bearer "+token).
		Expect().Status(iris.StatusOK).Body().
		Contains("Iauthenticated").Contains("with Jose")
}
