package main

import (
	freezerinv "freezer_proto/backend"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"gitlab.com/UrsusArcTech/logger"
	"gitlab.com/mgl-database/mgl-go/flags"
)

func main() {
	// Allow CORS for all origins and methods (for quick prototyping)
	handler := http.FileServer(http.Dir("static"))
	corsHandler := func(w http.ResponseWriter, r *http.Request) {
		// Allow everything for prototype purposes
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			return
		}
		handler.ServeHTTP(w, r)
	}

	flags.CreateFlag("-geturl", "http://dfo-db:8282/", "Set API url and port.")

	_ = godotenv.Load() // Loads .env file into environment variables
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		logger.LogError("DB_URL not set")
	}

	freezerinv.Init(dsn)

	http.HandleFunc("/getfreezerrooms", freezerinv.GetFreezerRooms)
	http.HandleFunc("/getfreezersinrooms", freezerinv.GetFreezersInRoom)
	http.HandleFunc("/getboxesbyfreezer", freezerinv.GetBoxesByFreezer)
	http.HandleFunc("/insertbox", freezerinv.InsertBox)
	http.HandleFunc("/updatebox", freezerinv.UpdateBox)
	http.HandleFunc("/deletebox", freezerinv.DeleteBox)
	http.HandleFunc("/getallboxes", freezerinv.GetAllBoxes)
	http.HandleFunc("/moveallboxestoshelf", freezerinv.MoveAllBoxesToShelf)
	http.HandleFunc("/getallfreezers", freezerinv.GetAllFreezers)

	//eDNA
	http.HandleFunc("/ednalinkbybox", freezerinv.EdnaLinkByBox)
	http.HandleFunc("/insertednalink", freezerinv.InsertEdnaLink)
	http.HandleFunc("/updateednalink", freezerinv.UpdateEdnaLink)
	http.HandleFunc("/checkednaalreadyinbox", freezerinv.CheckEdnaAlreadyInABox)
	http.HandleFunc("/deleteednalink", freezerinv.DeleteEdnaLink)

	//fish
	http.HandleFunc("/fishlinkbybox", freezerinv.FishLinkByBox)
	http.HandleFunc("/insertfishlink", freezerinv.InsertfishLink)
	http.HandleFunc("/updatefishlink", freezerinv.UpdateFishLink)
	http.HandleFunc("/checkfishalreadyinbox", freezerinv.CheckFishAlreadyInABox)
	http.HandleFunc("/deletefishlink", freezerinv.DeleteFishLink)

	http.HandleFunc("/", corsHandler)
	log.Println("Serving static/ on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
