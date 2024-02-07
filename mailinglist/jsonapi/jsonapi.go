package jsonapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/ivykus/gocourse/mailinglist/mdb"
)

func setJsonHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func fromJson[T any](body io.Reader, target T) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	json.Unmarshal(buf.Bytes(), &target)
}

func returnJson[T any](w http.ResponseWriter, withData func() (T, error)) {
	setJsonHeader(w)

	data, serverErr := withData()
	if serverErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		serverErrJson, err := json.Marshal(&serverErr)
		if err != nil {
			log.Printf("returnJson: %v", err)
			return
		}
		w.Write(serverErrJson)
		return
	}

	dataJson, err := json.Marshal(&data)
	if err != nil {
		log.Printf("returnJson: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(dataJson)
}

func returnErr(w http.ResponseWriter, err error, code int) {
	returnJson(w, func() (any, error) {
		errorMessage := struct {
			Err string
		}{
			Err: err.Error(),
		}
		w.WriteHeader(code)
		return errorMessage, nil
	})
}

func CreateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(r.Body, &entry)

		err := mdb.CreateEmail(db, entry.Email)
		if err != nil {
			returnErr(w, err, http.StatusBadRequest)
			return
		}

		returnJson(w, func() (any, error) {
			log.Printf("JSON Created email %s", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})

}

func GetEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(r.Body, &entry)

		returnJson(w, func() (any, error) {
			log.Printf("JSON Get email %s", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})

}

func UpdateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(r.Body, &entry)

		err := mdb.UpdateEmail(db, &entry)
		if err != nil {
			returnErr(w, err, http.StatusBadRequest)
			return
		}

		returnJson(w, func() (any, error) {
			log.Printf("JSON Update email %s", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func DeleteEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(r.Body, &entry)

		err := mdb.DeleteEmail(db, entry.Email)
		if err != nil {
			returnErr(w, err, http.StatusBadRequest)
			return
		}

		returnJson(w, func() (any, error) {
			log.Printf("JSON Delete email %s", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func GetEmailBatch(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}
		queryOpts := mdb.GetEmailBatchQueryParams{}
		fromJson(r.Body, &queryOpts)
		if queryOpts.Page < 1 || queryOpts.Count < 1 {
			returnErr(w, errors.New(
				"Page and Count fields must be greater than 0"),
				http.StatusBadRequest)
		}

		returnJson(w, func() (any, error) {
			log.Printf("JSON Get email batch")
			return mdb.GetEmailBatch(db, queryOpts)
		})
	})
}

func Serve(db *sql.DB, bind string) {
	http.Handle("/email/create", CreateEmail(db))
	http.Handle("/email/get", GetEmail(db))
	http.Handle("/email/get_batch", GetEmailBatch(db))
	http.Handle("/email/update", UpdateEmail(db))
	http.Handle("/email/delete", DeleteEmail(db))
	log.Println("JSON server listening on ", bind)
	err := http.ListenAndServe(bind, nil)
	if err != nil {

		log.Fatal("JSON server error: ", err)
	}
}
