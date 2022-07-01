package handlers

import (
	"encoding/hex"
	"encoding/json"
	"eth2-exporter/metadata"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
)

var txTemplate = template.Must(template.New("tx").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/tx.html"))
var txNotFoundTemplate = template.Must(template.New("txnotfound").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/txnotfound.html"))

// Tx will show the tx using a go template
func Tx(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)
	txHashString := strings.Replace(vars["txHash"], "0x", "", -1)

	data := InitPageData(w, r, "txs", "/tx", "Transaction")
	data.HeaderAd = true

	txHash, err := hex.DecodeString(strings.ReplaceAll(txHashString, "0x", ""))

	if err != nil {
		data.Meta.Title = fmt.Sprintf("%v - Transaction %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, txHashString, time.Now().Year())
		data.Meta.Path = "/tx/" + txHashString
		logger.Errorf("error parsing tx hash %v: %v", txHashString, err)
		err = txNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	data.Meta.Title = fmt.Sprintf("%v - Tx 0x%x - beaconcha.in - %v", utils.Config.Frontend.SiteName, txHash, time.Now().Year())
	data.Meta.Path = fmt.Sprintf("/tx/0x%x", txHash)

	txData, err := metadata.GetEth1Transaction(common.BytesToHash(txHash))

	if err != nil {
		data.Meta.Title = fmt.Sprintf("%v - Transaction %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, txHashString, time.Now().Year())
		data.Meta.Path = "/tx/" + txHashString
		logger.Errorf(" %v", err)
		err = txNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	data.Data = txData

	if utils.IsApiRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(data.Data)
	} else {
		err = txTemplate.ExecuteTemplate(w, "layout", data)
	}

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
