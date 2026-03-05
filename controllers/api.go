package controllers

import (
	"encoding/json"
	"net/http"
	"nexhub/models"
)

// API: Retorna JSON com lista de usuários para o autocomplete
func ApiPesquisarDevs(w http.ResponseWriter, r *http.Request) {
	termo := r.URL.Query().Get("q")

	if len(termo) < 2 {
		json.NewEncoder(w).Encode([]string{}) // Retorna vazio se for muito curto
		return
	}

	usuarios, err := models.PesquisarDevsPorNome(termo)
	if err != nil {
		http.Error(w, "Erro na busca", 500)
		return
	}

	// Configura o header para JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usuarios)
}
