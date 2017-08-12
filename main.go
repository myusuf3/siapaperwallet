package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/NebulousLabs/Sia/crypto"
	"github.com/NebulousLabs/Sia/modules"
	"github.com/NebulousLabs/Sia/types"
	"github.com/NebulousLabs/fastrand"
)

const nAddresses = 20

type Secret struct {
	Seed      string
	Addresses []types.UnlockHash
}

// getAddress returns an address generated from a seed at the index specified
// by `index`.
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
	seedStr, err := modules.SeedToString(seed, "english")
	if err != nil {
		log.Print(err)
		return nil, err
	}
	for i := uint64(0); i < nAddresses; i++ {
		addresses = append(addresses, getAddress(seed, i))
	}

	templateData := &Secret{
		Seed:      seedStr,
		Addresses: addresses,
	}
	return templateData, nil
}

func main() {

	port := os.Getenv("PORT")

	// generate a seed and a few addresses from that seed

	t := template.Must(template.ParseFiles("templates/secret.html"))
	domain := fmt.Sprintf(":%s", port)
	l, err := net.Listen("tcp", domain)
	if err != nil {
		log.Print(err)
	}

	done := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		templateData, err := GenerateNewSeedAddress()
		if err != nil {
			log.Fatal(err)
		}
		t.Execute(w, templateData)
		l.Close()
	})
	go http.Serve(l, handler)
	<-done
}
