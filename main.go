package main

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var (
	addr = flag.String("addr", ":8080", "The address and port to bind to for the webserver")
)

type Basebox struct {
	Namespace string
	Name      string
	Version   string
	Provider  string
	File      string
}

type Provider struct {
	Name string
	Url  string
}

type Version struct {
	Version  string
	Status   string
	Provders []Provider
}

type Manifest struct {
	Description      string
	ShortDescription string
	Name             string
	Versions         []Version
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Hello World")
}

func ManifestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]

	result := Basebox{}
	collection := session.DB("coeus").C("boxes")

	iter := collection.Find(bson.M{"namespace": namespace, "name": name}).Iter()
	var versions []Version
	for iter.Next(&result) {
		var p []Provider
		p = append(p, Provider{result.Provider, namespace + "/boxes/" + name + "/" + result.Version + "/" + result.Provider + ".box"})
		versions = append(versions, Version{result.Version, "active", p})
	}
	msg, err := json.Marshal(Manifest{
		"Generated manifest for " + namespace + ":" + name,
		namespace + ":" + name,
		name,
		versions,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Add("Content-Type", "text/json")
		w.WriteHeader(http.StatusOK)
		w.Write(msg)
	}
}

func ProviderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]
	version := vars["version"]
	provider := vars["provider"]

	result := Basebox{}
	collection := session.DB("coeus").C("boxes")

	collection.Find(bson.M{"namespace": namespace, "name": name, "version": version, "provider": provider[0:strings.LastIndex(provider, ".")]}).One(&result)
	file, err := session.DB("coeus").GridFS("coeus").Open(result.File)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Add("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		io.Copy(w, file)
		file.Close()
	}
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	name := vars["name"]
	version := vars["version"]
	provider := vars["provider"]
	basename := namespace + ":" + name + "-" + version + ".box"

	if err := r.ParseMultipartForm(1000000000); err != nil { // Parse using 1Gb chunks
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(r.MultipartForm.File) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, boxes := range r.MultipartForm.File {
		if len(boxes) > 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for i, _ := range boxes {
			f, err := boxes[i].Open()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer f.Close()

			if file, err := session.DB("coeus").GridFS("coeus").Create(basename); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				defer file.Close()
				file.SetName(basename)
				if _, err := io.Copy(file, f); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}
	}
	collection := session.DB("coeus").C("boxes")
	collection.Insert(&Basebox{namespace, name, version, provider, basename})
	w.WriteHeader(http.StatusOK)
}

var session *mgo.Session

func init() {
	flag.Parse()
	var err error
	session, err = mgo.Dial("localhost:27017/coeus")
	if err != nil {
		panic(err)
	}

	collection := session.DB("coeus").C("boxes")

	collection.DropIndex("name", "version")
	collection.EnsureIndex(mgo.Index{
		Key:    []string{"namespace", "name", "version"},
		Unique: true,
	})

}

func main() {
	print("Running coeus\n")

	router := mux.NewRouter()
	router.HandleFunc("/", RootHandler)
	router.HandleFunc("/{namespace}/{name}", ManifestHandler)
	router.HandleFunc("/{namespace}/boxes/{name}/{version:[0-9\\.]*}/{provider:.*.box}", ProviderHandler).Methods("GET")
	router.HandleFunc("/{namespace}/{name}/{version:[0-9\\.]*}/{provider:.*.box}", UploadHandler).Methods("POST")

	//http.Handle("/", router)
	http.ListenAndServe(*addr, handlers.LoggingHandler(os.Stdout, router))

	session.Close()
}
