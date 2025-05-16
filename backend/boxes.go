package freezerinv

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"gitlab.com/UrsusArcTech/logger"
)

type Box struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	FreezerId int    `json:"freezer_id"`
	Shelf     int    `json:"shelf"`
}

type BoxesInFreezers struct {
	Lab         string `json:"lab"`
	Floor       string `json:"floor"`
	FreezerName string `json:"freezer_name"`
	FreezerId   int    `json:"freezer_id"`
	BoxId       int    `json:"box_id"`
	Shelf       int    `json:"shelf"`
}

func MoveAllBoxesToShelf(w http.ResponseWriter, r *http.Request) {
	oldShelf := r.URL.Query().Get("oldshelf")
	newShelf := r.URL.Query().Get("newshelf")
	oldFreezer := r.URL.Query().Get("oldfreezer")
	newFreezer := r.URL.Query().Get("newfreezer")

	query := "UPDATE mgl_freezer_inventory.boxes SET shelf = $1, freezer_id = $2 WHERE shelf = $3 and freezer_id = $4"
	args := []interface{}{newShelf, newFreezer, oldShelf, oldFreezer}

	_, err := db.Exec(context.Background(), query, args...)
	if err != nil {
		logger.LogError("Database error: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Request was successful"))
}

func GetAllBoxes(w http.ResponseWriter, r *http.Request) {
	query := "select lab, floor, f.name as freezer_name, freezer_id, b.id as box_id, shelf from mgl_freezer_inventory.boxes b join mgl_freezer_inventory.freezer f on b.freezer_id = f.id join mgl_freezer_inventory.freezer_locations fl on fl.id = f.freezer_location_id"

	logger.LogMessage(query)
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var results []BoxesInFreezers

	for rows.Next() {
		var box BoxesInFreezers

		err := rows.Scan(
			&box.Lab,
			&box.Floor,
			&box.FreezerName,
			&box.FreezerId,
			&box.BoxId,
			&box.Shelf,
		)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		results = append(results, box)

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func GetBoxesByFreezer(w http.ResponseWriter, r *http.Request) {
	query := "select id, name, freezer_id, shelf from mgl_freezer_inventory.boxes where freezer_id = $1"
	args := []interface{}{}

	roomId := r.URL.Query().Get("freezerid")

	errRooms := errors.New("No freezer ID specified for box")
	if roomId == "" {
		logger.LogError("No freezer ID specified for box")
		http.Error(w, errRooms.Error(), 500)
		return
	}

	args = append(args, roomId)

	logger.LogMessage(query)
	rows, err := db.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var results []Box

	for rows.Next() {
		var box Box

		err := rows.Scan(
			&box.Id,
			&box.Name,
			&box.FreezerId,
			&box.Shelf,
		)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		results = append(results, box)

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)

}

// InsertBox handles HTTP POST requests to create a new box
func InsertBox(w http.ResponseWriter, r *http.Request) {

	freezerId := r.URL.Query().Get("freezerid")
	shelf := r.URL.Query().Get("shelf")
	name := r.URL.Query().Get("name")

	if freezerId == "" || name == "" || shelf == "" {
		logger.LogError("Missing required fields: name, shelf, and freezerid")
		http.Error(w, "Missing required fields: name, shelf, and freezerid", http.StatusBadRequest)
		return
	}

	query := "INSERT INTO mgl_freezer_inventory.boxes (name, freezer_id, shelf) VALUES ($1, $2, $3)"
	args := []interface{}{name, freezerId, shelf}

	_, err := db.Exec(context.Background(), query, args...)
	if err != nil {
		logger.LogError("Database error: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Request was successful"))
}

// InsertBox handles HTTP POST requests to create a new box
func DeleteBox(w http.ResponseWriter, r *http.Request) {

	boxid := r.URL.Query().Get("boxid")

	if boxid == "" {
		logger.LogError("Missing required fields: boxid")
		http.Error(w, "Missing required fields: boxid", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM mgl_freezer_inventory.boxes WHERE id = $1"
	args := []interface{}{boxid}

	_, err := db.Exec(context.Background(), query, args...)
	if err != nil {
		logger.LogError("Database error: " + err.Error())
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Request was successful"))
}

// UpdateBox handles HTTP PUT requests to update a box's FreezerID
func UpdateBox(w http.ResponseWriter, r *http.Request) {
	freezerId := r.URL.Query().Get("freezerid")
	name := r.URL.Query().Get("name")
	boxId := r.URL.Query().Get("boxid")
	shelf := r.URL.Query().Get("shelf")

	if freezerId == "" || name == "" || boxId == "" || shelf == "" {
		logger.LogError("Missing required fields: name, freezerid, shelf, and boxid")
		http.Error(w, "Missing required fields: name, freezerid, shelf, and boxid", http.StatusBadRequest)
		return
	}

	query := "UPDATE mgl_freezer_inventory.boxes SET freezer_id = $1, name = $2, shelf = $3 WHERE id = $4"
	args := []interface{}{freezerId, name, shelf, boxId}

	result, err := db.Exec(context.Background(), query, args...)
	if err != nil {
		logger.LogError("Database error: " + err.Error())
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		logger.LogError("No rows affected - box not found")
		http.Error(w, "Box not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
