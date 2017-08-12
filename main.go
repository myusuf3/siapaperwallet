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

const nAddresses = 20

type Secret struct {
	Seed         string
	Addresses    []types.UnlockHash
	AddressImage []string
}

func getAddress(seed modules.Seed, index uint64) types.UnlockHash {
	_, pk := crypto.GenerateKeyPairDeterministic(crypto.HashAll(seed, index))
	return types.UnlockConditions{
		PublicKeys:         []types.SiaPublicKey{types.Ed25519PublicKey(pk)},
		SignaturesRequired: 1,
	}.UnlockHash()
}

func GenerateNewSeedAddress() (*Secret, error) {
	var seed modules.Seed
	fastrand.Read(seed[:])
	var addresses []types.UnlockHash
	var imageAddresses []string
	seedStr, err := modules.SeedToString(seed, "english")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	for i := uint64(0); i < nAddresses; i++ {
		address := getAddress(seed, i)
		addresses = append(addresses, address)
		var png []byte
		png, err := qrcode.Encode(address.String(), qrcode.Medium, 256)
		if err != nil {
			log.Fatal(err)
		}
		imageAddress := base64.StdEncoding.EncodeToString(png)
		imageAddresses = append(imageAddresses, imageAddress)
	}

	templateData := &Secret{
		Seed:         seedStr,
		Addresses:    addresses,
		AddressImage: imageAddresses,
	}
	return templateData, nil
}

func YourHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/secret.html"))
	templateData, err := GenerateNewSeedAddress()
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, templateData)
}

// getAddress returns an address generated from a seed at the index specified
// by `index`.

func main() {
	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/", YourHandler)
	var port string
	port = os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}
	domain := fmt.Sprintf(":%s", port)
	log.Print(domain)
	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(domain, r))
}
