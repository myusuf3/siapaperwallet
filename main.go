package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/NebulousLabs/Sia/crypto"
	"github.com/NebulousLabs/Sia/modules"
	"github.com/NebulousLabs/Sia/types"
	"github.com/NebulousLabs/fastrand"
	"github.com/gorilla/mux"
	"github.com/skip2/go-qrcode"
)

const addressCount = 20

var homeTemplate = template.Must(template.ParseFiles("templates/layout.html", "templates/home.html"))
var secretTemplate = template.Must(template.ParseFiles("templates/layout.html", "templates/secret.html"))

type addressPair struct {
	Address      types.UnlockHash
	AddressImage string
}

type secret struct {
	Seed         string
	SeedImage    string
	AddressPairs []addressPair
}

func HandleWalletGenerator(w http.ResponseWriter, r *http.Request) {
	templateData := generateNewSeed()
	secretTemplate.Execute(w, templateData)
}

func HandleHome(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, nil)
}

func RedirectToHTTPSRouter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		proto := req.Header.Get("x-forwarded-proto")
		if proto == "http" || proto == "HTTP" {
			http.Redirect(res, req, fmt.Sprintf("https://%s%s", req.Host, req.URL), http.StatusPermanentRedirect)
			return
		}

		next.ServeHTTP(res, req)

	})
}

func generateAddress(seed modules.Seed, index uint64) types.UnlockHash {
	_, pk := crypto.GenerateKeyPairDeterministic(crypto.HashAll(seed, index))
	return types.UnlockConditions{
		PublicKeys:         []types.SiaPublicKey{types.Ed25519PublicKey(pk)},
		SignaturesRequired: 1,
	}.UnlockHash()
}

func generateNewSeed() *secret {
	var seed modules.Seed
	fastrand.Read(seed[:])
	var addressesPairs []addressPair
	var png []byte
	seedStr, err := modules.SeedToString(seed, "english")
	if err != nil {
		log.Fatal(err)
	}
	png, err = qrcode.Encode(seedStr, qrcode.Low, 256)
	if err != nil {
		log.Fatal(err)
	}
	seedImage := base64.StdEncoding.EncodeToString(png)

	for i := uint64(0); i < addressCount; i++ {
		address := generateAddress(seed, i)

		png, err := qrcode.Encode(address.String(), qrcode.Low, 256)
		if err != nil {
			log.Fatal(err)
		}
		imageAddress := base64.StdEncoding.EncodeToString(png)
		addressPair := addressPair{
			Address:      address,
			AddressImage: imageAddress,
		}
		addressesPairs = append(addressesPairs, addressPair)
	}

	templateData := &secret{
		Seed:         seedStr,
		SeedImage:    seedImage,
		AddressPairs: addressesPairs,
	}
	return templateData
}

func main() {

	var port string
	port = os.Getenv("PORT")

	r := mux.NewRouter()

	r.HandleFunc("/", HandleHome)
	r.HandleFunc("/wallet", HandleWalletGenerator)
	r.PathPrefix("/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))

	if port == "" {
		port = "8080"
	}

	domain := ":" + port

	finalRouter := RedirectToHTTPSRouter(r)

	// Fasten to port and pass in routes
	log.Fatal(http.ListenAndServe(domain, finalRouter))
}
