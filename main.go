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

// getAddress returns an address generated from a seed at the index specified
// by `index`.
func getAddress(seed modules.Seed, index uint64) types.UnlockHash {
	_, pk := crypto.GenerateKeyPairDeterministic(crypto.HashAll(seed, index))
	return types.UnlockConditions{
		PublicKeys:         []types.SiaPublicKey{types.Ed25519PublicKey(pk)},
		SignaturesRequired: 1,
	}.UnlockHash()
}

func main() {
	// generate a seed and a few addresses from that seed
	var seed modules.Seed
	fastrand.Read(seed[:])
	var addresses []types.UnlockHash
	seedStr, err := modules.SeedToString(seed, "english")
	if err != nil {
		log.Fatal(err)
	}
	for i := uint64(0); i < nAddresses; i++ {
		addresses = append(addresses, getAddress(seed, i))
	}

	templateData := struct {
		Seed      string
		Addresses []types.UnlockHash
	}{
		Seed:      seedStr,
		Addresses: addresses,
	}
	t, err := template.New("output").ParseFiles("./templates/secret.html")
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.Listen("tcp", "localhost:8087")
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Execute(w, templateData)
		l.Close()
		close(done)
	})
	go http.Serve(l, handler)

	// err = open.Run("http://localhost:8087")
	if err != nil {
		// fallback to console, clean up the server and exit
		l.Close()
		fmt.Println("Seed:", seedStr)
		fmt.Println("Addresses:")
		for _, address := range addresses {
			fmt.Println(address)
		}
		os.Exit(0)
	}
	<-done
}
