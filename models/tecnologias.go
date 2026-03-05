package models

import (
	"nexhub/db" // Confirme se o caminho do seu import é esse mesmo
	"nexhub/structs"
)

func BuscarTodasTecnologias() ([]structs.Tecnologia, error) {
	// Certifique-se que o nome da coluna no banco é 'nome' ou 'nome_tech'
	query := `SELECT id_tech, nome_tech FROM tecnologias ORDER BY nome_tech ASC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []structs.Tecnologia
	for rows.Next() {
		var t structs.Tecnologia
		if err := rows.Scan(&t.Id, &t.Nome); err == nil {
			lista = append(lista, t)
		}
	}
	return lista, nil
}
