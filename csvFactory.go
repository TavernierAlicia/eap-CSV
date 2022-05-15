package eapCSV

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

type OrderCSV struct {
	Id       int     `db:"id"`
	TotalTTC float64 `db:"totalTTC"`
	TotalHT  float64 `db:"totalHT"`
	Date     string  `db:"created"`
	Items    []*ItemCSV
}

type ItemCSV struct {
	Id       int     `db:"id"`
	Name     string  `db:"name"`
	Quantity int     `db:"quantity"`
	Price    float64 `db:"price"`
	Order_id int     `db:"order_id"`
}

func dbConnect() (db *sqlx.DB) {
	// connect database
	//// IMPORT CONFIG ////
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("reading config file failed: ", err)
	}

	//// DB CONNECTION ////
	pathSQL := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", viper.GetString("database.user"), viper.GetString("database.pass"), viper.GetString("database.host"), viper.GetInt("database.port"), viper.GetString("database.dbname"))
	db, err = sqlx.Connect("mysql", pathSQL)

	if err != nil {
		fmt.Println("Failed to connect to database: ", err)
	}

	return db
}

func dbGetCSVFacts(start string, end string, etabid int64) (result []*OrderCSV, err error) {
	db := dbConnect()

	err = db.Select(&result, "SELECT id, totalTTC, totalHT, created FROM orders WHERE etab_id = ? AND done = 1 AND created BETWEEN ? and ?", etabid, start, end)

	for i, order := range result {
		err = db.Select(&result[i].Items, "SELECT order_items.id, order_items.order_id, order_items.quantity, order_items.price, items.name FROM `order_items` JOIN items ON items.id = order_items.item_id WHERE order_items.order_id = ?", order.Id)
		fmt.Println("Error getting csv content, dbGetCSV: ", err)
	}

	fmt.Println(result)
	return result, err

}

// func FactstoCSV(content []*OrderCSV, etabid int64, start string, end string) (filepath string, err error) {

// 	var rows [][]string

// 	filepath = "media/csvs/" + strconv.FormatInt(etabid, 10) + "_" + strings.ReplaceAll(start, " ", "-") + "_to_" + strings.ReplaceAll(end, " ", "-") + "-export.csv"

// 	file, err := os.Create(filepath)

// 	if err != nil {
// 		fmt.Println("File creation failed, FactstoCSV: ", err)
// 	}

// 	writer := csv.NewWriter(file)

// 	for _, row := range content {

// 		fmt.Println(row.Id, row.Name, row.Quantity, row.Price, row.Order_id, row.Order_date)
// 		rows = append(rows, []string{strconv.Itoa(row.Id), row.Name, strconv.Itoa(row.Quantity), fmt.Sprintf("%.2f", row.Price), strconv.Itoa(row.Order_id), row.Order_date})

// 	}

// 	err = writer.WriteAll(rows)
// 	if err != nil {
// 		fmt.Println("Cannot write csv rows, FactstoCSV: ", err)
// 	}

// 	return filepath, err

// }
