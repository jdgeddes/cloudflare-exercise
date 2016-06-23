package main

import (
	"encoding/json"
	"flag"
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
	customerCol    *mgo.Collection
	certificateCol *mgo.Collection
)

//
// Utility Functions
//

func doesCustomerExist(email string) bool {
	// make sure the customer email exists in customer collection
	count, _ := customerCol.Find(bson.M{"email": email}).Count()
	if count == 0 {
		return false
	}

	return true
}

//
// Functions for creating, retrieving, and deleting customers
//

func createCustomer(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// get customer information from the POST
	customer := Customer{}
	json.NewDecoder(request.Body).Decode(&customer)

	// if customer email exists, return error
	if doesCustomerExist(customer.Email) {
		writer.WriteHeader(400)
		fmt.Fprintf(writer, "Customer with email exists.\n")
		return
	}

	// insert customer into database, on error return 500
	err := customerCol.Insert(customer)
	if err != nil {
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "Error inserting customer into database.\n")
		return
	}

	// respond with 201 and customer JSON information
	writer.Header().Set("Content-Type", "applications/json")
	writer.WriteHeader(201)
	response, _ := json.Marshal(&customer)
	fmt.Fprintf(writer, "%s\n", response)
}

func getCustomer(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// get email address from parmaters
	email := params.ByName("email")

	// lookup customer based on email, return 404 on not found
	customer := Customer{}
	err := customerCol.Find(bson.M{"email": email}).One(&customer)
	if err != nil {
		writer.WriteHeader(404)
		fmt.Fprintf(writer, "Customer does not exist\n")
		return
	}

	// respond with 200 and customer JSON information
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(200)
	response, _ := json.Marshal(&customer)
	fmt.Fprintf(writer, "%s\n", response)
}

func deleteCustomer(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// get the email address from parameters
	email := params.ByName("email")

	// remove the customer by email, return 404 on not found
	err := customerCol.Remove(bson.M{"email": email})
	if err != nil {
		writer.WriteHeader(404)
		fmt.Fprintf(writer, "Customer does not exist.\n")
		return
	}

	// respond with 200
	writer.WriteHeader(200)
}

//
// Functions for creating, updating, and print certficiates belonging to a customer
//

func createCertificate(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// get certificate information from POST
	certificate := Certificate{}
	json.NewDecoder(request.Body).Decode(&certificate)

	// make sure the customer email exists in customer collections
	if !doesCustomerExist(certificate.Email) {
		writer.WriteHeader(404)
		fmt.Fprintf(writer, "Customer does not exist\n")
		return
	}

	// create ID for certificate and insert it
	certificate.Id = bson.NewObjectId()
	certificateCol.Insert(certificate)

	// return 201 with JSON certificate
	writer.Header().Set("Content-Type", "applications/json")
	writer.WriteHeader(201)
	response, _ := json.Marshal(&certificate)
	fmt.Fprintf(writer, "%s\n", response)
}

func getCustomerCertificates(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// get email requested from parameters
	email := params.ByName("email")

	// if client does not exist, return 404
	if !doesCustomerExist(email) {
		writer.WriteHeader(404)
		fmt.Fprintf(writer, "Customer does not exist\n")
		return
	}

	// lookup all certificates.  note we check if the error is the email wasn't found.
	// possible for client to be created with no certificates
	var certificates []Certificate
	err := certificateCol.Find(bson.M{"email": email}).All(&certificates)
	if err != nil && err != mgo.ErrNotFound {
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "Error looking up customer certificates\n")
		return
	}

	// respond with 200 and JSON of certificates
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(200)
	response, _ := json.Marshal(&certificates)
	fmt.Fprintf(writer, "%s\n", response)
}

func updateCertificate(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	// get ID and check to make sure it's in the correct format
	id := params.ByName("id")
	if !bson.IsObjectIdHex(id) {
		writer.WriteHeader(404)
		fmt.Fprintf(writer, "Certificate not found.\n")
		return
	}
	objectId := bson.ObjectIdHex(id)

	// decode as certificate, but we're just interested in "active"
	activeVal := Certificate{}
	json.NewDecoder(request.Body).Decode(&activeVal)

	// update the certificate based on object ID, updating the active field
	query := bson.M{"id": objectId}
	change := bson.M{"$set": bson.M{"active": activeVal.Active}}
	err := certificateCol.Update(query, change)
	if err != nil {
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "Error updating certificate.\n")
		return
	}

	// return code 200
	writer.WriteHeader(200)
	fmt.Fprintf(writer, "Succesfully updated certificate\n")
}

func main() {
	// setup flags for database and server information
	dbhostPtr := flag.String("dbhost", "localhost", "Hostname of Mongo DB server")
	dbnamePtr := flag.String("dbname", "customerdb", "Database name to query and store customer information")
	portPtr := flag.Int("port", 8080, "Port for the server to run on")
	flag.Parse()

	// connect to mongo database
	session, err := mgo.Dial(*dbhostPtr)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	// get collections for storing customers and certificates
	customerCol = session.DB(*dbnamePtr).C("customers")
	certificateCol = session.DB(*dbnamePtr).C("certificates")

	// setup HTTP router for handling customer and certificate requests
	r := httprouter.New()

	r.POST("/customer", createCustomer)
	r.GET("/customer/:email", getCustomer)
	r.DELETE("/customer/:email", deleteCustomer)

	r.POST("/certificate", createCertificate)
	r.GET("/certificate/:email", getCustomerCertificates)
	r.PUT("/certificate/:id", updateCertificate)

	// start HTTP server and port specified
	address := fmt.Sprintf(":%d", *portPtr)
	http.ListenAndServe(address, r)
}
