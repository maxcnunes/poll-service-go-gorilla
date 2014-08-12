package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
)

var (
	session    *mgo.Session
	collection *mgo.Collection
)

// Poll Model
type Poll struct {
	Id          bson.ObjectId `bson:"_id" json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Published   bool          `json:"published"`
}

func createHanler(w http.ResponseWriter, r *http.Request) {
	var poll Poll

	err := json.NewDecoder(r.Body).Decode(&poll)
	if err != nil {
		panic(err)
	}

	poll.Id = bson.NewObjectId()

	err = collection.Insert(&poll)
	if err != nil {
		panic(err)
	}

	log.Printf("Inserted new poll %s with title %s", poll.Id, poll.Title)

	j, err := json.Marshal(poll)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	var polls []Poll

	iter := collection.Find(nil).Iter()
	result := Poll{}
	for iter.Next(&result) {
		polls = append(polls, result)
	}

	w.Header().Set("Content-Type", "application/json")

	j, err := json.Marshal(polls)
	if err != nil {
		panic(err)
	}

	w.Write(j)
	log.Println("Provided json")
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])

	err := collection.Remove(bson.M{"_id": id})
	if err != nil {
		log.Printf("Could not find poll %s to delete", id)
	}

	w.WriteHeader(http.StatusNoContent)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])

	var poll Poll
	err := json.NewDecoder(r.Body).Decode(&poll)
	if err != nil {
		panic(err)
	}

	err = collection.Update(bson.M{"_id": id},
		bson.M{
			"title":       poll.Title,
			"description": poll.Description,
			"published":   poll.Published,
		})

	if err != nil {
		panic(err)
	}

	log.Printf("Updated poll %s title to %s", id, poll.Title)

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	var err error

	log.Println("Starting Server")

	router := mux.NewRouter()
	router.HandleFunc("/", listHandler).Methods("GET")
	router.HandleFunc("/", createHanler).Methods("POST")
	router.HandleFunc("/{id}", updateHandler).Methods("PUT")
	router.HandleFunc("/{id}", deleteHandler).Methods("DELETE")

	http.Handle("/", router)

	log.Println("Starting mongo db session")

	session, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	collection = session.DB("poll_db").C("poll")

	log.Println("Listening on 3000")
	http.ListenAndServe(":3000", nil)
}
