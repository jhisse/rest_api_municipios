package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/gorilla/mux"
	"encoding/json"
	"database/sql"
)

type DB struct {
	*sql.DB
}

type municipio struct {
	CodigoMunicipio int               `json:"codigoMunicipio"`
	NomeMunicipio   string            `json:"nomeMunicipio"`
	UF              unidadeFederativa `json:"uf"`
}

type unidadeFederativa struct {
	CodigoUF int    `json:"codigoUF"`
	NomeUF   string `json:"nomeUF"`
	SiglaUF  string `json:"siglaUF"`
}

var (
	Municipios          []municipio
	UnidadesFederativas []unidadeFederativa
)

func abrirConexao() (*DB, error) {
	db, err := sql.Open("sqlite3", "./municipios.sqlite")
	return &DB{db}, err
}

func (db *DB) popularUnidadesFederativas() {
	rows, err := db.Query("select codigouf, siglauf, nomeuf from unidadesfederativas")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var codigo int
		var sigla string
		var nome string
		err = rows.Scan(&codigo, &sigla, &nome)
		if err != nil {
			log.Fatal(err)
		}
		UnidadesFederativas = append(UnidadesFederativas, unidadeFederativa{codigo, sigla, nome})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func (db *DB) popularMunicipios() {
	rows, err := db.Query("select codigomunicipio, nomemunicipio, unidadesfederacao_codigounidadefederacao from municipios")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var codigo int
		var nome string
		var ufFk int
		err = rows.Scan(&codigo, &nome, &ufFk)
		if err != nil {
			log.Fatal(err)
		}
		for _, uf := range UnidadesFederativas {
			if uf.CodigoUF == ufFk {
				Municipios = append(Municipios, municipio{codigo, nome, uf})
				break
			}
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	DB, err := abrirConexao()
	if err != nil {
		log.Panic(err)
	}
	DB.popularUnidadesFederativas()
	DB.popularMunicipios()
}

func getUFs(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(UnidadesFederativas)
}

func getUFsPorCodigo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for _, uf := range UnidadesFederativas {
		if strconv.Itoa(uf.CodigoUF) == params["codigo"] {
			json.NewEncoder(w).Encode(uf)
			return
		}
	}
	json.NewEncoder(w).Encode(nil)
}

func getMunicipios(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(&Municipios)
}

func getMunicipiosPorCodigo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for _, municipio := range Municipios {
		if strconv.Itoa(municipio.CodigoMunicipio) == params["codigo"] {
			json.NewEncoder(w).Encode(municipio)
			return
		}
	}
	json.NewEncoder(w).Encode(nil)
}

func addV1Routes(router *mux.Router) {
	router.HandleFunc("/unidadesFederativas", getUFs).Methods("GET")
	router.HandleFunc("/unidadesFederativas/{codigo:[0-9]{2}}", getUFsPorCodigo).Methods("GET")
	router.HandleFunc("/municipios", getMunicipios).Methods("GET")
	router.HandleFunc("/municipios/{codigo:[0-9]{7}}", getMunicipiosPorCodigo).Methods("GET")

	router.Headers("Content-Type", "application/json")
}

func main() {

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	v1 := api.PathPrefix("/v1").Subrouter()

	// Handle latest version
	addV1Routes(api)

	// Handle version 1
	addV1Routes(v1)

	fmt.Println("http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
