package database

import (
	"AdaptixServer/core/utils/logs"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
)

func (dbms *DBMS) DbConsoleInsert(agentId string, packet interface{}) error {
	var buffer bytes.Buffer
	_ = json.NewEncoder(&buffer).Encode(packet)

	ok := dbms.DatabaseExists()
	if !ok {
		return errors.New("database not exists")
	}

	insertQuery := `INSERT INTO Consoles (AgentId, Packet) values(?,?);`
	_, err := dbms.database.Exec(insertQuery, agentId, buffer.Bytes())
	return err
}

func (dbms *DBMS) DbConsoleDelete(agentId string) error {
	ok := dbms.DatabaseExists()
	if !ok {
		return errors.New("database not exists")
	}

	deleteQuery := `DELETE FROM Consoles WHERE AgentId = ?;`
	_, err := dbms.database.Exec(deleteQuery, agentId)

	return err
}

func (dbms *DBMS) DbConsoleAll(agentId string) [][]byte {
	var consoles [][]byte

	ok := dbms.DatabaseExists()
	if ok {
		selectQuery := `SELECT Packet FROM Consoles WHERE AgentId = ? ORDER BY Id;`
		query, err := dbms.database.Query(selectQuery, agentId)
		if err == nil {

			for query.Next() {
				var message []byte
				err = query.Scan(&message)
				if err != nil {
					continue
				}
				consoles = append(consoles, message)
			}
		} else {
			logs.Debug("", err.Error()+" --- Clear database file!")
		}
		defer func(query *sql.Rows) {
			_ = query.Close()
		}(query)
	}
	return consoles
}
