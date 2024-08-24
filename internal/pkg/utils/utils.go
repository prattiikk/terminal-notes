// package utils

// import (
// 	"database/sql"
// 	"fmt"
// 	"log"

// 	_ "github.com/lib/pq"
// )

// // openDB opens and returns a database connection.
// func openDB() (*sql.DB, error) {
// 	connStr := "postgresql://article%20list_owner:UnHc9jlDV7Oo@ep-orange-bush-a19fqe45.ap-southeast-1.aws.neon.tech/article%20list?sslmode=require"
// 	db, err := sql.Open("postgres", connStr)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return db, nil
// }
// func AddItemToDB(item listItemViewModel) error {
// 	db, err := openDB()
// 	if err != nil {
// 		return err
// 	}
// 	defer db.Close()

// 	query := `INSERT INTO items (title, description, content) VALUES ($1, $2, $3)`
// 	_, err = db.Exec(query, item.title, item.desc, item.content)
// 	return err
// }

// // Example usage to check the database connection
// func CheckDBVersion() {
// 	db, err := openDB()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()

// 	rows, err := db.Query("SELECT version()")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()

//		var version string
//		for rows.Next() {
//			err := rows.Scan(&version)
//			if err != nil {
//				log.Fatal(err)
//			}
//		}
//		fmt.Printf("version=%s\n", version)
//	}
package utils
