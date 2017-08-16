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

type AddressPair struct {
	Address      types.UnlockHash
	AddressImage string
}

type Secret struct {
	Seed         string
	SeedImage    string
	AddressPairs []AddressPair
}

func HandleWalletGenerator(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/secret.html"))
	templateData, err := GenerateNewSeedAddress()
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, templateData)
}

func HandleHome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hi! Keep it moving.\n"))
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
	var addressesPairs []AddressPair
	var png []byte
	seedStr, err := modules.SeedToString(seed, "english")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	png, err = qrcode.Encode(seedStr, qrcode.Low, 256)
	if err != nil {
		log.Fatal(err)
	}
	seedImage := base64.StdEncoding.EncodeToString(png)

	for i := uint64(0); i < nAddresses; i++ {
		address := getAddress(seed, i)

		png, err := qrcode.Encode(address.String(), qrcode.Low, 256)
		if err != nil {
			log.Fatal(err)
		}
		imageAddress := base64.StdEncoding.EncodeToString(png)
		addressPair := AddressPair{
			Address:      address,
			AddressImage: imageAddress,
		}
		addressesPairs = append(addressesPairs, addressPair)
	}

	templateData := &Secret{
		Seed:         seedStr,
		SeedImage:    seedImage,
		AddressPairs: addressesPairs,
	}
	return templateData, nil
}

func LoaderHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("loaderio-6b41e5868c37b084abcf848f4f65cd3b"))
}

// getAddress returns an address generated from a seed at the index specified
// by `index`.

func main() {
	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/", HandleHome)
	r.HandleFunc("/wallet/", HandleWalletGenerator)
	var port string
	port = os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}
	domain := fmt.Sprintf(":%s", port)
	log.Print(domain)

	finalRouter := RedirectToHTTPSRouter(r)
	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(domain, finalRouter))
}
