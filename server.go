package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type (
	Customer struct {
		Name  string `json:"name" bson:"name"`
		Email string `json:"email" bson:"email"`
	}

	Certificate struct {
		Id         bson.ObjectId `json:"id" bson:"id"`
		Email      string        `json:"email" bson:"email"`
		PrivateKey string        `json:"key" bson:"key"`
		Body       string        `json:"body" bson:"body"`
		Active     bool          `json:"active" bson:"active"`
	}
)

var (
	database     *mgo.Database
	databaseName = "customerdb"
)

//
// Functions for creating, retrieving, and deleting customers
//

func createCustomer(writer http.ResponseWriter, request *http.Request, p httprouter.Params) {
	customer := Customer{}

	json.NewDecoder(request.Body).Decode(&customer)

	database.C("customers").Insert(customer)

	response, _ := json.Marshal(&customer)

	writer.Header().Set("Content-Type", "applications/json")
	writer.WriteHeader(201)
	fmt.Fprintf(writer, "%s\n", response)
}

func getCustomer(writer http.ResponseWriter, request *http.Request, p httprouter.Params) {
	email := p.ByName("email")

	customer := Customer{}
	err := database.C("customers").Find(bson.M{"email": email}).One(&customer)
	if err != nil {
		writer.WriteHeader(404)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(200)
	fmt.Fprintf(writer, "%s\n", customer)
}

func deleteCustomer(writer http.ResponseWriter, request *http.Request, p httprouter.Params) {
	email := p.ByName("email")

	err := database.C("customers").Remove(bson.M{"email": email})
	if err != nil {
		writer.WriteHeader(404)
		return
	}

	writer.WriteHeader(200)
}

//
// Functions for creating, updating, and print certficiates belonging to a customer
//

func createCertificate(writer http.ResponseWriter, request *http.Request, p httprouter.Params) {
	certificate := Certificate{}

	json.NewDecoder(request.Body).Decode(&certificate)

	customer := Customer{}
	err := database.C("customers").Find(bson.M{"email": certificate.Email}).One(&customer)
	if err != nil {
		fmt.Fprintf(writer, "no customer for %s\n", certificate.Email)
		writer.WriteHeader(404)
		return
	}

	// create ID for certificate
	certificate.Id = bson.NewObjectId()

	database.C("certificates").Insert(certificate)

	response, _ := json.Marshal(&certificate)

	writer.Header().Set("Content-Type", "applications/json")
	writer.WriteHeader(201)
	fmt.Fprintf(writer, "%s\n", response)
}

func getCustomerCertificates(writer http.ResponseWriter, request *http.Request, p httprouter.Params) {
	email := p.ByName("email")

	var certificates []Certificate
	err := database.C("certificates").Find(bson.M{"email": email}).All(&certificates)
	if err != nil {
		writer.WriteHeader(404)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(200)

	for i := 0; i < len(certificates); i++ {
		response, _ := json.Marshal(&certificates[i])
		fmt.Fprintf(writer, "%s\n", response)
	}
}

func updateCertificate(writer http.ResponseWriter, request *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	if !bson.IsObjectIdHex(id) {
		writer.WriteHeader(404)
		return
	}
	objectId := bson.ObjectIdHex(id)

	activeVal := Certificate{}
	json.NewDecoder(request.Body).Decode(&activeVal)

	query := bson.M{"id": objectId}
	change := bson.M{"$set": bson.M{"active": activeVal.Active}}
	err := database.C("certificates").Update(query, change)
	if err != nil {
		writer.WriteHeader(404)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(200)
	fmt.Fprintf(writer, "Success\n")
}

func main() {
	fmt.Println("starting server")

	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	database = session.DB(databaseName)
	//database.DropDatabase()

	r := httprouter.New()
	r.POST("/customer", createCustomer)
	r.GET("/customer/:email", getCustomer)
	r.DELETE("/customer/:email", deleteCustomer)

	r.POST("/certificate", createCertificate)
	r.GET("/certificate/:email", getCustomerCertificates)
	r.PUT("/certificate/:id", updateCertificate)

	http.ListenAndServe(":8080", r)
}
