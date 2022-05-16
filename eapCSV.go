package eapCSV

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

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

func DbGetCSVFacts(start string, end string, etabid int64) (result []*OrderCSV, err error) {
	db := dbConnect()

	err = db.Select(&result, "SELECT id, totalTTC, totalHT, created FROM orders WHERE etab_id = ? AND done = 1 AND created BETWEEN ? and ? ORDER BY created ASC", etabid, start, end)

	for i, order := range result {
		err = db.Select(&result[i].Items, "SELECT order_items.id, order_items.order_id, order_items.quantity, order_items.price, items.name FROM `order_items` JOIN items ON items.id = order_items.item_id WHERE order_items.order_id = ?", order.Id)

		if err != nil {
			fmt.Println("Error getting csv content, dbGetCSV: ", err)
		}
	}

	return result, err

}

func FactstoCSV(content []*OrderCSV, etabid int64, start string, end string) (filepath string, err error) {

	var rows [][]string

	filepath = viper.GetString("links.cdn_csv") + strconv.FormatInt(etabid, 10) + "_" + strings.ReplaceAll(start, " ", "-") + "_to_" + strings.ReplaceAll(end, " ", "-") + "-export.csv"

	file, err := os.Create(filepath)

	if err != nil {
		fmt.Println("File creation failed, FactstoCSV: ", err)
	}

	writer := csv.NewWriter(file)

	rows = append(rows, []string{"Numéro", "Date", "Total HT", "Total TTC"})

	for i, command := range content {
		fmt.Println(command)
		rows = append(rows, []string{strconv.Itoa(command.Id), command.Date, fmt.Sprintf("%.2f", command.TotalHT), fmt.Sprintf("%.2f", command.TotalTTC)})

		rows = append(rows, []string{"", "ID", "Désignation", "Quantité", "Prix unitaire"})
		for _, item := range content[i].Items {
			rows = append(rows, []string{"", strconv.Itoa(item.Id), item.Name, strconv.Itoa(item.Quantity), fmt.Sprintf("%.2f", item.Price)})
		}
		rows = append(rows, []string{"", "", "", "", ""})
	}

	err = writer.WriteAll(rows)
	if err != nil {
		fmt.Println("Cannot write csv rows, FactstoCSV: ", err)
	}

	return filepath, err

}
