package controllers

import (
	"encoding/json"
	"net/http"
	"nexhub/models"
)

// API: Retorna JSON com lista de alunos para o autocomplete (Vínculo de equipe)
func ApiPesquisarAlunos(w http.ResponseWriter, r *http.Request) {
	termo := r.URL.Query().Get("q")

	if len(termo) < 2 {
		json.NewEncoder(w).Encode([]string{}) // Retorna vazio se for muito curto
		return
	}

	alunos, err := models.PesquisarAlunosPorNome(termo)
	if err != nil {
		http.Error(w, "Erro na busca", 500)
		return
	}

	// Configura o header para JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alunos)
}
