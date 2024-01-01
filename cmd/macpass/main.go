package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"internal/comunication"

	"github.com/musianisamuele/macpass/cmd/macpass/config"
	"github.com/musianisamuele/macpass/cmd/macpass/db"
	"github.com/musianisamuele/macpass/cmd/macpass/input"

	krbclient "github.com/jcmturner/gokrb5/v8/client"
	krbconfig "github.com/jcmturner/gokrb5/v8/config"
)

func main() {
	config.ParseConfig("./config.yaml")
	conf := config.Get()
	db.Connect(conf.DBPath)

	user := login()
	macAdd := input.Mac(user)
	time := input.RegistrationTime()

	fmt.Print(macAdd + "\t" + user + "\t")
	fmt.Println(time)

	db.InsertUser(user, macAdd)
	send(comunication.Request{User: user, Mac: macAdd, Duration: time})
}

func login() string {
	const krb5Conf = `[libdefaults]
  dns_lookup_realm = true
  dns_lookup_kdc = true
  `
	krbconf, err := krbconfig.NewFromString(krb5Conf)
	if err != nil {
		log.Fatal(err)
	}

	user, passwd := input.Credential()

	logins := config.Get().Login

	// test keberos logins
	for _, d := range logins.KerberosDomains {
		cl := krbclient.NewWithPassword(strings.Split(user, "@")[0], d.Realm,
			passwd, krbconf, krbclient.DisablePAFXFAST(d.DisablePAFXFAST))

		if err := cl.Login(); err == nil {
			// If login is succesful we return the user to bind the MAC address
			return user
		}
	}

	// test ldap logins
	// for _, d := range logins.LdapDomains {
	// }

	log.Fatal("The user can't be autenticated")
	return "" // unreachable code
}

func send(r comunication.Request) {
	// Connect to macpassd socket
	conn, err := net.Dial("unix", config.Get().SocketPath)
	if err != nil {
		log.Fatal(err)
	}

	jsonData, err := json.Marshal(r)
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Write(jsonData)
	if err != nil {
		log.Fatal(err)
	}

	conn.Close()
}
